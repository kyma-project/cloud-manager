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

func ObjectToFeature(obj client.Object, scheme *runtime.Scheme) FeatureName {
	ffCtx := ContextBuilderFromCtx(context.Background()).
		KindsFromObject(obj, scheme).
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
