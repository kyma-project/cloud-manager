package composed

import (
	"github.com/elliotchance/pie/v2"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func SetExclusiveConditions(conditions *[]metav1.Condition, conditionsToSet ...metav1.Condition) bool {
	changed := false

	conditionsToRemove := pie.Filter(*conditions, func(condition metav1.Condition) bool {
		return meta.FindStatusCondition(conditionsToSet, condition.Type) == nil
	})

	pie.Each(conditionsToRemove, func(condition metav1.Condition) {
		changed = changed || meta.RemoveStatusCondition(conditions, condition.Type)
	})

	pie.Each(conditionsToSet, func(condition metav1.Condition) {
		changed = changed || meta.SetStatusCondition(conditions, condition)
	})

	return changed
}
