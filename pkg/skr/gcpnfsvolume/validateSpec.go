package gcpnfsvolume

import (
	"context"
	"fmt"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

func validateCapacity(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	if composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, nil
	}
	// Capacity hasn't changed. No need to validate
	if state.ObjAsGcpNfsVolume().Spec.CapacityGb == state.ObjAsGcpNfsVolume().Status.CapacityGb {
		return nil, nil
	}
	tier := state.ObjAsGcpNfsVolume().Spec.Tier
	capacity := state.ObjAsGcpNfsVolume().Spec.CapacityGb
	switch tier {
	case cloudresourcesv1beta1.BASIC_SSD, cloudresourcesv1beta1.PREMIUM:
		return validateCapacityForTier(ctx, st, 2560, 65400, capacity)
	case cloudresourcesv1beta1.HIGH_SCALE_SSD:
		return validateCapacityForTier(ctx, st, 10240, 102400, capacity)
	case cloudresourcesv1beta1.ZONAL:
		return validateCapacityForTier(ctx, st, 1024, 9980, capacity)
	case cloudresourcesv1beta1.ENTERPRISE, cloudresourcesv1beta1.REGIONAL:
		return validateCapacityForTier(ctx, st, 1024, 10240, capacity)
	case cloudresourcesv1beta1.BASIC_HDD, cloudresourcesv1beta1.STANDARD:
		return validateCapacityForTier(ctx, st, 1024, 65400, capacity)
	default:
		return composed.UpdateStatus(state.ObjAsGcpNfsVolume()).
			SetCondition(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonTierInvalid,
				Message: fmt.Sprintf("Tier is not valid"),
			}).
			RemoveConditions(cloudresourcesv1beta1.ConditionTypeReady).
			ErrorLogMessage("Error updating GcpNfsVolume status with invalid capacity").
			Run(ctx, state)
	}
}

func validateCapacityForTier(ctx context.Context, st composed.State, min, max, capacity int) (error, context.Context) {
	state := st.(*State)
	if capacity < min || capacity > max {
		return composed.UpdateStatus(state.ObjAsGcpNfsVolume()).
			SetCondition(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonCapacityInvalid,
				Message: fmt.Sprintf("CapacityGb for this tier must be between %d and %d", min, max),
			}).
			RemoveConditions(cloudresourcesv1beta1.ConditionTypeReady).
			ErrorLogMessage("Error updating GcpNfsVolume status with invalid capacity").
			Run(ctx, state)
	} else {
		// Capacity is the only mutable field. If it succeeds, we should remove the error condition if the reason was invalid capacity
		return composed.UpdateStatus(state.ObjAsGcpNfsVolume()).
			RemoveConditionIfReasonMatched(cloudresourcesv1beta1.ConditionTypeError, cloudresourcesv1beta1.ConditionReasonCapacityInvalid).
			ErrorLogMessage("Error removing conditionType Error").
			OnUpdateSuccess(func(ctx context.Context) (error, context.Context) {
				return nil, nil
			}).
			Run(ctx, state)
	}
}

func validateIpRange(ctx context.Context, st composed.State) (error, context.Context) {
	if composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, nil
	}
	logger := composed.LoggerFromCtx(ctx)

	state := st.(*State)
	ipRangeName := state.ObjAsGcpNfsVolume().Spec.IpRange
	ipRange := &cloudresourcesv1beta1.IpRange{}
	err := st.Cluster().K8sClient().Get(ctx,
		types.NamespacedName{
			Namespace: state.Obj().GetNamespace(),
			Name:      ipRangeName.Name},
		ipRange)
	if client.IgnoreNotFound(err) != nil {
		return composed.LogErrorAndReturn(err, "Error loading referred IpRange", composed.StopWithRequeue, ctx)
	}
	if err != nil {
		logger.
			WithValues("ipRange", ipRangeName).
			Error(err, "Referred IpRange does not exist")
		return composed.UpdateStatus(state.ObjAsGcpNfsVolume()).
			SetCondition(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonIpRangeNotFound,
				Message: fmt.Sprintf("IpRange for this GcpNfsVolume does not exist"),
			}).
			RemoveConditions(cloudresourcesv1beta1.ConditionTypeReady).
			ErrorLogMessage("Error updating GcpNfsVolume status with invalid ipRange").
			Run(ctx, state)
	}

	if !meta.IsStatusConditionTrue(ipRange.Status.Conditions, cloudresourcesv1beta1.ConditionTypeReady) {
		logger.
			WithValues("ipRange", ipRangeName).
			Error(err, "Referred IpRange is not Ready")
		return composed.UpdateStatus(state.ObjAsGcpNfsVolume()).
			SetCondition(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonIpRangeNotReady,
				Message: fmt.Sprintf("IpRange for this GcpNfsVolume is not in Ready condition"),
			}).
			OnUpdateSuccess(func(ctx context.Context) (error, context.Context) {
				return composed.StopWithRequeueDelay(3 * time.Second), nil
			}).
			RemoveConditions(cloudresourcesv1beta1.ConditionTypeReady).
			ErrorLogMessage("Error updating GcpNfsVolume status with invalid ipRange").
			Run(ctx, state)
	}
	// If validation succeeds, we should remove the condition if the reason was invalid or not found ipRange
	return composed.UpdateStatus(state.ObjAsGcpNfsVolume()).
		RemoveConditionIfReasonMatched(cloudresourcesv1beta1.ConditionTypeError, cloudresourcesv1beta1.ConditionReasonIpRangeNotReady).
		RemoveConditionIfReasonMatched(cloudresourcesv1beta1.ConditionTypeError, cloudresourcesv1beta1.ConditionReasonIpRangeNotFound).
		ErrorLogMessage("Error removing conditionType Error").
		OnUpdateSuccess(func(ctx context.Context) (error, context.Context) {
			return nil, nil
		}).
		Run(ctx, state)
}

func validateFileShareName(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	if composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, nil
	}
	tier := state.ObjAsGcpNfsVolume().Spec.Tier
	fileShareName := state.ObjAsGcpNfsVolume().Spec.FileShareName
	switch tier {
	case cloudresourcesv1beta1.BASIC_SSD, cloudresourcesv1beta1.PREMIUM, cloudresourcesv1beta1.BASIC_HDD, cloudresourcesv1beta1.STANDARD:
		if len(fileShareName) > 16 {
			return composed.UpdateStatus(state.ObjAsGcpNfsVolume()).
				SetCondition(metav1.Condition{
					Type:    cloudresourcesv1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  cloudresourcesv1beta1.ConditionReasonFileShareNameInvalid,
					Message: fmt.Sprintf("FileShareName for this tier must be 16 characters or less"),
				}).
				ErrorLogMessage("Error updating GcpNfsVolume status with invalid fileShareName").
				RemoveConditions(cloudresourcesv1beta1.ConditionTypeReady).
				Run(ctx, state)
		} else {
			// if validation succeeds, we don't need to update the status
			return nil, nil
		}
	default:
		if len(fileShareName) > 64 {
			return composed.UpdateStatus(state.ObjAsGcpNfsVolume()).
				SetCondition(metav1.Condition{
					Type:    cloudresourcesv1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  cloudresourcesv1beta1.ConditionReasonFileShareNameInvalid,
					Message: fmt.Sprintf("FileShareName for this tier must be 64 characters or less"),
				}).
				ErrorLogMessage("Error updating GcpNfsVolume status with invalid fileShareName").
				RemoveConditions(cloudresourcesv1beta1.ConditionTypeReady).
				Run(ctx, state)
		} else {
			// if validation succeeds, we don't need to update the status
			return nil, nil
		}
	}
}
