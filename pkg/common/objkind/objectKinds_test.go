package objkind

import (
	"regexp"
	"testing"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	commonscheme "github.com/kyma-project/cloud-manager/pkg/common/scheme"
	"github.com/kyma-project/cloud-manager/pkg/feature/types"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var schemeIgnoreRegex *regexp.Regexp

func init() {
	schemeIgnoreRegex = regexp.MustCompile("^CreateOptions|GetOptions|ListOptions|WatchOptions|PatchOptions|DeleteOptions|UpdateOptions|WatchEvent.*|.*List$")
}

func TestCloudResourcesObjectsImplementFeatureAwareObject(t *testing.T) {

	for gvk := range commonscheme.SkrScheme.AllKnownTypes() {
		if gvk.Group == cloudresourcesv1beta1.GroupVersion.Group {
			if schemeIgnoreRegex.Match([]byte(gvk.Kind)) {
				continue
			}

			t.Run(gvk.Kind, func(t *testing.T) {
				obj, err := commonscheme.SkrScheme.New(gvk)
				assert.NoError(t, err)
				assert.Implements(t, (*types.FeatureAwareObject)(nil), obj, "does not implement FeatureAwareObject")
			})
		}
	}
}

func TestCloudResourcesObjectsImplementProviderAwareObject(t *testing.T) {
	for gvk := range commonscheme.SkrScheme.AllKnownTypes() {
		if gvk.Group == cloudresourcesv1beta1.GroupVersion.Group {
			if schemeIgnoreRegex.Match([]byte(gvk.Kind)) {
				continue
			}

			t.Run(gvk.Kind, func(t *testing.T) {
				obj, err := commonscheme.SkrScheme.New(gvk)
				assert.NoError(t, err)
				assert.Implements(t, (*types.ProviderAwareObject)(nil), obj, "does not implement ProviderAwareObject")
			})
		}
	}
}

func TestObjectGroupVersionInfo(t *testing.T) {
	skrScheme := commonscheme.SkrScheme

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

		{"unstructured AwsNfsVolume", NewUnstructuredWithGVK(g, v, "AwsNfsVolume"), schema.GroupKind{Group: g, Kind: "AwsNfsVolume"}, schema.GroupKind{}, schema.GroupKind{}},
		{"unstructured AwsNfsVolumeBackup", NewUnstructuredWithGVK(g, v, "AwsNfsVolumeBackup"), schema.GroupKind{Group: g, Kind: "AwsNfsVolumeBackup"}, schema.GroupKind{}, schema.GroupKind{}},
		{"unstructured AwsVpcPeering", NewUnstructuredWithGVK(g, v, "AwsVpcPeering"), schema.GroupKind{Group: g, Kind: "AwsVpcPeering"}, schema.GroupKind{}, schema.GroupKind{}},
		{"unstructured AzureVpcPeering", NewUnstructuredWithGVK(g, v, "AzureVpcPeering"), schema.GroupKind{Group: g, Kind: "AzureVpcPeering"}, schema.GroupKind{}, schema.GroupKind{}},
		{"unstructured CloudResources", NewUnstructuredWithGVK(g, v, "CloudResources"), schema.GroupKind{Group: g, Kind: "CloudResources"}, schema.GroupKind{}, schema.GroupKind{}},
		{"unstructured GcpNfsVolumeBackup", NewUnstructuredWithGVK(g, v, "GcpNfsVolumeBackup"), schema.GroupKind{Group: g, Kind: "GcpNfsVolumeBackup"}, schema.GroupKind{}, schema.GroupKind{}},
		{"unstructured GcpNfsVolumeRestore", NewUnstructuredWithGVK(g, v, "GcpNfsVolumeRestore"), schema.GroupKind{Group: g, Kind: "GcpNfsVolumeRestore"}, schema.GroupKind{}, schema.GroupKind{}},
		{"unstructured GcpNfsVolume", NewUnstructuredWithGVK(g, v, "GcpNfsVolume"), schema.GroupKind{Group: g, Kind: "GcpNfsVolume"}, schema.GroupKind{}, schema.GroupKind{}},
		{"unstructured GcpRedisInstance", NewUnstructuredWithGVK(g, v, "GcpRedisInstance"), schema.GroupKind{Group: g, Kind: "GcpRedisInstance"}, schema.GroupKind{}, schema.GroupKind{}},
		{"unstructured GcpVpcPeering", NewUnstructuredWithGVK(g, v, "GcpVpcPeering"), schema.GroupKind{Group: g, Kind: "GcpVpcPeering"}, schema.GroupKind{}, schema.GroupKind{}},
		{"unstructured IpRange", NewUnstructuredWithGVK(g, v, "IpRange"), schema.GroupKind{Group: g, Kind: "IpRange"}, schema.GroupKind{}, schema.GroupKind{}},

		{"crdTyped AwsNfsVolume", NewCrdTypedV1WithKindGroup(t, "AwsNfsVolume", g), schema.GroupKind{Group: gCrd, Kind: kCrd}, schema.GroupKind{Group: g, Kind: "AwsNfsVolume"}, schema.GroupKind{}},
		{"crdTyped AwsNfsVolumeBackup", NewCrdTypedV1WithKindGroup(t, "AwsNfsVolumeBackup", g), schema.GroupKind{Group: gCrd, Kind: kCrd}, schema.GroupKind{Group: g, Kind: "AwsNfsVolumeBackup"}, schema.GroupKind{}},
		{"crdTyped AwsVpcPeering", NewCrdTypedV1WithKindGroup(t, "AwsVpcPeering", g), schema.GroupKind{Group: gCrd, Kind: kCrd}, schema.GroupKind{Group: g, Kind: "AwsVpcPeering"}, schema.GroupKind{}},
		{"crdTyped AzureVpcPeering", NewCrdTypedV1WithKindGroup(t, "AzureVpcPeering", g), schema.GroupKind{Group: gCrd, Kind: kCrd}, schema.GroupKind{Group: g, Kind: "AzureVpcPeering"}, schema.GroupKind{}},
		{"crdTyped CloudResources", NewCrdTypedV1WithKindGroup(t, "CloudResources", g), schema.GroupKind{Group: gCrd, Kind: kCrd}, schema.GroupKind{Group: g, Kind: "CloudResources"}, schema.GroupKind{}},
		{"crdTyped GcpNfsVolumeBackup", NewCrdTypedV1WithKindGroup(t, "GcpNfsVolumeBackup", g), schema.GroupKind{Group: gCrd, Kind: kCrd}, schema.GroupKind{Group: g, Kind: "GcpNfsVolumeBackup"}, schema.GroupKind{}},
		{"crdTyped GcpNfsVolumeRestore", NewCrdTypedV1WithKindGroup(t, "GcpNfsVolumeRestore", g), schema.GroupKind{Group: gCrd, Kind: kCrd}, schema.GroupKind{Group: g, Kind: "GcpNfsVolumeRestore"}, schema.GroupKind{}},
		{"crdTyped GcpNfsVolume", NewCrdTypedV1WithKindGroup(t, "GcpNfsVolume", g), schema.GroupKind{Group: gCrd, Kind: kCrd}, schema.GroupKind{Group: g, Kind: "GcpNfsVolume"}, schema.GroupKind{}},
		{"crdTyped GcpRedisInstance", NewCrdTypedV1WithKindGroup(t, "GcpRedisInstance", g), schema.GroupKind{Group: gCrd, Kind: kCrd}, schema.GroupKind{Group: g, Kind: "GcpRedisInstance"}, schema.GroupKind{}},
		{"crdTyped GcpVpcPeering", NewCrdTypedV1WithKindGroup(t, "GcpVpcPeering", g), schema.GroupKind{Group: gCrd, Kind: kCrd}, schema.GroupKind{Group: g, Kind: "GcpVpcPeering"}, schema.GroupKind{}},
		{"crdTyped IpRange", NewCrdTypedV1WithKindGroup(t, "IpRange", g), schema.GroupKind{Group: gCrd, Kind: kCrd}, schema.GroupKind{Group: g, Kind: "IpRange"}, schema.GroupKind{}},

		{"crdUnstructured AwsNfsVolume", NewCrdUnstructuredWithKindGroup(t, baseCrdUnstructured, "AwsNfsVolume", g), schema.GroupKind{Group: gCrd, Kind: kCrd}, schema.GroupKind{Group: g, Kind: "AwsNfsVolume"}, schema.GroupKind{}},
		{"crdUnstructured AwsNfsVolumeBackup", NewCrdUnstructuredWithKindGroup(t, baseCrdUnstructured, "AwsNfsVolumeBackup", g), schema.GroupKind{Group: gCrd, Kind: kCrd}, schema.GroupKind{Group: g, Kind: "AwsNfsVolumeBackup"}, schema.GroupKind{}},
		{"crdUnstructured AwsVpcPeering", NewCrdUnstructuredWithKindGroup(t, baseCrdUnstructured, "AwsVpcPeering", g), schema.GroupKind{Group: gCrd, Kind: kCrd}, schema.GroupKind{Group: g, Kind: "AwsVpcPeering"}, schema.GroupKind{}},
		{"crdUnstructured AzureVpcPeering", NewCrdUnstructuredWithKindGroup(t, baseCrdUnstructured, "AzureVpcPeering", g), schema.GroupKind{Group: gCrd, Kind: kCrd}, schema.GroupKind{Group: g, Kind: "AzureVpcPeering"}, schema.GroupKind{}},
		{"crdUnstructured CloudResources", NewCrdUnstructuredWithKindGroup(t, baseCrdUnstructured, "CloudResources", g), schema.GroupKind{Group: gCrd, Kind: kCrd}, schema.GroupKind{Group: g, Kind: "CloudResources"}, schema.GroupKind{}},
		{"crdUnstructured GcpNfsVolumeBackup", NewCrdUnstructuredWithKindGroup(t, baseCrdUnstructured, "GcpNfsVolumeBackup", g), schema.GroupKind{Group: gCrd, Kind: kCrd}, schema.GroupKind{Group: g, Kind: "GcpNfsVolumeBackup"}, schema.GroupKind{}},
		{"crdUnstructured GcpNfsVolumeRestore", NewCrdUnstructuredWithKindGroup(t, baseCrdUnstructured, "GcpNfsVolumeRestore", g), schema.GroupKind{Group: gCrd, Kind: kCrd}, schema.GroupKind{Group: g, Kind: "GcpNfsVolumeRestore"}, schema.GroupKind{}},
		{"crdUnstructured GcpNfsVolume", NewCrdUnstructuredWithKindGroup(t, baseCrdUnstructured, "GcpNfsVolume", g), schema.GroupKind{Group: gCrd, Kind: kCrd}, schema.GroupKind{Group: g, Kind: "GcpNfsVolume"}, schema.GroupKind{}},
		{"crdUnstructured GcpRedisInstance", NewCrdUnstructuredWithKindGroup(t, baseCrdUnstructured, "GcpRedisInstance", g), schema.GroupKind{Group: gCrd, Kind: kCrd}, schema.GroupKind{Group: g, Kind: "GcpRedisInstance"}, schema.GroupKind{}},
		{"crdUnstructured GcpVpcPeering", NewCrdUnstructuredWithKindGroup(t, baseCrdUnstructured, "GcpVpcPeering", g), schema.GroupKind{Group: gCrd, Kind: kCrd}, schema.GroupKind{Group: g, Kind: "GcpVpcPeering"}, schema.GroupKind{}},
		{"crdUnstructured IpRange", NewCrdUnstructuredWithKindGroup(t, baseCrdUnstructured, "IpRange", g), schema.GroupKind{Group: gCrd, Kind: kCrd}, schema.GroupKind{Group: g, Kind: "IpRange"}, schema.GroupKind{}},

		{"busolaTyped AwsNfsVolume", NewBusolaCmTypedKindGroup(t, "AwsNfsVolume"), schema.GroupKind{Group: gCm, Kind: kCm}, schema.GroupKind{}, schema.GroupKind{Group: g, Kind: "AwsNfsVolume"}},
		{"busolaTyped AwsNfsVolumeBackup", NewBusolaCmTypedKindGroup(t, "AwsNfsVolumeBackup"), schema.GroupKind{Group: gCm, Kind: kCm}, schema.GroupKind{}, schema.GroupKind{Group: g, Kind: "AwsNfsVolumeBackup"}},
		{"busolaTyped AwsVpcPeering", NewBusolaCmTypedKindGroup(t, "AwsVpcPeering"), schema.GroupKind{Group: gCm, Kind: kCm}, schema.GroupKind{}, schema.GroupKind{Group: g, Kind: "AwsVpcPeering"}},
		{"busolaTyped AzureVpcPeering", NewBusolaCmTypedKindGroup(t, "AzureVpcPeering"), schema.GroupKind{Group: gCm, Kind: kCm}, schema.GroupKind{}, schema.GroupKind{Group: g, Kind: "AzureVpcPeering"}},
		{"busolaTyped CloudResources", NewBusolaCmTypedKindGroup(t, "CloudResources"), schema.GroupKind{Group: gCm, Kind: kCm}, schema.GroupKind{}, schema.GroupKind{Group: g, Kind: "CloudResources"}},
		{"busolaTyped GcpNfsVolumeBackup", NewBusolaCmTypedKindGroup(t, "GcpNfsVolumeBackup"), schema.GroupKind{Group: gCm, Kind: kCm}, schema.GroupKind{}, schema.GroupKind{Group: g, Kind: "GcpNfsVolumeBackup"}},
		{"busolaTyped GcpNfsVolumeRestore", NewBusolaCmTypedKindGroup(t, "GcpNfsVolumeRestore"), schema.GroupKind{Group: gCm, Kind: kCm}, schema.GroupKind{}, schema.GroupKind{Group: g, Kind: "GcpNfsVolumeRestore"}},
		{"busolaTyped GcpNfsVolume", NewBusolaCmTypedKindGroup(t, "GcpNfsVolume"), schema.GroupKind{Group: gCm, Kind: kCm}, schema.GroupKind{}, schema.GroupKind{Group: g, Kind: "GcpNfsVolume"}},
		{"busolaTyped GcpRedisInstance", NewBusolaCmTypedKindGroup(t, "GcpRedisInstance"), schema.GroupKind{Group: gCm, Kind: kCm}, schema.GroupKind{}, schema.GroupKind{Group: g, Kind: "GcpRedisInstance"}},
		{"busolaTyped GcpVpcPeering", NewBusolaCmTypedKindGroup(t, "GcpVpcPeering"), schema.GroupKind{Group: gCm, Kind: kCm}, schema.GroupKind{}, schema.GroupKind{Group: g, Kind: "GcpVpcPeering"}},
		{"busolaTyped IpRange", NewBusolaCmTypedKindGroup(t, "IpRange"), schema.GroupKind{Group: gCm, Kind: kCm}, schema.GroupKind{}, schema.GroupKind{Group: g, Kind: "IpRange"}},

		{"busolaUnstructured AwsNfsVolume", NewBusolaCmUnstructuredKindGroup(t, "AwsNfsVolume"), schema.GroupKind{Group: gCm, Kind: kCm}, schema.GroupKind{}, schema.GroupKind{Group: g, Kind: "AwsNfsVolume"}},
		{"busolaUnstructured AwsNfsVolumeBackup", NewBusolaCmUnstructuredKindGroup(t, "AwsNfsVolumeBackup"), schema.GroupKind{Group: gCm, Kind: kCm}, schema.GroupKind{}, schema.GroupKind{Group: g, Kind: "AwsNfsVolumeBackup"}},
		{"busolaUnstructured AwsVpcPeering", NewBusolaCmUnstructuredKindGroup(t, "AwsVpcPeering"), schema.GroupKind{Group: gCm, Kind: kCm}, schema.GroupKind{}, schema.GroupKind{Group: g, Kind: "AwsVpcPeering"}},
		{"busolaUnstructured AzureVpcPeering", NewBusolaCmUnstructuredKindGroup(t, "AzureVpcPeering"), schema.GroupKind{Group: gCm, Kind: kCm}, schema.GroupKind{}, schema.GroupKind{Group: g, Kind: "AzureVpcPeering"}},
		{"busolaUnstructured CloudResources", NewBusolaCmUnstructuredKindGroup(t, "CloudResources"), schema.GroupKind{Group: gCm, Kind: kCm}, schema.GroupKind{}, schema.GroupKind{Group: g, Kind: "CloudResources"}},
		{"busolaUnstructured GcpNfsVolumeBackup", NewBusolaCmUnstructuredKindGroup(t, "GcpNfsVolumeBackup"), schema.GroupKind{Group: gCm, Kind: kCm}, schema.GroupKind{}, schema.GroupKind{Group: g, Kind: "GcpNfsVolumeBackup"}},
		{"busolaUnstructured GcpNfsVolumeRestore", NewBusolaCmUnstructuredKindGroup(t, "GcpNfsVolumeRestore"), schema.GroupKind{Group: gCm, Kind: kCm}, schema.GroupKind{}, schema.GroupKind{Group: g, Kind: "GcpNfsVolumeRestore"}},
		{"busolaUnstructured GcpNfsVolume", NewBusolaCmUnstructuredKindGroup(t, "GcpNfsVolume"), schema.GroupKind{Group: gCm, Kind: kCm}, schema.GroupKind{}, schema.GroupKind{Group: g, Kind: "GcpNfsVolume"}},
		{"busolaUnstructured GcpRedisInstance", NewBusolaCmUnstructuredKindGroup(t, "GcpRedisInstance"), schema.GroupKind{Group: gCm, Kind: kCm}, schema.GroupKind{}, schema.GroupKind{Group: g, Kind: "GcpRedisInstance"}},
		{"busolaUnstructured GcpVpcPeering", NewBusolaCmUnstructuredKindGroup(t, "GcpVpcPeering"), schema.GroupKind{Group: gCm, Kind: kCm}, schema.GroupKind{}, schema.GroupKind{Group: g, Kind: "GcpVpcPeering"}},
		{"busolaUnstructured IpRange", NewBusolaCmUnstructuredKindGroup(t, "IpRange"), schema.GroupKind{Group: gCm, Kind: kCm}, schema.GroupKind{}, schema.GroupKind{Group: g, Kind: "IpRange"}},
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
