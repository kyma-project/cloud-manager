package managedredis

import (
	"context"
	"errors"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redisenterprise/armredisenterprise/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func createManagedRedis(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	obj := state.ObjAsAzureManagedRedis()

	if state.managedRedis != nil {
		return nil, ctx
	}

	skuName := armredisenterprise.SKUName(obj.Spec.SKU)
	tlsVersion := armredisenterprise.TLSVersionOne2
	region := state.VpcNetwork().Spec.Region
	publicNetworkAccess := armredisenterprise.PublicNetworkAccessDisabled
	props := &armredisenterprise.ClusterCreateProperties{
		MinimumTLSVersion:   &tlsVersion,
		PublicNetworkAccess: &publicNetworkAccess,
	}
	// Only send HighAvailability when the user wants it Disabled (non-HA Balanced tiers).
	// On ComputeOptimized SKUs in zonal regions, Azure rejects an explicit
	// HighAvailability=Enabled with "Specifying zones for SKU '...' is not supported.
	// Zone redundancy is enabled by default for this SKU in regions with zones."
	// Leaving the field unset lets Azure pick the SKU/region default (Enabled+ZR
	// for ComputeOptimized in zonal regions).
	if !obj.Spec.HighAvailability {
		ha := armredisenterprise.HighAvailabilityDisabled
		props.HighAvailability = &ha
	}
	cluster := armredisenterprise.Cluster{
		Location: &region,
		SKU: &armredisenterprise.SKU{
			Name: &skuName,
		},
		Properties: props,
	}

	haStr := "<unset>"
	if props.HighAvailability != nil {
		haStr = string(*props.HighAvailability)
	}
	composed.LoggerFromCtx(ctx).
		WithValues(
			"clusterName", obj.Name,
			"sku", obj.Spec.SKU,
			"region", region,
			"highAvailability", haStr,
		).
		Info("Submitting Azure Managed Redis cluster create request")

	// Note: PublicNetworkAccess is set on create only. There is no updateManagedRedis
	// action, so clusters created before this field was introduced will not be patched.
	// This is intentional: AMR is a new resource type with no existing instances in the
	// field at the time this was added.

	err := state.client.CreateOrUpdateCluster(ctx, state.resourceGroupName, obj.Name, cluster)
	if err != nil {
		var respErr *azcore.ResponseError
		if errors.As(err, &respErr) {
			composed.LoggerFromCtx(ctx).
				WithValues(
					"errorCode", respErr.ErrorCode,
					"statusCode", respErr.StatusCode,
				).
				Error(err, "Azure rejected Managed Redis cluster create (ResponseError)")
		} else {
			composed.LoggerFromCtx(ctx).Error(err, "Error creating Azure Managed Redis cluster")
		}
		obj.Status.State = string(cloudcontrolv1beta1.StateError)
		return composed.UpdateStatus(obj).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonCloudProviderError,
				Message: fmt.Sprintf("Failed to create Azure Managed Redis: %s", err),
			}).
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T60000ms())).
			Run(ctx, st)
	}

	obj.Status.State = string(cloudcontrolv1beta1.StateProcessing)
	composed.LoggerFromCtx(ctx).Info("Azure Managed Redis create submitted, requeuing in 60s")
	return composed.UpdateStatus(obj).
		SuccessError(composed.StopWithRequeueDelay(util.Timing.T60000ms())).
		Run(ctx, st)
}
