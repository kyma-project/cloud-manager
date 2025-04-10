package nfsinstance

import (
	"context"
	"fmt"
	efstypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"time"
)

func waitMountTargetsDeleted(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	for _, mt := range state.mountTargets {
		if mt.LifeCycleState == efstypes.LifeCycleStateDeleting {
			logger.Info("Waiting mount targets get deleted")
			return composed.StopWithRequeueDelay(300 * time.Millisecond), nil
		}
	}

	// if any mount target left in error state, set error condition and stopAndForget
	for _, mt := range state.mountTargets {
		if mt.LifeCycleState == efstypes.LifeCycleStateError {
			return composed.UpdateStatus(state.ObjAsNfsInstance()).
				SetExclusiveConditions(metav1.Condition{
					Type:    cloudcontrolv1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  cloudcontrolv1beta1.ReasonUnknown,
					Message: fmt.Sprintf("Mount target %s in error state", ptr.Deref(mt.MountTargetId, "")),
				}).
				SuccessError(composed.StopAndForget).
				Run(ctx, st)
		}
	}

	return nil, nil
}
