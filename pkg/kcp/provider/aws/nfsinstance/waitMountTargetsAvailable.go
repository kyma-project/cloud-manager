package nfsinstance

import (
	"context"
	"fmt"
	efstypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"time"
)

func waitMountTargetsAvailable(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	actMap := util.NewDelayActIgnoreBuilder[efstypes.LifeCycleState](util.Ignore).
		Ignore(efstypes.LifeCycleStateDeleted).
		Delay(efstypes.LifeCycleStateCreating, efstypes.LifeCycleStateUpdating, efstypes.LifeCycleStateDeleting).
		Error(efstypes.LifeCycleStateError).
		Act(efstypes.LifeCycleStateAvailable).
		Build()

	var mtStates []string
	for _, mt := range state.mountTargets {
		mtStates = append(mtStates, fmt.Sprintf("{%s/%s/%s}",
			ptr.Deref(mt.MountTargetId, ""),
			ptr.Deref(mt.AvailabilityZoneName, ""),
			mt.LifeCycleState,
		))
	}
	logger.WithValues("mountTargetStates", fmt.Sprintf("%v", mtStates))

	for _, mt := range state.mountTargets {
		lll := logger.WithValues(
			"mountTargetId", mt.MountTargetId,
			"mountTargetZone", mt.AvailabilityZoneName,
			"mountTargetState", mt.LifeCycleState,
		)
		outcome := actMap.Case(mt.LifeCycleState)

		switch outcome {

		case util.Ignore:
			lll.Info("Ignoring mount target")
			continue

		case util.Delay:
			lll.Info("Waiting for mount target to be available")
			return composed.StopWithRequeueDelay(300 * time.Millisecond), nil

		case util.Error:
			lll.Info("Mount target in error state")
			return composed.UpdateStatus(state.ObjAsNfsInstance()).
				SetExclusiveConditions(metav1.Condition{
					Type:    cloudcontrolv1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  cloudcontrolv1beta1.ReasonUnknown,
					Message: fmt.Sprintf("Mount target %s/%s in error state", ptr.Deref(mt.MountTargetId, ""), ptr.Deref(mt.AvailabilityZoneName, "")),
				}).
				SuccessError(composed.StopAndForget).
				Run(ctx, state)

		case util.Act:
			// it's available
			continue
		} // switch
	} // for

	return nil, nil
}
