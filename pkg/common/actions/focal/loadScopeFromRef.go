package focal

import (
	"context"
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func loadScopeFromRef(ctx context.Context, st composed.State) (error, context.Context) {
	logger := composed.LoggerFromCtx(ctx)
	state := st.(State)

	logger = logger.WithValues(
		"scope", state.ObjAsCommonObj().ScopeRef().Name,
		"scopeNamespace", state.ObjAsCommonObj().GetNamespace(),
	)
	ctx = composed.LoggerIntoCtx(ctx, logger)

	scope := &cloudcontrolv1beta1.Scope{}
	err := state.Cluster().K8sClient().Get(ctx, types.NamespacedName{
		Name:      state.ObjAsCommonObj().ScopeRef().Name,
		Namespace: state.ObjAsCommonObj().GetNamespace(),
	}, scope)

	if apierrors.IsNotFound(err) {
		logger.Info("Scope not found")

		if state.isScopeOptional() {
			return nil, ctx
		}

		return composed.UpdateStatus(state.ObjAsCommonObj()).
			SetCondition(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonScopeNotFound,
				Message: fmt.Sprintf("Scope %s does not exist", state.ObjAsCommonObj().ScopeRef().Name),
			}).
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T60000ms())).
			Run(ctx, state)
	}

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error loading Scope", composed.StopWithRequeue, ctx)
	}

	logger = logger.WithValues(
		"provider", scope.Spec.Provider,
		"region", scope.Spec.Region,
		"shootName", scope.Spec.ShootName,
	)
	if scope.Spec.Provider == cloudcontrolv1beta1.ProviderAws && scope.Spec.Scope.Aws != nil {
		logger = logger.WithValues(
			"awsAccount", scope.Spec.Scope.Aws.AccountId,
		)
	}
	ctx = composed.LoggerIntoCtx(ctx, logger)

	state.SetScope(scope)

	return nil, ctx
}
