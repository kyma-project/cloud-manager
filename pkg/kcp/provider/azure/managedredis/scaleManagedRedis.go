package managedredis

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

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

	allowedSKUs, err := state.client.ListSKUsForScaling(ctx, state.resourceGroupName, obj.Name)
	if err != nil {
		var respErr *azcore.ResponseError
		if errors.As(err, &respErr) && respErr.StatusCode == 403 {
			// Service principal lacks listSkusForScaling/action permission — skip validation and attempt PATCH directly.
			composed.LoggerFromCtx(ctx).
				WithValues("currentSKU", currentSKU, "desiredSKU", desiredSKU).
				Info("WARNING: listSkusForScaling returned 403, SKU validation skipped - add Microsoft.Cache/redisEnterprise/listSkusForScaling/action permission to enable it")
		} else {
			composed.LoggerFromCtx(ctx).Error(err, "Failed to list allowed SKUs for scaling Azure Managed Redis")
			obj.Status.State = string(cloudcontrolv1beta1.StateError)
			return composed.UpdateStatus(obj).
				SetExclusiveConditions(metav1.Condition{
					Type:    cloudcontrolv1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  cloudcontrolv1beta1.ReasonCloudProviderError,
					Message: fmt.Sprintf("Failed to list allowed SKUs for scaling Azure Managed Redis: %s", err),
				}).
				SuccessError(composed.StopWithRequeueDelay(util.Timing.T60000ms())).
				Run(ctx, st)
		}
	}
	if len(allowedSKUs) > 0 && !slices.Contains(allowedSKUs, desiredSKU) {
		msg := fmt.Sprintf("SKU %s is not a valid scale target from %s (allowed: %s)", desiredSKU, currentSKU, strings.Join(allowedSKUs, ", "))
		composed.LoggerFromCtx(ctx).
			WithValues("currentSKU", currentSKU, "desiredSKU", desiredSKU, "allowedSKUs", allowedSKUs).
			Error(errors.New(msg), "Azure Managed Redis scale target not allowed")
		obj.Status.State = string(cloudcontrolv1beta1.StateError)
		return composed.UpdateStatus(obj).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonCloudProviderError,
				Message: msg,
			}).
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T60000ms())).
			Run(ctx, st)
	}

	composed.LoggerFromCtx(ctx).
		WithValues(
			"clusterName", obj.Name,
			"currentSKU", currentSKU,
			"desiredSKU", desiredSKU,
		).
		Info("Submitting Azure Managed Redis cluster scale (SKU update)")

	err = state.client.UpdateCluster(ctx, state.resourceGroupName, obj.Name, update)
	if err != nil {
		var respErr *azcore.ResponseError
		if errors.As(err, &respErr) {
			// Azure returns 400 "busy undergoing system maintenance" shortly after
			// provisioning completes — treat as transient, not a permanent error.
			if strings.Contains(respErr.Error(), "busy undergoing system maintenance") {
				return composed.LogErrorAndReturn(err, "Azure Managed Redis busy, retrying scale", composed.StopWithRequeueDelay(util.Timing.T60000ms()), ctx)
			}
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
