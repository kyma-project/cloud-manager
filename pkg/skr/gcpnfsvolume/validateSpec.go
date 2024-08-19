package gcpnfsvolume

import (
	"context"
	"fmt"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
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
		return validateCapacityForTier(ctx, st, capacity, capacityRange{2560, 65400, 1})
	case cloudresourcesv1beta1.HIGH_SCALE_SSD:
		return validateCapacityForTier(ctx, st, capacity, capacityRange{10240, 102400, 1})
	case cloudresourcesv1beta1.ZONAL:
		return validateCapacityForTier(ctx, st, capacity, capacityRange{1024, 9984, 256}, capacityRange{10240, 102400, 2560})
	case cloudresourcesv1beta1.ENTERPRISE:
		return validateCapacityForTier(ctx, st, capacity, capacityRange{1024, 10240, 256})
	case cloudresourcesv1beta1.REGIONAL:
		return validateCapacityForTier(ctx, st, capacity, capacityRange{1024, 9984, 256}, capacityRange{10240, 102400, 2560})
	case cloudresourcesv1beta1.BASIC_HDD, cloudresourcesv1beta1.STANDARD:
		return validateCapacityForTier(ctx, st, capacity, capacityRange{1024, 65400, 1})
	default:
		state.ObjAsGcpNfsVolume().Status.State = cloudresourcesv1beta1.GcpNfsVolumeError
		return composed.UpdateStatus(state.ObjAsGcpNfsVolume()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonTierInvalid,
				Message: "Tier is not valid",
			}).
			ErrorLogMessage("Error updating GcpNfsVolume status with invalid capacity").
			Run(ctx, state)
	}
}

type capacityRange struct {
	min        int
	max        int
	increments int
}

func validateCapacityForTier(ctx context.Context, st composed.State, capacity int, validRanges ...capacityRange) (error, context.Context) {
	state := st.(*State)
	valid := false
	var rangesString []string
	for _, validRange := range validRanges {
		if capacity >= validRange.min && capacity <= validRange.max && (capacity-validRange.min)%validRange.increments == 0 {
			valid = true
		}
		if validRange.increments == 1 {
			rangesString = append(rangesString, fmt.Sprintf("%d to %d", validRange.min, validRange.max))
		} else {
			rangesString = append(rangesString, fmt.Sprintf("%d to %d scaling in increments of %d", validRange.min, validRange.max, validRange.increments))
		}
	}
	if valid {
		// Capacity is the only mutable field. If it succeeds, we should remove the error condition if the reason was invalid capacity
		return composed.UpdateStatus(state.ObjAsGcpNfsVolume()).
			RemoveConditionIfReasonMatched(cloudresourcesv1beta1.ConditionTypeError, cloudresourcesv1beta1.ConditionReasonCapacityInvalid).
			ErrorLogMessage("Error removing conditionType Error").
			OnUpdateSuccess(func(ctx context.Context) (error, context.Context) {
				return nil, nil
			}).
			Run(ctx, state)

	} else {
		state.ObjAsGcpNfsVolume().Status.State = cloudresourcesv1beta1.GcpNfsVolumeError
		return composed.UpdateStatus(state.ObjAsGcpNfsVolume()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonCapacityInvalid,
				Message: fmt.Sprintf("CapacityGb for this tier must be between %s", strings.Join(rangesString, " or ")),
			}).
			ErrorLogMessage("Error updating GcpNfsVolume status with invalid capacity").
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
		return composed.UpdateStatus(state.ObjAsGcpNfsVolume()).
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
		return composed.UpdateStatus(state.ObjAsGcpNfsVolume()).
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
			state.ObjAsGcpNfsVolume().Status.State = cloudresourcesv1beta1.GcpNfsVolumeError
			return composed.UpdateStatus(state.ObjAsGcpNfsVolume()).
				SetExclusiveConditions(metav1.Condition{
					Type:    cloudresourcesv1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  cloudresourcesv1beta1.ConditionReasonFileShareNameInvalid,
					Message: "FileShareName for this tier must be 16 characters or less",
				}).
				ErrorLogMessage("Error updating GcpNfsVolume status with invalid fileShareName").
				Run(ctx, state)
		} else {
			// if validation succeeds, we don't need to update the status
			return nil, nil
		}
	default:
		if len(fileShareName) > 64 {
			state.ObjAsGcpNfsVolume().Status.State = cloudresourcesv1beta1.GcpNfsVolumeError
			return composed.UpdateStatus(state.ObjAsGcpNfsVolume()).
				SetExclusiveConditions(metav1.Condition{
					Type:    cloudresourcesv1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  cloudresourcesv1beta1.ConditionReasonFileShareNameInvalid,
					Message: "FileShareName for this tier must be 64 characters or less",
				}).
				ErrorLogMessage("Error updating GcpNfsVolume status with invalid fileShareName").
				Run(ctx, state)
		} else {
			// if validation succeeds, we don't need to update the status
			return nil, nil
		}
	}
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
		return composed.UpdateStatus(state.ObjAsGcpNfsVolume()).
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

// This is to prevent legacy tiers when creating new volumes
func validateTier(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	var message string
	switch state.ObjAsGcpNfsVolume().Spec.Tier {
	case cloudresourcesv1beta1.ENTERPRISE:
		message = "ENTERPRISE is a legacy tier. Use REGIONAL tier instead."
	case cloudresourcesv1beta1.HIGH_SCALE_SSD:
		message = "HIGH_SCALE_SSD is a legacy tier. Use ZONAL tier instead."
	case cloudresourcesv1beta1.STANDARD:
		message = "STANDARD is a legacy tier. Use BASIC_HDD tier instead."
	case cloudresourcesv1beta1.PREMIUM:
		message = "PREMIUM is a legacy tier. Use BASIC_SSD tier instead."
	case cloudresourcesv1beta1.BASIC_SSD, cloudresourcesv1beta1.BASIC_HDD, cloudresourcesv1beta1.ZONAL, cloudresourcesv1beta1.REGIONAL:
		return nil, nil
	default:
		state.ObjAsGcpNfsVolume().Status.State = cloudresourcesv1beta1.GcpNfsVolumeError
		return composed.UpdateStatus(state.ObjAsGcpNfsVolume()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonTierInvalid,
				Message: "Tier is not valid",
			}).
			ErrorLogMessage("Error updating GcpNfsVolume status with invalid tier error").
			Run(ctx, state)
	}
	state.ObjAsGcpNfsVolume().Status.State = cloudresourcesv1beta1.GcpNfsVolumeError
	return composed.UpdateStatus(state.ObjAsGcpNfsVolume()).
		SetExclusiveConditions(metav1.Condition{
			Type:    cloudresourcesv1beta1.ConditionTypeError,
			Status:  metav1.ConditionTrue,
			Reason:  cloudresourcesv1beta1.ConditionReasonTierLegacy,
			Message: message,
		}).
		ErrorLogMessage("Error updating GcpNfsVolume status with legacy tier error").
		Run(ctx, state)
}
