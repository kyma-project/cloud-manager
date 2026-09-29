package vpcpeering

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
)

func removeReadyCondition(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	obj := state.ObjAsVpcPeering()

	if !meta.IsStatusConditionTrue(*obj.Conditions(), cloudcontrolv1beta1.ConditionTypeReady) {
		return nil, ctx
	}

	logger.Info("Removing Ready condition")

	meta.RemoveStatusCondition(obj.Conditions(), cloudcontrolv1beta1.ConditionTypeReady)

	return composed.PatchStatus(obj).
		ErrorLogMessage("Error removing ready condition from AliCloud VpcPeering").
		FailedError(composed.StopWithRequeue).
		SuccessError(composed.StopWithRequeue).
		Run(ctx, state)
}
