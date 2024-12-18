package gcpnfsvolume

import (
	"context"
	"time"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func validateIpRange(ctx context.Context, st composed.State) (error, context.Context) {
	if composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, nil
	}
	logger := composed.LoggerFromCtx(ctx)

	state := st.(*State)
	ipRangeName := state.ObjAsGcpNfsVolume().Spec.IpRange
	if ipRangeName.Name == "" {
		return nil, nil
	}
	ipRange := &cloudresourcesv1beta1.IpRange{}
	err := st.Cluster().K8sClient().Get(ctx,
		ipRangeName.ObjKey(),
		ipRange)
	if client.IgnoreNotFound(err) != nil {
		return composed.LogErrorAndReturn(err, "Error loading referred IpRange", composed.StopWithRequeue, ctx)
	}
	if err != nil {
		logger.
			WithValues("ipRange", ipRangeName).
			Error(err, "Referred IpRange does not exist")
		state.ObjAsGcpNfsVolume().Status.State = cloudresourcesv1beta1.GcpNfsVolumeError
		return composed.PatchStatus(state.ObjAsGcpNfsVolume()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonIpRangeNotFound,
				Message: "IpRange for this GcpNfsVolume does not exist",
			}).
			ErrorLogMessage("Error updating GcpNfsVolume status with invalid ipRange").
			Run(ctx, state)
	}

	if !meta.IsStatusConditionTrue(ipRange.Status.Conditions, cloudresourcesv1beta1.ConditionTypeReady) {
		logger.
			WithValues("ipRange", ipRangeName).
			Error(err, "Referred IpRange is not Ready")
		state.ObjAsGcpNfsVolume().Status.State = cloudresourcesv1beta1.GcpNfsVolumeError
		return composed.PatchStatus(state.ObjAsGcpNfsVolume()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonIpRangeNotReady,
				Message: "IpRange for this GcpNfsVolume is not in Ready condition",
			}).
			OnUpdateSuccess(func(ctx context.Context) (error, context.Context) {
				return composed.StopWithRequeueDelay(3 * time.Second), nil
			}).
			ErrorLogMessage("Error updating GcpNfsVolume status with invalid ipRange").
			Run(ctx, state)
	}
	// If validation succeeds, we should remove the condition if the reason was invalid or not found ipRange
	return composed.PatchStatus(state.ObjAsGcpNfsVolume()).
		RemoveConditionIfReasonMatched(cloudresourcesv1beta1.ConditionTypeError, cloudresourcesv1beta1.ConditionReasonIpRangeNotReady).
		RemoveConditionIfReasonMatched(cloudresourcesv1beta1.ConditionTypeError, cloudresourcesv1beta1.ConditionReasonIpRangeNotFound).
		ErrorLogMessage("Error removing conditionType Error").
		OnUpdateSuccess(func(ctx context.Context) (error, context.Context) {
			return nil, nil
		}).
		Run(ctx, state)
}

func validateLocation(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	if composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, nil
	}
	location := state.ObjAsGcpNfsVolume().Spec.Location
	if location == "" {
		if feature.GcpNfsVolumeAutomaticLocationAllocation.Value(ctx) {
			return nil, nil
		}
		state.ObjAsGcpNfsVolume().Status.State = cloudresourcesv1beta1.GcpNfsVolumeError
		return composed.PatchStatus(state.ObjAsGcpNfsVolume()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonLocationInvalid,
				Message: "Location is required",
			}).
			ErrorLogMessage("Error updating GcpNfsVolume status with invalid location").
			Run(ctx, state)
	} else {
		// if validation succeeds, we don't need to update the status
		return nil, nil
	}
}
