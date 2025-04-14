package composed

import (
	"github.com/elliotchance/pie/v2"
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

// StatusCopyConditionsAndState copies conditions from source to the destination. If destination
// has a condition that source doesn't have it will be removed. If both
// implement ObjWithConditionsAndState it will also copy status.state.
func StatusCopyConditionsAndState(source ObjWithConditions, destination ObjWithConditions) (changed bool, addedConditions []string, removedConditions []string, newState string) {
	changed = false
	for _, srcCond := range *source.Conditions() {
		added := meta.SetStatusCondition(destination.Conditions(), metav1.Condition{
			Type:    srcCond.Type,
			Status:  srcCond.Status,
			Reason:  srcCond.Reason,
			Message: srcCond.Message,
		})
		if added {
			changed = true
			addedConditions = append(addedConditions, srcCond.Type)
		}
	}
	for _, dstCond := range *destination.Conditions() {
		if !HasCondition(dstCond, *source.Conditions()) {
			changed = true
			meta.RemoveStatusCondition(destination.Conditions(), dstCond.Type)
			removedConditions = append(removedConditions, dstCond.Type)
		}
	}

	if srcWStatus, srcOk := source.(ObjWithConditionsAndState); srcOk {
		if dstWStatus, dstOk := destination.(ObjWithConditionsAndState); dstOk {
			if srcWStatus.State() != "" && srcWStatus.State() != dstWStatus.State() {
				changed = true
				dstWStatus.SetState(srcWStatus.State())
			}
			newState = srcWStatus.State()
		}
	}

	return
}

func AnyConditionChanged(obj ObjWithConditions, conditionsToSet ...metav1.Condition) bool {
	return pie.All(conditionsToSet, func(x metav1.Condition) bool {
		c := meta.FindStatusCondition(*obj.Conditions(), x.Type)
		return c == nil || c.Reason != x.Reason || c.Message != x.Message || c.Status != x.Status
	})
}

func SyncConditions(obj ObjWithConditions, conditionsToSet ...metav1.Condition) bool {
	conditionsToRemove := pie.Filter(*obj.Conditions(), func(x metav1.Condition) bool {
		return meta.FindStatusCondition(conditionsToSet, x.Type) == nil
	})

	changed := false
	for _, condition := range conditionsToRemove {
		if meta.RemoveStatusCondition(obj.Conditions(), condition.Type) {
			changed = true
		}
	}

	for _, condition := range conditionsToSet {
		if meta.SetStatusCondition(obj.Conditions(), condition) {
			changed = true
		}
	}
	return changed
}
