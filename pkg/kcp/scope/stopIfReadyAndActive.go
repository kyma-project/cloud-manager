package scope

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
)

func stopIfReadyAndActive(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.ObjAsScope() == nil {
		logger.Info("Continuing since Scope is not created")
		return nil, nil
	}

	readyCond := meta.FindStatusCondition(*state.ObjAsScope().Conditions(), cloudcontrolv1beta1.ConditionTypeReady)
	if readyCond == nil {
		logger.Info("Continuing since Scope does not have Ready condition")
		return nil, nil
	}

	isSkrActive := state.activeSkrCollection.Contains(state.kyma.GetName())
	if !isSkrActive {
		logger.Info("Continuing since SKR is not activated")
		return nil, nil
	}

	logger.Info("Ignoring Scope with Ready condition and activated SKR")

	return composed.StopAndForget, nil
}
