package actions

import (
	"context"
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
)

func WaitStatusReadyAndState(statusState string) composed.Action {
	return func(ctx context.Context, state composed.State) (error, context.Context) {
		logger := composed.LoggerFromCtx(ctx)

		obj, ok := state.Obj().(composed.ObjWithConditionsAndState)

		if !ok {
			return composed.LogErrorAndReturn(common.ErrLogical,
				fmt.Sprintf("%T is not of type composed.ObjWithConditionsAndState", state.Obj()),
				composed.StopAndForget,
				ctx)
		}

		if meta.IsStatusConditionTrue(*obj.Conditions(), cloudcontrolv1beta1.ConditionTypeReady) &&
			obj.State() == statusState {
			return nil, ctx
		}

		logger.Info(fmt.Sprintf("Waiting status condition Ready and status state %s", statusState))

		return composed.StopWithRequeueDelay(util.Timing.T1000ms()), ctx
	}
}
