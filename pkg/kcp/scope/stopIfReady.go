package scope

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
)

func stopIfReady(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.ObjAsScope() == nil {
		return nil, nil
	}

	readyCond := meta.FindStatusCondition(*state.ObjAsScope().Conditions(), cloudcontrolv1beta1.ConditionTypeReady)
	if readyCond == nil {
		return nil, nil
	}

	logger.Info("Ignoring Scope with Ready condition")

	return composed.StopAndForget, nil
}
