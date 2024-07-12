package feature

import (
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/feature/types"
	"github.com/stretchr/testify/assert"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"regexp"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
)

var schemeIgnoreRegex *regexp.Regexp

func init() {
	schemeIgnoreRegex = regexp.MustCompile("^CreateOptions|GetOptions|ListOptions|WatchOptions|PatchOptions|DeleteOptions|UpdateOptions|WatchEvent.*|.*List$")
}

func TestCloudResourcesObjectsImplementFeatureAwareObject(t *testing.T) {
	skrScheme := runtime.NewScheme()
	utilruntime.Must(scheme.AddToScheme(skrScheme))
	utilruntime.Must(cloudresourcesv1beta1.AddToScheme(skrScheme))
	utilruntime.Must(apiextensions.AddToScheme(skrScheme))

	for gvk := range skrScheme.AllKnownTypes() {
		if gvk.Group == cloudresourcesv1beta1.GroupVersion.Group {
			if schemeIgnoreRegex.Match([]byte(gvk.Kind)) {
				continue
			}

			t.Run(gvk.Kind, func(t *testing.T) {
				obj, err := skrScheme.New(gvk)
				assert.NoError(t, err)
				assert.Implements(t, (*types.FeatureAwareObject)(nil), obj, "does not implement FeatureAwareObject")
			})
		}
	}
}

func TestCloudResourcesObjectsImplementProviderAwareObject(t *testing.T) {
	skrScheme := runtime.NewScheme()
	utilruntime.Must(scheme.AddToScheme(skrScheme))
	utilruntime.Must(cloudresourcesv1beta1.AddToScheme(skrScheme))
	utilruntime.Must(apiextensions.AddToScheme(skrScheme))

	for gvk := range skrScheme.AllKnownTypes() {
		if gvk.Group == cloudresourcesv1beta1.GroupVersion.Group {
			if schemeIgnoreRegex.Match([]byte(gvk.Kind)) {
				continue
			}

			t.Run(gvk.Kind, func(t *testing.T) {
				obj, err := skrScheme.New(gvk)
				assert.NoError(t, err)
				assert.Implements(t, (*types.ProviderAwareObject)(nil), obj, "does not implement ProviderAwareObject")
			})
		}
	}
}

