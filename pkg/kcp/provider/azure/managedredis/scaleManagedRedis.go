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

// scaleManagedRedis issues a PATCH SKU update when the cluster exists but its
// current SKU differs from the desired spec SKU (within-family resize).
func scaleManagedRedis(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	obj := state.ObjAsAzureManagedRedis()

	// Only act when the cluster already exists.
	if state.managedRedis == nil {
		return nil, ctx
	}

	// Compare current Azure SKU with desired spec SKU.
	if state.managedRedis.SKU == nil || state.managedRedis.SKU.Name == nil {
		return nil, ctx
	}
	currentSKU := string(*state.managedRedis.SKU.Name)
	desiredSKU := obj.Spec.SKU
	if currentSKU == desiredSKU {
		return nil, ctx
	}

	skuName := armredisenterprise.SKUName(desiredSKU)
	update := armredisenterprise.ClusterUpdate{
		SKU: &armredisenterprise.SKU{
			Name: &skuName,
		},
	}

	composed.LoggerFromCtx(ctx).
		WithValues(
			"clusterName", obj.Name,
			"currentSKU", currentSKU,
			"desiredSKU", desiredSKU,
		).
		Info("Submitting Azure Managed Redis cluster scale (SKU update)")

	err := state.client.UpdateCluster(ctx, state.resourceGroupName, obj.Name, update)
	if err != nil {
		var respErr *azcore.ResponseError
		if errors.As(err, &respErr) {
			composed.LoggerFromCtx(ctx).
				WithValues(
					"errorCode", respErr.ErrorCode,
					"statusCode", respErr.StatusCode,
				).
				Error(err, "Azure rejected Managed Redis cluster scale (ResponseError)")
		} else {
			composed.LoggerFromCtx(ctx).Error(err, "Error scaling Azure Managed Redis cluster")
		}
		obj.Status.State = string(cloudcontrolv1beta1.StateError)
		return composed.UpdateStatus(obj).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonCloudProviderError,
				Message: fmt.Sprintf("Failed to scale Azure Managed Redis: %s", err),
			}).
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T60000ms())).
			Run(ctx, st)
	}

	obj.Status.State = string(cloudcontrolv1beta1.StateProcessing)
	composed.LoggerFromCtx(ctx).Info("Azure Managed Redis scale submitted, requeuing in 60s")
	return composed.UpdateStatus(obj).
		SuccessError(composed.StopWithRequeueDelay(util.Timing.T60000ms())).
		Run(ctx, st)
}
