package crgcpnfsvolume

import (
	"context"
	"fmt"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func validateCapacity(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
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
		state.ObjAsGcpNfsVolume().Status.State = cloudresourcesv1beta1.ErrorState
		return composed.UpdateStatus(state.ObjAsGcpNfsVolume()).
			SetCondition(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeSpecValid,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonCapacityInvalid,
				Message: fmt.Sprintf("Tier is not valid"),
			}).
			ErrorLogMessage("Error updating GcpNfsVolume status with invalid capacity").
			Run(ctx, state)
	}
}

func validateCapacityForTier(ctx context.Context, st composed.State, min, max, capacity int) (error, context.Context) {
	state := st.(*State)
	if capacity < min || capacity > max {
		state.ObjAsGcpNfsVolume().Status.State = cloudresourcesv1beta1.ErrorState
		return composed.UpdateStatus(state.ObjAsGcpNfsVolume()).
			SetCondition(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeSpecValid,
				Status:  metav1.ConditionFalse,
				Reason:  cloudresourcesv1beta1.ConditionReasonCapacityInvalid,
				Message: fmt.Sprintf("CapacityGb for this tier must be between %d and %d", min, max),
			}).
			ErrorLogMessage("Error updating GcpNfsVolume status with invalid capacity").
			Run(ctx, state)
	} else {
		return composed.UpdateStatus(state.ObjAsGcpNfsVolume()).
			SetCondition(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeSpecValid,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonCapacityValid,
				Message: fmt.Sprintf("CapacityGb for this tier is valid"),
			}).
			ErrorLogMessage("Error updating GcpNfsVolume status with valid capacity").
			Run(ctx, state)
	}
}

func validateIpRange(ctx context.Context, st composed.State) (error, context.Context) {
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
		return composed.LogErrorAndReturn(err, "Error loading referred IpRange", composed.StopWithRequeue, nil)
	}
	if err != nil {
		logger.
			WithValues("ipRange", ipRangeName).
			Error(err, "Referred IpRange does not exist")
		state.ObjAsGcpNfsVolume().Status.State = cloudresourcesv1beta1.ErrorState
		return composed.UpdateStatus(state.ObjAsGcpNfsVolume()).
			SetCondition(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeSpecValid,
				Status:  metav1.ConditionFalse,
				Reason:  cloudresourcesv1beta1.ConditionReasonIpRangeInvalid,
				Message: fmt.Sprintf("IpRange for this GcpNfsVolume does not exist"),
			}).
			ErrorLogMessage("Error updating GcpNfsVolume status with invalid ipRange").
			Run(ctx, state)
	}
	condReady := meta.FindStatusCondition(ipRange.Status.Conditions, cloudresourcesv1beta1.ConditionTypeReady)
	if condReady == nil || condReady.Status != metav1.ConditionTrue {
		logger.
			WithValues("ipRange", ipRangeName).
			Error(err, "Referred IpRange is not Ready")
		state.ObjAsGcpNfsVolume().Status.State = cloudresourcesv1beta1.ErrorState
		return composed.UpdateStatus(state.ObjAsGcpNfsVolume()).
			SetCondition(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeSpecValid,
				Status:  metav1.ConditionFalse,
				Reason:  cloudresourcesv1beta1.ConditionReasonIpRangeInvalid,
				Message: fmt.Sprintf("IpRange for this GcpNfsVolume is not in Ready condition"),
			}).
			ErrorLogMessage("Error updating GcpNfsVolume status with invalid ipRange").
			Run(ctx, state)
	}
	return composed.UpdateStatus(state.ObjAsGcpNfsVolume()).
		SetCondition(metav1.Condition{
			Type:    cloudresourcesv1beta1.ConditionTypeSpecValid,
			Status:  metav1.ConditionTrue,
			Reason:  cloudresourcesv1beta1.ConditionReasonIpRangeValid,
			Message: fmt.Sprintf("IpRange is valid"),
		}).
		ErrorLogMessage("Error updating GcpNfsVolume status with valid ipRange").
		Run(ctx, state)
}

func validateFileShareName(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	tier := state.ObjAsGcpNfsVolume().Spec.Tier
	fileShareName := state.ObjAsGcpNfsVolume().Spec.FileShareName
	switch tier {
	case cloudresourcesv1beta1.BASIC_SSD, cloudresourcesv1beta1.PREMIUM, cloudresourcesv1beta1.BASIC_HDD, cloudresourcesv1beta1.STANDARD:
		if len(fileShareName) > 16 {
			state.ObjAsGcpNfsVolume().Status.State = cloudresourcesv1beta1.ErrorState
			return composed.UpdateStatus(state.ObjAsGcpNfsVolume()).
				SetCondition(metav1.Condition{
					Type:    cloudresourcesv1beta1.ConditionTypeSpecValid,
					Status:  metav1.ConditionFalse,
					Reason:  cloudresourcesv1beta1.ConditionReasonFileShareNameInvalid,
					Message: fmt.Sprintf("FileShareName for this tier must be 16 characters or less"),
				}).
				ErrorLogMessage("Error updating GcpNfsVolume status with invalid fileShareName").
				Run(ctx, state)
		} else {
			return composed.UpdateStatus(state.ObjAsGcpNfsVolume()).
				SetCondition(metav1.Condition{
					Type:    cloudresourcesv1beta1.ConditionTypeSpecValid,
					Status:  metav1.ConditionTrue,
					Reason:  cloudresourcesv1beta1.ConditionReasonFileShareNameValid,
					Message: fmt.Sprintf("FileShareName for this tier is valid"),
				}).
				ErrorLogMessage("Error updating GcpNfsVolume status with valid fileShareName").
				Run(ctx, state)
		}
	default:
		if len(fileShareName) > 64 {
			state.ObjAsGcpNfsVolume().Status.State = cloudresourcesv1beta1.ErrorState
			return composed.UpdateStatus(state.ObjAsGcpNfsVolume()).
				SetCondition(metav1.Condition{
					Type:    cloudresourcesv1beta1.ConditionTypeSpecValid,
					Status:  metav1.ConditionFalse,
					Reason:  cloudresourcesv1beta1.ConditionReasonFileShareNameInvalid,
					Message: fmt.Sprintf("FileShareName for this tier must be 64 characters or less"),
				}).
				ErrorLogMessage("Error updating GcpNfsVolume status with invalid fileShareName").
				Run(ctx, state)
		} else {
			return composed.UpdateStatus(state.ObjAsGcpNfsVolume()).
				SetCondition(metav1.Condition{
					Type:    cloudresourcesv1beta1.ConditionTypeSpecValid,
					Status:  metav1.ConditionTrue,
					Reason:  cloudresourcesv1beta1.ConditionReasonFileShareNameValid,
					Message: fmt.Sprintf("FileShareName for this tier is valid"),
				}).
				ErrorLogMessage("Error updating GcpNfsVolume status with valid fileShareName").
				Run(ctx, state)
		}
	}
}
