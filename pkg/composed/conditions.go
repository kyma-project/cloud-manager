package composed

import (
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func HasCondition(desired metav1.Condition, existingConditions []metav1.Condition) bool {
	cond := meta.FindStatusCondition(existingConditions, desired.Type)
	if cond == nil {
		return false
	}
	if cond.Status != desired.Status {
		return false
	}
	if cond.Reason != desired.Reason {
		return false
	}
	if cond.Message != desired.Message {
		return false
	}
	return true
}
