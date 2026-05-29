package managedredis

import (
	"context"
	"fmt"

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

	composed.LoggerFromCtx(ctx).Info("Creating Azure Managed Redis", "name", obj.Name)

	skuName := armredisenterprise.SKUName(obj.Spec.SKU)
	tlsVersion := armredisenterprise.TLSVersionOne2
	region := state.VpcNetwork().Spec.Region
	highAvailability := armredisenterprise.HighAvailabilityEnabled
	if !obj.Spec.HighAvailability {
		highAvailability = armredisenterprise.HighAvailabilityDisabled
	}
	publicNetworkAccess := armredisenterprise.PublicNetworkAccessDisabled
	cluster := armredisenterprise.Cluster{
		Location: &region,
		SKU: &armredisenterprise.SKU{
			Name: &skuName,
		},
		Properties: &armredisenterprise.ClusterCreateProperties{
			MinimumTLSVersion:   &tlsVersion,
			HighAvailability:    &highAvailability,
			PublicNetworkAccess: &publicNetworkAccess,
		},
	}

	// Note: Azure Managed Redis (Balanced/ComputeOptimized/MemoryOptimized/Flash SKUs)
	// does NOT support specifying zones — zone redundancy is enabled by default in
	// regions with availability zones. Do not set cluster.Zones for these SKUs.

	// Note: PublicNetworkAccess is set on create only. There is no updateManagedRedis
	// action, so clusters created before this field was introduced will not be patched.
	// This is intentional: AMR is a new resource type with no existing instances in the
	// field at the time this was added.

	err := state.client.CreateOrUpdateCluster(ctx, state.resourceGroupName, obj.Name, cluster)
	if err != nil {
		composed.LoggerFromCtx(ctx).Error(err, "Error creating Azure Managed Redis cluster")
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
	return composed.UpdateStatus(obj).
		SuccessError(composed.StopWithRequeueDelay(util.Timing.T60000ms())).
		Run(ctx, st)
}
