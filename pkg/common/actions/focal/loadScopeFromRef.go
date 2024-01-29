package focal

import (
	"context"
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func loadScopeFromRef(ctx context.Context, st composed.State) (error, context.Context) {
	logger := composed.LoggerFromCtx(ctx)
	state := st.(State)

	logger.Info("Loading Scope from reference")

	scope := &cloudcontrolv1beta1.Scope{}
	err := state.Cluster().K8sClient().Get(ctx, types.NamespacedName{
		Name:      state.ObjAsCommonObj().ScopeRef().Name,
		Namespace: state.ObjAsCommonObj().GetNamespace(),
	}, scope)
	if apierrors.IsNotFound(err) {
		return composed.UpdateStatus(state.ObjAsCommonObj()).
			SetCondition(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonScopeNotFound,
				Message: fmt.Sprintf("Scope %s does not exist", state.ObjAsCommonObj().ScopeRef().Name),
			}).
			Run(ctx, state)
	}

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error loading Scope", composed.StopWithRequeue, nil)
	}

	logger.Info("Loaded Scope from reference")

	state.SetScope(scope)

	return nil, nil
}
