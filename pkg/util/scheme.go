package util

import (
	"fmt"
	"reflect"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// GroupVersionSchemeBuilder binds types to a GroupVersion using only apimachinery,
// providing the same Register/AddToScheme API as controller-runtime's scheme.Builder
// without the controller-runtime dependency.
type GroupVersionSchemeBuilder struct {
	GroupVersion schema.GroupVersion
	runtime.SchemeBuilder
}

func (b *GroupVersionSchemeBuilder) Register(object ...runtime.Object) {
	b.SchemeBuilder.Register(func(s *runtime.Scheme) error {
		s.AddKnownTypes(b.GroupVersion, object...)
		metav1.AddToGroupVersion(s, b.GroupVersion)
		return nil
	})
}

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
