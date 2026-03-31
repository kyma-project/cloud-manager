package clean

import (
	"reflect"
	"strings"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

type Matcher func(gvk schema.GroupVersionKind, tp reflect.Type) bool

func MatchAll(matchers ...Matcher) Matcher {
	return func(gvk schema.GroupVersionKind, tp reflect.Type) bool {
		for _, m := range matchers {
			if !m(gvk, tp) {
				return false
			}
		}
		return true
	}
}

func NotMatch(matcher Matcher) Matcher {
	return func(gvk schema.GroupVersionKind, tp reflect.Type) bool {
		result := !matcher(gvk, tp)
		return result
	}
}

func MatchingGroup(group string) Matcher {
	return func(gvk schema.GroupVersionKind, _ reflect.Type) bool {
		result := gvk.Group == group
		return result
	}
}

func MatchingKind(kind string) Matcher {
	listKind := kind + "List"
	return func(gvk schema.GroupVersionKind, _ reflect.Type) bool {
		return strings.EqualFold(gvk.Kind, kind) || strings.EqualFold(gvk.Kind, listKind)
	}
}
