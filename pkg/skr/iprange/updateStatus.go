package iprange

import (
	"context"
	"fmt"
	"github.com/elliotchance/pie/v2"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func updateStatus(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	logger = logger.WithValues("kcpIpRangeConditions", pie.Map(state.KcpIpRange.Status.Conditions, func(c metav1.Condition) string {
		return fmt.Sprintf("%s:%s", c.Type, c.Status)
	}))

	kcpCondErr := meta.FindStatusCondition(state.KcpIpRange.Status.Conditions, cloudcontrolv1beta1.ConditionTypeError)
	kcpCondReady := meta.FindStatusCondition(state.KcpIpRange.Status.Conditions, cloudcontrolv1beta1.ConditionTypeReady)
	kcpMarkedForDeletion := composed.IsMarkedForDeletion(state.KcpIpRange)

	if kcpCondErr != nil {
		desiredCond := metav1.Condition{
			Type:    cloudresourcesv1beta1.ConditionTypeError,
			Status:  metav1.ConditionTrue,
			Reason:  cloudresourcesv1beta1.ConditionReasonError,
			Message: kcpCondErr.Message,
		}
		if !composed.HasCondition(desiredCond, state.ObjAsIpRange().Status.Conditions) {
			logger.Info("Updating IpRange status with Error condition")
			return composed.UpdateStatus(state.ObjAsIpRange()).
				SetExclusiveConditions(desiredCond).
				DeriveStateFromConditions(state.MapConditionToState()).
				ErrorLogMessage("Error updating IpRange status with not ready condition due to KCP error").
				Run(ctx, state)
		}
	}

	if kcpCondReady != nil && !kcpMarkedForDeletion {
		desiredCond := metav1.Condition{
			Type:    cloudresourcesv1beta1.ConditionTypeReady,
			Status:  metav1.ConditionTrue,
			Reason:  cloudresourcesv1beta1.ConditionTypeReady,
			Message: kcpCondReady.Message,
		}
		if !composed.HasCondition(desiredCond, state.ObjAsIpRange().Status.Conditions) {
			logger.Info("Updating IpRange status with Ready condition")
			return composed.UpdateStatus(state.ObjAsIpRange()).
				SetExclusiveConditions(desiredCond).
				DeriveStateFromConditions(state.MapConditionToState()).
				ErrorLogMessage("Error updating IpRange status with ready condition").
				Run(ctx, state)
		}
	}

	// keep looping until KCP IpRange gets some condition
	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
}
