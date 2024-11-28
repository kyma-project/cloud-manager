package dsl

import (
	"fmt"
	"strings"

	"github.com/elliotchance/pie/v2"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func SkrReadyCondition() metav1.Condition {
	return metav1.Condition{
		Type:    cloudresourcesv1beta1.ConditionTypeReady,
		Status:  metav1.ConditionTrue,
		Reason:  cloudresourcesv1beta1.ConditionTypeReady,
		Message: "Ready",
	}
}

func KcpReadyCondition() metav1.Condition {
	return metav1.Condition{
		Type:    cloudcontrolv1beta1.ConditionTypeReady,
		Status:  metav1.ConditionTrue,
		Reason:  cloudcontrolv1beta1.ConditionTypeReady,
		Message: "Ready",
	}
}

func KcpWarningCondition() metav1.Condition {
	return metav1.Condition{
		Type:    cloudcontrolv1beta1.ConditionTypeWarning,
		Status:  metav1.ConditionTrue,
		Reason:  cloudcontrolv1beta1.ConditionTypeWarning,
		Message: "Warning",
	}
}

func WithoutConditions(removeConds ...string) ObjStatusAction {
	return &objStatusAction{
		f: func(obj client.Object) {
			if x, ok := obj.(composed.ObjWithConditions); ok {
				for _, c := range removeConds {
					meta.RemoveStatusCondition(x.Conditions(), c)
				}
			}
		},
	}
}

func WithConditions(setConds ...metav1.Condition) ObjStatusAction {
	return &objStatusAction{
		f: func(obj client.Object) {
			if x, ok := obj.(composed.ObjWithConditions); ok {
				for _, c := range setConds {
					meta.SetStatusCondition(x.Conditions(), c)
				}
			}
		},
	}
}

func WithState(state string) ObjStatusAction {
	return &objStatusAction{
		f: func(obj client.Object) {
			if x, ok := obj.(composed.ObjWithConditionsAndState); ok {
				x.SetState(state)
			}
		},
	}
}

func HavingState(state string) ObjAssertion {
	return func(obj client.Object) error {
		if x, ok := obj.(composed.ObjWithConditionsAndState); ok {
			if x.State() != state {
				conditions := pie.Map(*x.Conditions(), func(c metav1.Condition) string {
					return fmt.Sprintf("%s:%s:%s", c.Type, c.Status, c.Reason)
				})
				return fmt.Errorf("expected state %s but got %s with conditions: %v", state, x.State(), conditions)
			}
			return nil
		}
		return fmt.Errorf("type %T does not implement ObjWithConditionsAndState", obj)
	}
}

func HavingConditionTrue(conditionType string) ObjAssertion {
	return func(obj client.Object) error {
		if x, ok := obj.(composed.ObjWithConditions); ok {
			if !meta.IsStatusConditionTrue(*x.Conditions(), conditionType) {
				return fmt.Errorf(
					"expected object %T %s/%s to have status condition %s true, but following conditions found: %v",
					obj,
					obj.GetNamespace(), obj.GetName(),
					conditionType,
					pie.Map(*x.Conditions(), func(c metav1.Condition) string {
						return fmt.Sprintf("%s:%s:%s", c.Type, c.Status, c.Reason)
					}),
				)
			}
		}
		return nil
	}
}

func NotHavingConditionTrue(conditionType string) ObjAssertion {
	return func(obj client.Object) error {
		if x, ok := obj.(composed.ObjWithConditions); ok {
			if meta.IsStatusConditionTrue(*x.Conditions(), conditionType) {
				return fmt.Errorf(
					"expected object %T %s/%s not to have status condition %s true, but following conditions found: %v",
					obj,
					obj.GetNamespace(), obj.GetName(),
					conditionType,
					pie.Map(*x.Conditions(), func(c metav1.Condition) string {
						return fmt.Sprintf("%s:%s:%s:%s", c.Type, c.Status, c.Reason, c.Message)
					}),
				)
			}
		}
		return nil
	}
}

func HavingCondition(conditionType string, status metav1.ConditionStatus, reason, msg string) ObjAssertion {
	return func(obj client.Object) error {
		if x, ok := obj.(composed.ObjWithConditions); ok {
			cond := meta.FindStatusCondition(*x.Conditions(), conditionType)
			if cond == nil ||
				cond.Status != status ||
				cond.Reason != reason ||
				!strings.Contains(cond.Message, msg) {
				return fmt.Errorf(
					"expected object %T %s/%s to have status condition %s %s %s with '%s', but following conditions found: %v",
					obj,
					obj.GetNamespace(),
					obj.GetName(),
					conditionType,
					status,
					reason,
					msg,
					pie.Map(*x.Conditions(), func(c metav1.Condition) string {
						return fmt.Sprintf("%s:%s:%s", c.Type, c.Status, c.Reason)
					}),
				)
			}
		}
		return nil
	}
}

func HaveFinalizer(finalizer string) ObjAssertion {
	return func(obj client.Object) error {
		if len(obj.GetFinalizers()) == 0 {
			return fmt.Errorf(
				"expected object %T %s/%s to have finalizer %s, but following finalizers found: %v",
				obj,
				obj.GetNamespace(),
				obj.GetName(),
				finalizer,
				obj.GetFinalizers(),
			)
		}
		return nil
	}
}
