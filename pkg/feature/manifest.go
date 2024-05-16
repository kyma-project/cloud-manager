package feature

import (
	"context"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type featureDeterminator func(mi *manifestInfo) (bool, FeatureName)

type manifestInfo struct {
	obj             client.Object
	name            string
	namespace       string
	kindGroup       string
	labels          map[string]string
	annotations     map[string]string
	crdKindGroup    string
	busolaKindGroup string
}

var featureDeterminators = []featureDeterminator{
	featureDeterminatorByKindGroup,
	featureDeterminatorByCrdKindGroup,
	featureDeterminatorByBusolaKindGroup,
}

var featuresByKindGroup = map[string]FeatureName{
	// Please respect the order of the filenames in config/crd/bases
	// and list items here in the same order as files in that dir are.

	// KCP ==============================================================
	// scope has no feature defined

	// iprange atm is for nfs feature only, but in the future other features might include it as well,
	// so we will define iprange feature from start as undefined

	"nfsinstance.cloud-control.kyma-project.io": FeatureNfs,
	"vpcpeering.cloud-control.kyma-project.io":  FeaturePeering,

	// SKR ==============================================================

	"awsnfsvolumebackup.cloud-resources.kyma-project.io": FeatureNfsBackup,
	"awsnfsvolume.cloud-resources.kyma-project.io":       FeatureNfs,
	// cloudresouces module CR has no feature
	"gcpnfsvolumebackup.cloud-resources.kyma-project.io":  FeatureNfsBackup,
	"gcpnfsvolumerestore.cloud-resources.kyma-project.io": FeatureNfsBackup,
	"gcpnfsvolume.cloud-resources.kyma-project.io":        FeatureNfs,
	// iprange has no feature, see comment from KCP about it
}

func featureDeterminatorByKindGroup(mi *manifestInfo) (bool, FeatureName) {
	f, ok := featuresByKindGroup[mi.kindGroup]
	return ok, f
}

func featureDeterminatorByCrdKindGroup(mi *manifestInfo) (bool, FeatureName) {
	f, ok := featuresByKindGroup[mi.crdKindGroup]
	return ok, f
}

func featureDeterminatorByBusolaKindGroup(mi *manifestInfo) (bool, FeatureName) {
	f, ok := featuresByKindGroup[mi.busolaKindGroup]
	return ok, f
}

func ManifestResourceToFeature(obj client.Object, scheme *runtime.Scheme) FeatureName {
	ffCtx := ContextBuilderFromCtx(context.Background()).
		Object(obj, scheme).
		FFCtx()

	intfToString := func(x interface{}) string {
		if x == nil {
			return ""
		}
		return x.(string)
	}

	mi := &manifestInfo{
		obj:             obj,
		name:            obj.GetName(),
		namespace:       obj.GetNamespace(),
		kindGroup:       intfToString(ffCtx.GetCustom()[KeyKindGroup]),
		labels:          obj.GetLabels(),
		annotations:     obj.GetAnnotations(),
		crdKindGroup:    intfToString(ffCtx.GetCustom()[KeyCrdKindGroup]),
		busolaKindGroup: intfToString(ffCtx.GetCustom()[KeyBusolaKindGroup]),
	}

	for _, det := range featureDeterminators {
		ok, f := det(mi)
		if ok {
			return f
		}
	}

	return ""
}

//func ManifestResourceToFeature_OLD(obj client.Object, scheme *runtime.Scheme) FeatureName {
//	var err error
//	gvk := obj.GetObjectKind().GroupVersionKind()
//	if gvk.Kind == "" {
//		gvk, err = apiutil.GVKForObject(obj, scheme)
//		if err != nil {
//			return ""
//		}
//	}
//
//	mi := &manifestInfo{
//		obj:         obj,
//		name:        obj.GetName(),
//		namespace:   obj.GetNamespace(),
//		group:       gvk.Group,
//		kind:        strings.ToLower(gvk.Kind),
//		labels:      obj.GetLabels(),
//		annotations: obj.GetAnnotations(),
//		crdGroup:    "",
//		crdKind:     "",
//	}
//
//	if mi.group == "apiextensions.k8s.io" && mi.kind == "customresourcedefinition" {
//		if u, ok := obj.(*unstructured.Unstructured); ok {
//			crdGroup, groupFound, groupErr := unstructured.NestedString(u.Object, "spec", "group")
//			crdKind, kindFound, kindErr := unstructured.NestedString(u.Object, "spec", "names", "kind")
//			if groupFound && kindFound && groupErr == nil && kindErr == nil {
//				mi.crdGroup = crdGroup
//				mi.crdKind = strings.ToLower(crdKind)
//			}
//		}
//		if crd, ok := obj.(*apiextensions.CustomResourceDefinition); ok {
//			crdGroup := crd.Spec.Group
//			crdKind := crd.Spec.Names.Kind
//			mi.crdGroup = crdGroup
//			mi.crdKind = strings.ToLower(crdKind)
//		}
//	}
//
//	if mi.group == "" && mi.kind == "configmap" &&
//		obj.GetLabels() != nil && obj.GetLabels()["busola.io/extension"] != "" {
//
//		var general string
//		if cm, ok := obj.(*unstructured.Unstructured); ok {
//			gen, found, err := unstructured.NestedString(cm.Object, "data", "general")
//			if found && err == nil {
//				general = gen
//			}
//		}
//		if cm, ok := obj.(*corev1.ConfigMap); ok {
//			gen, found := cm.Data["general"]
//			if found {
//				general = gen
//			}
//		}
//
//		if len(general) > 0 {
//			obj := map[string]interface{}{}
//			if err := yaml.Unmarshal([]byte(general), &obj); err == nil {
//				cmGroup, groupFound, groupErr := unstructured.NestedString(obj, "resource", "group")
//				cmKind, kindFound, kindErr := unstructured.NestedString(obj, "resource", "kind")
//				if groupFound && kindFound && groupErr == nil && kindErr == nil {
//					mi.busolaGroup = cmGroup
//					mi.busolaKind = strings.ToLower(cmKind)
//				}
//			}
//		}
//	}
//
//	for _, det := range featureDeterminators {
//		ok, f := det(mi)
//		if ok {
//			return f
//		}
//	}
//
//	return ""
//}
