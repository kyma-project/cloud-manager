package feature

import (
	"github.com/kyma-project/cloud-manager/pkg/common/objkind"
	"github.com/kyma-project/cloud-manager/pkg/feature/types"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func tryFeatureAwareOnGK(gk schema.GroupKind, scheme *runtime.Scheme) (types.FeatureName, bool) {
	versions := scheme.VersionsForGroupKind(schema.GroupKind{
		Group: gk.Group,
		Kind:  gk.Kind,
	})
	for _, version := range versions {
		x, err := scheme.New(schema.GroupVersionKind{
			Group:   version.Group,
			Version: version.Version,
			Kind:    gk.Kind,
		})
		if err == nil {
			if fa, ok := x.(types.FeatureAwareObject); ok {
				return fa.SpecificToFeature(), true
			}
		}
	}
	return "", false
}

// ObjectToFeature returns FeatureName specific to the object. It can return empty string
// which means object is not specific to one certain feature, but to many or all.
// It first checks if object implements types.FeatureAwareObject. If not then feature is determined
// by fixed, hard coded determinations by its kind and group. If object has TypeMeta it will
// be used to find GVK, otherwise provided scheme is looked up.
func ObjectToFeature(obj client.Object, scheme *runtime.Scheme) types.FeatureName {
	if fa, ok := obj.(types.FeatureAwareObject); ok {
		return fa.SpecificToFeature()
	}

	if u, ok := obj.(*unstructured.Unstructured); ok {
		x, err := scheme.New(u.GroupVersionKind())
		if err == nil {
			if ufa, ok := x.(types.FeatureAwareObject); ok {
				return ufa.SpecificToFeature()
			}
		}
	}

	kindInfo := objkind.ObjectKinds(obj, scheme)

	if kindInfo.CrdOK {
		if f, ok := tryFeatureAwareOnGK(kindInfo.CrdGK, scheme); ok {
			return f
		}
	}

	if kindInfo.BusolaOK {
		if f, ok := tryFeatureAwareOnGK(kindInfo.BusolaGK, scheme); ok {
			return f
		}
	}

	return types.FeatureUnknown
}
