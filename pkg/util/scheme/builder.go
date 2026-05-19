package scheme

import (
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
