package composed

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ConditionPredicate func(condition metav1.Condition) bool

func NewFilterAllowedConditionsPredicate(allowedTypes ...string) ConditionPredicate {
	return func(condition metav1.Condition) bool {
		for _, allowedType := range allowedTypes {
			if condition.Type == allowedType {
				return true
			}
		}
		return false
	}
}

func NewFilterProhibitedConditionsPredicate(prohibitedTypes ...string) ConditionPredicate {
	return func(condition metav1.Condition) bool {
		for _, allowedType := range prohibitedTypes {
			if condition.Type == allowedType {
				return false
			}
		}
		return true
	}
}

type ConditionFilterWrapper interface {
	ObjWithConditions
	Inner() ObjWithConditions
}

type conditionFilter struct {
	ObjWithConditions
	filter ConditionPredicate
}

func (f *conditionFilter) Inner() ObjWithConditions {
	return f.ObjWithConditions
}

func FilterAllowedConditions(inner ObjWithConditions, allowedConditionTypes ...string) ConditionFilterWrapper {
	return &conditionFilter{
		ObjWithConditions: inner,
		filter:            NewFilterAllowedConditionsPredicate(allowedConditionTypes...),
	}
}

func FilterProhibitedConditions(inner ObjWithConditions, prohibitedConditionTypes ...string) ConditionFilterWrapper {
	return &conditionFilter{
		ObjWithConditions: inner,
		filter:            NewFilterProhibitedConditionsPredicate(prohibitedConditionTypes...),
	}
}
