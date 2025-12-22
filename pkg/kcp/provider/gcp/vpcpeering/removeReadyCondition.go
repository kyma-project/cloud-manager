package vpcpeering

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
)

func removeReadyCondition(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	readyCondition := meta.FindStatusCondition(ptr.Deref(state.ObjAsVpcPeering().Conditions(), []metav1.Condition{}), cloudcontrolv1beta1.ConditionTypeReady)
	if readyCondition == nil {
		return nil, nil
	}

	logger.Info("Removing Ready condition")

	meta.RemoveStatusCondition(state.ObjAsVpcPeering().Conditions(), cloudcontrolv1beta1.ConditionTypeReady)
	err := state.UpdateObjStatus(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating KCP VpcPeering status after removing Ready condition", composed.StopWithRequeue, ctx)
	}

	return composed.StopWithRequeue, ctx
}
