package feature

import (
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/objkind"
	commonscheme "github.com/kyma-project/cloud-manager/pkg/common/scheme"
	"github.com/kyma-project/cloud-manager/pkg/feature/types"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"testing"
)

func TestManifestResourceToFeature(t *testing.T) {

	t.Run("ObjectToFeature from FeatureAwareObject", func(t *testing.T) {
		emptyScheme := runtime.NewScheme()
		objList := []struct {
			title    string
			obj      client.Object
			expected types.FeatureName
		}{
			{"AwsNfsVolume", &cloudresourcesv1beta1.AwsNfsVolume{}, types.FeatureNfs},
			{"AwsNfsVolumeBackup", &cloudresourcesv1beta1.AwsNfsVolumeBackup{}, types.FeatureNfsBackup},
			{"AwsVpcPeering", &cloudresourcesv1beta1.AwsVpcPeering{}, types.FeaturePeering},
			{"AzureVpcPeering", &cloudresourcesv1beta1.AzureVpcPeering{}, types.FeaturePeering},
			{"CloudResources", &cloudresourcesv1beta1.CloudResources{}, ""},
			{"GcpNfsVolumeBackup", &cloudresourcesv1beta1.GcpNfsVolumeBackup{}, types.FeatureNfsBackup},
			{"GcpNfsVolumeRestore", &cloudresourcesv1beta1.GcpNfsVolumeRestore{}, types.FeatureNfsBackup},
			{"GcpNfsVolume", &cloudresourcesv1beta1.GcpNfsVolume{}, types.FeatureNfs},
			{"GcpRedisInstance", &cloudresourcesv1beta1.GcpRedisInstance{}, types.FeatureRedis},
			{"GcpRedisInstance", &cloudresourcesv1beta1.AwsRedisInstance{}, types.FeatureRedis},
			{"GcpVpcPeering", &cloudresourcesv1beta1.GcpVpcPeering{}, types.FeaturePeering},
			{"IpRange", &cloudresourcesv1beta1.IpRange{}, ""},
		}

		for _, info := range objList {
			t.Run(info.title, func(t *testing.T) {
				actual := ObjectToFeature(info.obj, emptyScheme)
				assert.Equal(t, info.expected, actual)
			})
		}
	})

	t.Run("ObjectToFeature from Unstructured", func(t *testing.T) {
		g := cloudresourcesv1beta1.GroupVersion.Group
		v := cloudresourcesv1beta1.GroupVersion.Version

		objList := []struct {
			title    string
			obj      *unstructured.Unstructured
			expected types.FeatureName
		}{
			{"AwsNfsVolume", objkind.NewUnstructuredWithGVK(g, v, "AwsNfsVolume"), types.FeatureNfs},
			{"AwsNfsVolumeBackup", objkind.NewUnstructuredWithGVK(g, v, "AwsNfsVolumeBackup"), types.FeatureNfsBackup},
			{"AwsVpcPeering", objkind.NewUnstructuredWithGVK(g, v, "AwsVpcPeering"), types.FeaturePeering},
			{"AzureVpcPeering", objkind.NewUnstructuredWithGVK(g, v, "AzureVpcPeering"), types.FeaturePeering},
			{"CloudResources", objkind.NewUnstructuredWithGVK(g, v, "CloudResources"), ""},
			{"GcpNfsVolumeBackup", objkind.NewUnstructuredWithGVK(g, v, "GcpNfsVolumeBackup"), types.FeatureNfsBackup},
			{"GcpNfsVolumeRestore", objkind.NewUnstructuredWithGVK(g, v, "GcpNfsVolumeRestore"), types.FeatureNfsBackup},
			{"GcpNfsVolume", objkind.NewUnstructuredWithGVK(g, v, "GcpNfsVolume"), types.FeatureNfs},
			{"GcpRedisInstance", objkind.NewUnstructuredWithGVK(g, v, "GcpRedisInstance"), types.FeatureRedis},
			{"GcpRedisInstance", objkind.NewUnstructuredWithGVK(g, v, "AwsRedisInstance"), types.FeatureRedis},
			{"GcpVpcPeering", objkind.NewUnstructuredWithGVK(g, v, "GcpVpcPeering"), types.FeaturePeering},
			{"IpRange", objkind.NewUnstructuredWithGVK(g, v, "IpRange"), ""},
		}

		for _, info := range objList {
			t.Run(info.title, func(t *testing.T) {
				actual := ObjectToFeature(info.obj, commonscheme.SkrScheme)
				assert.Equal(t, info.expected, actual)
			})
		}
	})

	t.Run("ObjectToFeature from CRD Unstructured", func(t *testing.T) {
		g := cloudresourcesv1beta1.GroupVersion.Group

		baseCrdUnstructured := &unstructured.Unstructured{Object: map[string]interface{}{}}
		baseCrdUnstructured.SetAPIVersion("apiextensions.k8s.io/v1")
		baseCrdUnstructured.SetKind("CustomResourceDefinition")

		objList := []struct {
			title    string
			obj      *unstructured.Unstructured
			expected types.FeatureName
		}{
			{"AwsNfsVolume", objkind.NewCrdUnstructuredWithKindGroup(t, baseCrdUnstructured, "AwsNfsVolume", g), types.FeatureNfs},
			{"AwsNfsVolumeBackup", objkind.NewCrdUnstructuredWithKindGroup(t, baseCrdUnstructured, "AwsNfsVolumeBackup", g), types.FeatureNfsBackup},
			{"AwsVpcPeering", objkind.NewCrdUnstructuredWithKindGroup(t, baseCrdUnstructured, "AwsVpcPeering", g), types.FeaturePeering},
			{"AzureVpcPeering", objkind.NewCrdUnstructuredWithKindGroup(t, baseCrdUnstructured, "AzureVpcPeering", g), types.FeaturePeering},
			{"CloudResources", objkind.NewCrdUnstructuredWithKindGroup(t, baseCrdUnstructured, "CloudResources", g), ""},
			{"GcpNfsVolumeBackup", objkind.NewCrdUnstructuredWithKindGroup(t, baseCrdUnstructured, "GcpNfsVolumeBackup", g), types.FeatureNfsBackup},
			{"GcpNfsVolumeRestore", objkind.NewCrdUnstructuredWithKindGroup(t, baseCrdUnstructured, "GcpNfsVolumeRestore", g), types.FeatureNfsBackup},
			{"GcpNfsVolume", objkind.NewCrdUnstructuredWithKindGroup(t, baseCrdUnstructured, "GcpNfsVolume", g), types.FeatureNfs},
			{"GcpRedisInstance", objkind.NewCrdUnstructuredWithKindGroup(t, baseCrdUnstructured, "GcpRedisInstance", g), types.FeatureRedis},
			{"GcpRedisInstance", objkind.NewCrdUnstructuredWithKindGroup(t, baseCrdUnstructured, "AwsRedisInstance", g), types.FeatureRedis},
			{"GcpVpcPeering", objkind.NewCrdUnstructuredWithKindGroup(t, baseCrdUnstructured, "GcpVpcPeering", g), types.FeaturePeering},
			{"IpRange", objkind.NewCrdUnstructuredWithKindGroup(t, baseCrdUnstructured, "IpRange", g), ""},
		}

		for _, info := range objList {
			t.Run(info.title, func(t *testing.T) {
				actual := ObjectToFeature(info.obj, commonscheme.SkrScheme)
				assert.Equal(t, info.expected, actual)
			})
		}
	})

	t.Run("ObjectToFeature from Busola CM Unstructured", func(t *testing.T) {
		baseCrdUnstructured := &unstructured.Unstructured{Object: map[string]interface{}{}}
		baseCrdUnstructured.SetAPIVersion("apiextensions.k8s.io/v1")
		baseCrdUnstructured.SetKind("CustomResourceDefinition")

		objList := []struct {
			title    string
			obj      *unstructured.Unstructured
			expected types.FeatureName
		}{
			{"AwsNfsVolume", objkind.NewBusolaCmUnstructuredKindGroup(t, "AwsNfsVolume"), types.FeatureNfs},
			{"AwsNfsVolumeBackup", objkind.NewBusolaCmUnstructuredKindGroup(t, "AwsNfsVolumeBackup"), types.FeatureNfsBackup},
			{"AwsVpcPeering", objkind.NewBusolaCmUnstructuredKindGroup(t, "AwsVpcPeering"), types.FeaturePeering},
			{"AzureVpcPeering", objkind.NewBusolaCmUnstructuredKindGroup(t, "AzureVpcPeering"), types.FeaturePeering},
			{"CloudResources", objkind.NewBusolaCmUnstructuredKindGroup(t, "CloudResources"), ""},
			{"GcpNfsVolumeBackup", objkind.NewBusolaCmUnstructuredKindGroup(t, "GcpNfsVolumeBackup"), types.FeatureNfsBackup},
			{"GcpNfsVolumeRestore", objkind.NewBusolaCmUnstructuredKindGroup(t, "GcpNfsVolumeRestore"), types.FeatureNfsBackup},
			{"GcpNfsVolume", objkind.NewBusolaCmUnstructuredKindGroup(t, "GcpNfsVolume"), types.FeatureNfs},
			{"GcpRedisInstance", objkind.NewBusolaCmUnstructuredKindGroup(t, "GcpRedisInstance"), types.FeatureRedis},
			{"GcpRedisInstance", objkind.NewBusolaCmUnstructuredKindGroup(t, "AwsRedisInstance"), types.FeatureRedis},
			{"GcpVpcPeering", objkind.NewBusolaCmUnstructuredKindGroup(t, "GcpVpcPeering"), types.FeaturePeering},
			{"IpRange", objkind.NewBusolaCmUnstructuredKindGroup(t, "IpRange"), ""},
		}

		for _, info := range objList {
			t.Run(info.title, func(t *testing.T) {
				actual := ObjectToFeature(info.obj, commonscheme.SkrScheme)
				assert.Equal(t, info.expected, actual)
			})
		}
	})

}
