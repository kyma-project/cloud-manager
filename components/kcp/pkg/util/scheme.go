package util

import (
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"reflect"
	"strings"
)

func GetObjGvk(scheme *runtime.Scheme, out runtime.Object) (gvk schema.GroupVersionKind, e error) {
	gvkArr, _, err := scheme.ObjectKinds(out)
	if err != nil {
		return gvk, err
	}
	if len(gvkArr) == 0 {
		return gvk, runtime.NewNotRegisteredErrForType(scheme.Name(), reflect.TypeOf(out))
	}
	//if len(gvkArr) > 1 {
	//	logger.Info("type %T has more then one Scheme registration", out)
	//}
	gvk = gvkArr[0]
	return
}

func GetObjGvkFromListGkv(scheme *runtime.Scheme, listGvk schema.GroupVersionKind) (gvk schema.GroupVersionKind, e error) {
	if !strings.HasSuffix(listGvk.Kind, "List") {
		return gvk, fmt.Errorf("the GVK %s is not a list", listGvk)
	}
	kind := strings.TrimSuffix(listGvk.Kind, "List")
	for item := range scheme.AllKnownTypes() {
		if item.GroupVersion() == listGvk.GroupVersion() && item.Kind == kind {
			return item, nil
		}
	}
	targetGvk := listGvk
	targetGvk.Kind = kind
	return gvk, fmt.Errorf("the GVK %s unlisted from %s not found in scheme", targetGvk, listGvk)
}
