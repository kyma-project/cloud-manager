package commonAction

import (
	"context"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	commonrate "github.com/kyma-project/cloud-manager/pkg/common/rate"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func genericDependencyLoad(ctx context.Context, dependencyObj client.Object, objReconciled composed.ObjWithStatus, c client.Client, namespace, name, kind string) (error, context.Context) {
	err := c.Get(ctx, types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, dependencyObj)

	if client.IgnoreNotFound(err) != nil {
		return composed.LogErrorAndReturn(err, fmt.Sprintf("Error loading dependency  %s %s", kind, name), composed.StopWithRequeue, ctx)
	}
	if err != nil {
		return composed.NewStatusPatcherComposed(objReconciled).
			MutateStatus(func(obj composed.ObjWithStatus) {
				meta.SetStatusCondition(obj.Conditions(), metav1.Condition{
					Type:               cloudcontrolv1beta1.ConditionTypeReady,
					Status:             metav1.ConditionFalse,
					ObservedGeneration: obj.GetGeneration(),
					Reason:             cloudcontrolv1beta1.ReasonInvalidDependency,
					Message:            fmt.Sprintf("%s %s is not found in namespace %q", kind, name, namespace),
				})
			}).
			OnSuccess(composed.RequeueAfter(commonrate.Slow10s.When(dependencyObj))).
			Run(ctx, c)
	}

	dependencyObjWithStatus, ok := dependencyObj.(composed.ObjWithStatus)
	if !ok {
		return nil, ctx
	}

	readyCond := meta.FindStatusCondition(ptr.Deref(dependencyObjWithStatus.Conditions(), nil), cloudcontrolv1beta1.ConditionTypeReady)

	if readyCond == nil || readyCond.Status == metav1.ConditionUnknown {
		// a transient status, either will get the ready condition soon, or it's being processed, requeue
		return composed.StopWithRequeueDelay(commonrate.Slow1s.When(objReconciled)), ctx
	}

	if readyCond.Status != metav1.ConditionTrue {
		return composed.NewStatusPatcherComposed(dependencyObjWithStatus).
			MutateStatus(func(obj composed.ObjWithStatus) {
				meta.SetStatusCondition(obj.Conditions(), metav1.Condition{
					Type:               cloudcontrolv1beta1.ConditionTypeReady,
					Status:             metav1.ConditionFalse,
					ObservedGeneration: obj.GetGeneration(),
					Reason:             cloudcontrolv1beta1.ReasonInvalidDependency,
					Message:            fmt.Sprintf("%s %s is not ready", kind, name),
				})
			}).
			OnSuccess(composed.RequeueAfter(commonrate.Slow10s.When(objReconciled))).
			Run(ctx, c)
	}

	return nil, ctx
}
