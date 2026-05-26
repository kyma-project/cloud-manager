package managedredis

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func loadScope(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	obj := state.ObjAsAzureManagedRedis()

	logger = logger.WithValues(
		"scope", obj.Spec.Scope.Name,
		"scopeNamespace", obj.GetNamespace(),
	)
	ctx = composed.LoggerIntoCtx(ctx, logger)

	scope := &cloudcontrolv1beta1.Scope{}
	err := state.Cluster().K8sClient().Get(ctx, types.NamespacedName{
		Name:      obj.Spec.Scope.Name,
		Namespace: obj.GetNamespace(),
	}, scope)

	if apierrors.IsNotFound(err) {
		return composed.UpdateStatus(obj).
			SetCondition(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonScopeNotFound,
				Message: fmt.Sprintf("Scope %s does not exist", obj.Spec.Scope.Name),
			}).
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T60000ms())).
			Run(ctx, st)
	}

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error loading Scope", composed.StopWithRequeue, ctx)
	}

	state.SetScope(scope)
	return nil, ctx
}