func TestObjectGroupVersionInfo(t *testing.T) {
	skrScheme := runtime.NewScheme()
	utilruntime.Must(scheme.AddToScheme(skrScheme))
	utilruntime.Must(cloudresourcesv1beta1.AddToScheme(skrScheme))
	utilruntime.Must(apiextensions.AddToScheme(skrScheme))

	baseCrdTyped := &apiextensions.CustomResourceDefinition{}

	gCrd := "apiextensions.k8s.io"
	kCrd := "CustomResourceDefinition"
	baseCrdUnstructured := &unstructured.Unstructured{Object: map[string]interface{}{}}
	baseCrdUnstructured.SetAPIVersion(gCrd + "/v1")
	baseCrdUnstructured.SetKind(kCrd)

	gCm := ""
	kCm := "ConfigMap"

	g := cloudresourcesv1beta1.GroupVersion.Group
	v := cloudresourcesv1beta1.GroupVersion.Version

	objList := []struct {
		title    string
		obj      client.Object
		objGK    schema.GroupKind
		crdGK    schema.GroupKind
		busolaGK schema.GroupKind
	}{
		{"typed AwsNfsVolume", &cloudresourcesv1beta1.AwsNfsVolume{}, schema.GroupKind{Group: g, Kind: "AwsNfsVolume"}, schema.GroupKind{}, schema.GroupKind{}},
		{"typed AwsNfsVolumeBackup", &cloudresourcesv1beta1.AwsNfsVolumeBackup{}, schema.GroupKind{Group: g, Kind: "AwsNfsVolumeBackup"}, schema.GroupKind{}, schema.GroupKind{}},
		{"typed AwsVpcPeering", &cloudresourcesv1beta1.AwsVpcPeering{}, schema.GroupKind{Group: g, Kind: "AwsVpcPeering"}, schema.GroupKind{}, schema.GroupKind{}},
		{"typed AzureVpcPeering", &cloudresourcesv1beta1.AzureVpcPeering{}, schema.GroupKind{Group: g, Kind: "AzureVpcPeering"}, schema.GroupKind{}, schema.GroupKind{}},
		{"typed CloudResources", &cloudresourcesv1beta1.CloudResources{}, schema.GroupKind{Group: g, Kind: "CloudResources"}, schema.GroupKind{}, schema.GroupKind{}},
		{"typed GcpNfsVolumeBackup", &cloudresourcesv1beta1.GcpNfsVolumeBackup{}, schema.GroupKind{Group: g, Kind: "GcpNfsVolumeBackup"}, schema.GroupKind{}, schema.GroupKind{}},
		{"typed GcpNfsVolumeRestore", &cloudresourcesv1beta1.GcpNfsVolumeRestore{}, schema.GroupKind{Group: g, Kind: "GcpNfsVolumeRestore"}, schema.GroupKind{}, schema.GroupKind{}},
		{"typed GcpNfsVolume", &cloudresourcesv1beta1.GcpNfsVolume{}, schema.GroupKind{Group: g, Kind: "GcpNfsVolume"}, schema.GroupKind{}, schema.GroupKind{}},
		{"typed GcpRedisInstance", &cloudresourcesv1beta1.GcpRedisInstance{}, schema.GroupKind{Group: g, Kind: "GcpRedisInstance"}, schema.GroupKind{}, schema.GroupKind{}},
		{"typed GcpVpcPeering", &cloudresourcesv1beta1.GcpVpcPeering{}, schema.GroupKind{Group: g, Kind: "GcpVpcPeering"}, schema.GroupKind{}, schema.GroupKind{}},
		{"typed IpRange", &cloudresourcesv1beta1.IpRange{}, schema.GroupKind{Group: g, Kind: "IpRange"}, schema.GroupKind{}, schema.GroupKind{}},

		{"unstructured AwsNfsVolume", newUnstructuredWithGVK(g, v, "AwsNfsVolume"), schema.GroupKind{Group: g, Kind: "AwsNfsVolume"}, schema.GroupKind{}, schema.GroupKind{}},
		{"unstructured AwsNfsVolumeBackup", newUnstructuredWithGVK(g, v, "AwsNfsVolumeBackup"), schema.GroupKind{Group: g, Kind: "AwsNfsVolumeBackup"}, schema.GroupKind{}, schema.GroupKind{}},
		{"unstructured AwsVpcPeering", newUnstructuredWithGVK(g, v, "AwsVpcPeering"), schema.GroupKind{Group: g, Kind: "AwsVpcPeering"}, schema.GroupKind{}, schema.GroupKind{}},
		{"unstructured AzureVpcPeering", newUnstructuredWithGVK(g, v, "AzureVpcPeering"), schema.GroupKind{Group: g, Kind: "AzureVpcPeering"}, schema.GroupKind{}, schema.GroupKind{}},
		{"unstructured CloudResources", newUnstructuredWithGVK(g, v, "CloudResources"), schema.GroupKind{Group: g, Kind: "CloudResources"}, schema.GroupKind{}, schema.GroupKind{}},
		{"unstructured GcpNfsVolumeBackup", newUnstructuredWithGVK(g, v, "GcpNfsVolumeBackup"), schema.GroupKind{Group: g, Kind: "GcpNfsVolumeBackup"}, schema.GroupKind{}, schema.GroupKind{}},
		{"unstructured GcpNfsVolumeRestore", newUnstructuredWithGVK(g, v, "GcpNfsVolumeRestore"), schema.GroupKind{Group: g, Kind: "GcpNfsVolumeRestore"}, schema.GroupKind{}, schema.GroupKind{}},
		{"unstructured GcpNfsVolume", newUnstructuredWithGVK(g, v, "GcpNfsVolume"), schema.GroupKind{Group: g, Kind: "GcpNfsVolume"}, schema.GroupKind{}, schema.GroupKind{}},
		{"unstructured GcpRedisInstance", newUnstructuredWithGVK(g, v, "GcpRedisInstance"), schema.GroupKind{Group: g, Kind: "GcpRedisInstance"}, schema.GroupKind{}, schema.GroupKind{}},
		{"unstructured GcpVpcPeering", newUnstructuredWithGVK(g, v, "GcpVpcPeering"), schema.GroupKind{Group: g, Kind: "GcpVpcPeering"}, schema.GroupKind{}, schema.GroupKind{}},
		{"unstructured IpRange", newUnstructuredWithGVK(g, v, "IpRange"), schema.GroupKind{Group: g, Kind: "IpRange"}, schema.GroupKind{}, schema.GroupKind{}},

		{"crdTyped AwsNfsVolume", crdTypedWithKindGroup(t, baseCrdTyped, "AwsNfsVolume", g), schema.GroupKind{Group: gCrd, Kind: kCrd}, schema.GroupKind{Group: g, Kind: "AwsNfsVolume"}, schema.GroupKind{}},
		{"crdTyped AwsNfsVolumeBackup", crdTypedWithKindGroup(t, baseCrdTyped, "AwsNfsVolumeBackup", g), schema.GroupKind{Group: gCrd, Kind: kCrd}, schema.GroupKind{Group: g, Kind: "AwsNfsVolumeBackup"}, schema.GroupKind{}},
		{"crdTyped AwsVpcPeering", crdTypedWithKindGroup(t, baseCrdTyped, "AwsVpcPeering", g), schema.GroupKind{Group: gCrd, Kind: kCrd}, schema.GroupKind{Group: g, Kind: "AwsVpcPeering"}, schema.GroupKind{}},
		{"crdTyped AzureVpcPeering", crdTypedWithKindGroup(t, baseCrdTyped, "AzureVpcPeering", g), schema.GroupKind{Group: gCrd, Kind: kCrd}, schema.GroupKind{Group: g, Kind: "AzureVpcPeering"}, schema.GroupKind{}},
		{"crdTyped CloudResources", crdTypedWithKindGroup(t, baseCrdTyped, "CloudResources", g), schema.GroupKind{Group: gCrd, Kind: kCrd}, schema.GroupKind{Group: g, Kind: "CloudResources"}, schema.GroupKind{}},
		{"crdTyped GcpNfsVolumeBackup", crdTypedWithKindGroup(t, baseCrdTyped, "GcpNfsVolumeBackup", g), schema.GroupKind{Group: gCrd, Kind: kCrd}, schema.GroupKind{Group: g, Kind: "GcpNfsVolumeBackup"}, schema.GroupKind{}},
		{"crdTyped GcpNfsVolumeRestore", crdTypedWithKindGroup(t, baseCrdTyped, "GcpNfsVolumeRestore", g), schema.GroupKind{Group: gCrd, Kind: kCrd}, schema.GroupKind{Group: g, Kind: "GcpNfsVolumeRestore"}, schema.GroupKind{}},
		{"crdTyped GcpNfsVolume", crdTypedWithKindGroup(t, baseCrdTyped, "GcpNfsVolume", g), schema.GroupKind{Group: gCrd, Kind: kCrd}, schema.GroupKind{Group: g, Kind: "GcpNfsVolume"}, schema.GroupKind{}},
		{"crdTyped GcpRedisInstance", crdTypedWithKindGroup(t, baseCrdTyped, "GcpRedisInstance", g), schema.GroupKind{Group: gCrd, Kind: kCrd}, schema.GroupKind{Group: g, Kind: "GcpRedisInstance"}, schema.GroupKind{}},
		{"crdTyped GcpVpcPeering", crdTypedWithKindGroup(t, baseCrdTyped, "GcpVpcPeering", g), schema.GroupKind{Group: gCrd, Kind: kCrd}, schema.GroupKind{Group: g, Kind: "GcpVpcPeering"}, schema.GroupKind{}},
		{"crdTyped IpRange", crdTypedWithKindGroup(t, baseCrdTyped, "IpRange", g), schema.GroupKind{Group: gCrd, Kind: kCrd}, schema.GroupKind{Group: g, Kind: "IpRange"}, schema.GroupKind{}},

		{"crdUnstructured AwsNfsVolume", crdUnstructuredWithKindGroup(t, baseCrdUnstructured, "AwsNfsVolume", g), schema.GroupKind{Group: gCrd, Kind: kCrd}, schema.GroupKind{Group: g, Kind: "AwsNfsVolume"}, schema.GroupKind{}},
		{"crdUnstructured AwsNfsVolumeBackup", crdUnstructuredWithKindGroup(t, baseCrdUnstructured, "AwsNfsVolumeBackup", g), schema.GroupKind{Group: gCrd, Kind: kCrd}, schema.GroupKind{Group: g, Kind: "AwsNfsVolumeBackup"}, schema.GroupKind{}},
		{"crdUnstructured AwsVpcPeering", crdUnstructuredWithKindGroup(t, baseCrdUnstructured, "AwsVpcPeering", g), schema.GroupKind{Group: gCrd, Kind: kCrd}, schema.GroupKind{Group: g, Kind: "AwsVpcPeering"}, schema.GroupKind{}},
		{"crdUnstructured AzureVpcPeering", crdUnstructuredWithKindGroup(t, baseCrdUnstructured, "AzureVpcPeering", g), schema.GroupKind{Group: gCrd, Kind: kCrd}, schema.GroupKind{Group: g, Kind: "AzureVpcPeering"}, schema.GroupKind{}},
		{"crdUnstructured CloudResources", crdUnstructuredWithKindGroup(t, baseCrdUnstructured, "CloudResources", g), schema.GroupKind{Group: gCrd, Kind: kCrd}, schema.GroupKind{Group: g, Kind: "CloudResources"}, schema.GroupKind{}},
		{"crdUnstructured GcpNfsVolumeBackup", crdUnstructuredWithKindGroup(t, baseCrdUnstructured, "GcpNfsVolumeBackup", g), schema.GroupKind{Group: gCrd, Kind: kCrd}, schema.GroupKind{Group: g, Kind: "GcpNfsVolumeBackup"}, schema.GroupKind{}},
		{"crdUnstructured GcpNfsVolumeRestore", crdUnstructuredWithKindGroup(t, baseCrdUnstructured, "GcpNfsVolumeRestore", g), schema.GroupKind{Group: gCrd, Kind: kCrd}, schema.GroupKind{Group: g, Kind: "GcpNfsVolumeRestore"}, schema.GroupKind{}},
		{"crdUnstructured GcpNfsVolume", crdUnstructuredWithKindGroup(t, baseCrdUnstructured, "GcpNfsVolume", g), schema.GroupKind{Group: gCrd, Kind: kCrd}, schema.GroupKind{Group: g, Kind: "GcpNfsVolume"}, schema.GroupKind{}},
		{"crdUnstructured GcpRedisInstance", crdUnstructuredWithKindGroup(t, baseCrdUnstructured, "GcpRedisInstance", g), schema.GroupKind{Group: gCrd, Kind: kCrd}, schema.GroupKind{Group: g, Kind: "GcpRedisInstance"}, schema.GroupKind{}},
		{"crdUnstructured GcpVpcPeering", crdUnstructuredWithKindGroup(t, baseCrdUnstructured, "GcpVpcPeering", g), schema.GroupKind{Group: gCrd, Kind: kCrd}, schema.GroupKind{Group: g, Kind: "GcpVpcPeering"}, schema.GroupKind{}},
		{"crdUnstructured IpRange", crdUnstructuredWithKindGroup(t, baseCrdUnstructured, "IpRange", g), schema.GroupKind{Group: gCrd, Kind: kCrd}, schema.GroupKind{Group: g, Kind: "IpRange"}, schema.GroupKind{}},

		{"busolaTyped AwsNfsVolume", busolaCmTypedKindGroup(t, "AwsNfsVolume"), schema.GroupKind{Group: gCm, Kind: kCm}, schema.GroupKind{}, schema.GroupKind{Group: g, Kind: "AwsNfsVolume"}},
		{"busolaTyped AwsNfsVolumeBackup", busolaCmTypedKindGroup(t, "AwsNfsVolumeBackup"), schema.GroupKind{Group: gCm, Kind: kCm}, schema.GroupKind{}, schema.GroupKind{Group: g, Kind: "AwsNfsVolumeBackup"}},
		{"busolaTyped AwsVpcPeering", busolaCmTypedKindGroup(t, "AwsVpcPeering"), schema.GroupKind{Group: gCm, Kind: kCm}, schema.GroupKind{}, schema.GroupKind{Group: g, Kind: "AwsVpcPeering"}},
		{"busolaTyped AzureVpcPeering", busolaCmTypedKindGroup(t, "AzureVpcPeering"), schema.GroupKind{Group: gCm, Kind: kCm}, schema.GroupKind{}, schema.GroupKind{Group: g, Kind: "AzureVpcPeering"}},
		{"busolaTyped CloudResources", busolaCmTypedKindGroup(t, "CloudResources"), schema.GroupKind{Group: gCm, Kind: kCm}, schema.GroupKind{}, schema.GroupKind{Group: g, Kind: "CloudResources"}},
		{"busolaTyped GcpNfsVolumeBackup", busolaCmTypedKindGroup(t, "GcpNfsVolumeBackup"), schema.GroupKind{Group: gCm, Kind: kCm}, schema.GroupKind{}, schema.GroupKind{Group: g, Kind: "GcpNfsVolumeBackup"}},
		{"busolaTyped GcpNfsVolumeRestore", busolaCmTypedKindGroup(t, "GcpNfsVolumeRestore"), schema.GroupKind{Group: gCm, Kind: kCm}, schema.GroupKind{}, schema.GroupKind{Group: g, Kind: "GcpNfsVolumeRestore"}},
		{"busolaTyped GcpNfsVolume", busolaCmTypedKindGroup(t, "GcpNfsVolume"), schema.GroupKind{Group: gCm, Kind: kCm}, schema.GroupKind{}, schema.GroupKind{Group: g, Kind: "GcpNfsVolume"}},
		{"busolaTyped GcpRedisInstance", busolaCmTypedKindGroup(t, "GcpRedisInstance"), schema.GroupKind{Group: gCm, Kind: kCm}, schema.GroupKind{}, schema.GroupKind{Group: g, Kind: "GcpRedisInstance"}},
		{"busolaTyped GcpVpcPeering", busolaCmTypedKindGroup(t, "GcpVpcPeering"), schema.GroupKind{Group: gCm, Kind: kCm}, schema.GroupKind{}, schema.GroupKind{Group: g, Kind: "GcpVpcPeering"}},
		{"busolaTyped IpRange", busolaCmTypedKindGroup(t, "IpRange"), schema.GroupKind{Group: gCm, Kind: kCm}, schema.GroupKind{}, schema.GroupKind{Group: g, Kind: "IpRange"}},

		{"busolaUnstructured AwsNfsVolume", busolaCmUnstructuredKindGroup(t, "AwsNfsVolume"), schema.GroupKind{Group: gCm, Kind: kCm}, schema.GroupKind{}, schema.GroupKind{Group: g, Kind: "AwsNfsVolume"}},
		{"busolaUnstructured AwsNfsVolumeBackup", busolaCmUnstructuredKindGroup(t, "AwsNfsVolumeBackup"), schema.GroupKind{Group: gCm, Kind: kCm}, schema.GroupKind{}, schema.GroupKind{Group: g, Kind: "AwsNfsVolumeBackup"}},
		{"busolaUnstructured AwsVpcPeering", busolaCmUnstructuredKindGroup(t, "AwsVpcPeering"), schema.GroupKind{Group: gCm, Kind: kCm}, schema.GroupKind{}, schema.GroupKind{Group: g, Kind: "AwsVpcPeering"}},
		{"busolaUnstructured AzureVpcPeering", busolaCmUnstructuredKindGroup(t, "AzureVpcPeering"), schema.GroupKind{Group: gCm, Kind: kCm}, schema.GroupKind{}, schema.GroupKind{Group: g, Kind: "AzureVpcPeering"}},
		{"busolaUnstructured CloudResources", busolaCmUnstructuredKindGroup(t, "CloudResources"), schema.GroupKind{Group: gCm, Kind: kCm}, schema.GroupKind{}, schema.GroupKind{Group: g, Kind: "CloudResources"}},
		{"busolaUnstructured GcpNfsVolumeBackup", busolaCmUnstructuredKindGroup(t, "GcpNfsVolumeBackup"), schema.GroupKind{Group: gCm, Kind: kCm}, schema.GroupKind{}, schema.GroupKind{Group: g, Kind: "GcpNfsVolumeBackup"}},
		{"busolaUnstructured GcpNfsVolumeRestore", busolaCmUnstructuredKindGroup(t, "GcpNfsVolumeRestore"), schema.GroupKind{Group: gCm, Kind: kCm}, schema.GroupKind{}, schema.GroupKind{Group: g, Kind: "GcpNfsVolumeRestore"}},
		{"busolaUnstructured GcpNfsVolume", busolaCmUnstructuredKindGroup(t, "GcpNfsVolume"), schema.GroupKind{Group: gCm, Kind: kCm}, schema.GroupKind{}, schema.GroupKind{Group: g, Kind: "GcpNfsVolume"}},
		{"busolaUnstructured GcpRedisInstance", busolaCmUnstructuredKindGroup(t, "GcpRedisInstance"), schema.GroupKind{Group: gCm, Kind: kCm}, schema.GroupKind{}, schema.GroupKind{Group: g, Kind: "GcpRedisInstance"}},
		{"busolaUnstructured GcpVpcPeering", busolaCmUnstructuredKindGroup(t, "GcpVpcPeering"), schema.GroupKind{Group: gCm, Kind: kCm}, schema.GroupKind{}, schema.GroupKind{Group: g, Kind: "GcpVpcPeering"}},
		{"busolaUnstructured IpRange", busolaCmUnstructuredKindGroup(t, "IpRange"), schema.GroupKind{Group: gCm, Kind: kCm}, schema.GroupKind{}, schema.GroupKind{Group: g, Kind: "IpRange"}},
	}

	for _, expected := range objList {
		t.Run(expected.title, func(t *testing.T) {
			actual := ObjectKinds(expected.obj, skrScheme)
			assert.Equal(t, expected.objGK.String(), actual.ObjGK.String(), "unexpected ObjGK")
			assert.Equal(t, expected.crdGK.String(), actual.CrdGK.String(), "unexpected CrdGK")
			assert.Equal(t, expected.busolaGK.String(), actual.BusolaGK.String(), "unexpected BusolaGK")
			if len(expected.crdGK.Kind) == 0 {
				assert.False(t, actual.CrdOK, "unexpected CrdOK")
			} else {
				assert.True(t, actual.CrdOK, "unexpected CrdOK")
			}
			if len(expected.busolaGK.Kind) == 0 {
				assert.False(t, actual.BusolaOK, "unexpected BusolaOK")
			} else {
				assert.True(t, actual.BusolaOK, "unexpected BusolaOK")
			}
		})
	}
}
