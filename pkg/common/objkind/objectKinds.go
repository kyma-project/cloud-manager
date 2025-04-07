package objkind

import (
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"strings"
)

type ObjectKindsInfo struct {
	ObjOK    bool
	ObjGK    schema.GroupKind
	CrdOK    bool
	CrdGK    schema.GroupKind
	BusolaOK bool
	BusolaGK schema.GroupKind
}

func (info *ObjectKindsInfo) RealObjGK() schema.GroupKind {
	if info.CrdOK {
		return info.CrdGK
	}
	if info.BusolaOK {
		return info.BusolaGK
	}
	return info.ObjGK
}

func (info *ObjectKindsInfo) AnyOK() bool {
	return info.ObjOK || info.CrdOK || info.BusolaOK
}

// ObjectKinds returns info about the provided object kinds. It accepts both types and unstructured objects.
// It can distinguish three different cases:
// * object is a CRD defining some obj kind
// * object is busola configmap defining some obj kind
// * object is itself some obj kind
func ObjectKinds(obj client.Object, scheme *runtime.Scheme) (result ObjectKindsInfo) {
	result.ObjGK = obj.GetObjectKind().GroupVersionKind().GroupKind()
	if result.ObjGK.Kind == "" {
		gvk, err := apiutil.GVKForObject(obj, scheme)
		if err != nil {
			return
		}
		result.ObjGK = gvk.GroupKind()
	}
	result.ObjOK = true

	kg := strings.ToLower(result.ObjGK.String())
	if kg == "customresourcedefinition.apiextensions.k8s.io" {
		if u, ok := obj.(*unstructured.Unstructured); ok {
			crdGroup, groupFound, groupErr := unstructured.NestedString(u.Object, "spec", "group")
			crdKind, kindFound, kindErr := unstructured.NestedString(u.Object, "spec", "names", "kind")
			if groupFound && kindFound && groupErr == nil && kindErr == nil {
				result.CrdGK.Group = crdGroup
				result.CrdGK.Kind = crdKind
				result.CrdOK = true
			}
		}
		if crd, ok := obj.(*apiextensions.CustomResourceDefinition); ok {
			crdGroup := crd.Spec.Group
			crdKind := crd.Spec.Names.Kind
			result.CrdGK.Group = crdGroup
			result.CrdGK.Kind = crdKind
			result.CrdOK = true
		}
		if crd, ok := obj.(*apiextensionsv1.CustomResourceDefinition); ok {
			crdGroup := crd.Spec.Group
			crdKind := crd.Spec.Names.Kind
			result.CrdGK.Group = crdGroup
			result.CrdGK.Kind = crdKind
			result.CrdOK = true
		}
		if crd, ok := obj.(*apiextensionsv1beta1.CustomResourceDefinition); ok {
			crdGroup := crd.Spec.Group
			crdKind := crd.Spec.Names.Kind
			result.CrdGK.Group = crdGroup
			result.CrdGK.Kind = crdKind
			result.CrdOK = true
		}
	}

	if kg == "configmap" &&
		obj.GetLabels() != nil && obj.GetLabels()["busola.io/extension"] != "" {

		var general string
		if cm, ok := obj.(*unstructured.Unstructured); ok {
			gen, found, err := unstructured.NestedString(cm.Object, "data", "general")
			if found && err == nil {
				general = gen
			}
		}
		if cm, ok := obj.(*corev1.ConfigMap); ok {
			gen, found := cm.Data["general"]
			if found {
				general = gen
			}
		}

		if len(general) > 0 {
			obj := map[string]interface{}{}
			if err := yaml.Unmarshal([]byte(general), &obj); err == nil {
				cmGroup, groupFound, groupErr := unstructured.NestedString(obj, "resource", "group")
				cmKind, kindFound, kindErr := unstructured.NestedString(obj, "resource", "kind")
				if groupFound && kindFound && groupErr == nil && kindErr == nil {
					result.BusolaGK.Group = cmGroup
					result.BusolaGK.Kind = cmKind
					result.BusolaOK = true
				}
			}
		}
	}

	return
}
