package feature

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/feature/types"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

/*
All feature determination implementation in this file is used only in case
object does not implement types.FeatureAware interface
*/

type featureDeterminator func(mi *manifestInfo) (bool, types.FeatureName)

type manifestInfo struct {
	obj             client.Object
	name            string
	namespace       string
	objKindGroup    string
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

var featuresByKindGroup = map[string]types.FeatureName{
	// Please respect the order of the filenames in config/crd/bases
	// and list items here in the same order as files in that dir are.

	// KCP ==============================================================
	// scope has no feature defined
	"scope.cloud-control.kyma-project.io": "",
	// iprange atm is for nfs feature only, but in the future other features might include it as well,
	// so we will define iprange feature from start as undefined
	"iprange.cloud-control.kyma-project.io":     "",
	"nfsinstance.cloud-control.kyma-project.io": types.FeatureNfs,
	"vpcpeering.cloud-control.kyma-project.io":  types.FeaturePeering,

	// SKR ==============================================================

	"awsnfsvolumebackup.cloud-resources.kyma-project.io": types.FeatureNfsBackup,
	"awsnfsvolume.cloud-resources.kyma-project.io":       types.FeatureNfs,
	// cloudresouces module CR has no feature
	"cloudresouces.cloud-control.kyma-project.io":         "",
	"gcpnfsvolumebackup.cloud-resources.kyma-project.io":  types.FeatureNfsBackup,
	"gcpnfsvolumerestore.cloud-resources.kyma-project.io": types.FeatureNfsBackup,
	"gcpnfsvolume.cloud-resources.kyma-project.io":        types.FeatureNfs,
	// iprange has no feature, see comment from KCP about it
	"iprange.cloud-resources.kyma-project.io": "",
}

func featureDeterminatorByKindGroup(mi *manifestInfo) (bool, types.FeatureName) {
	f, ok := featuresByKindGroup[mi.objKindGroup]
	return ok, f
}

func featureDeterminatorByCrdKindGroup(mi *manifestInfo) (bool, types.FeatureName) {
	f, ok := featuresByKindGroup[mi.crdKindGroup]
	return ok, f
}

func featureDeterminatorByBusolaKindGroup(mi *manifestInfo) (bool, types.FeatureName) {
	f, ok := featuresByKindGroup[mi.busolaKindGroup]
	return ok, f
}

// ObjectToFeature returns FeatureName specific to the object. It can return empty string
// which means object is not specific to one certain feature, but to many or all.
// It first checks if object implements types.FeatureAware. If not then feature is determined
// by fixed, hard coded determinations by its kind and group. If object has TypeMeta it will
// be used to find GVK, otherwise provided scheme is looked up.
func ObjectToFeature(obj client.Object, scheme *runtime.Scheme) types.FeatureName {
	if fa, ok := obj.(types.FeatureAware); ok {
		return fa.SpecificToFeature()
	}
	return objectToFeaturePredetermined(obj, scheme)
}

func objectToFeaturePredetermined(obj client.Object, scheme *runtime.Scheme) types.FeatureName {
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
		objKindGroup:    intfToString(ffCtx.GetCustom()[types.KeyObjKindGroup]),
		labels:          obj.GetLabels(),
		annotations:     obj.GetAnnotations(),
		crdKindGroup:    intfToString(ffCtx.GetCustom()[types.KeyCrdKindGroup]),
		busolaKindGroup: intfToString(ffCtx.GetCustom()[types.KeyBusolaKindGroup]),
	}

	for _, det := range featureDeterminators {
		ok, f := det(mi)
		if ok {
			return f
		}
	}

	return ""
}
