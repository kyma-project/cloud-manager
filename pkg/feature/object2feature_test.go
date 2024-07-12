package feature

import (
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/feature/types"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
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
		skrScheme := runtime.NewScheme()
		utilruntime.Must(scheme.AddToScheme(skrScheme))
		utilruntime.Must(cloudresourcesv1beta1.AddToScheme(skrScheme))
		utilruntime.Must(apiextensions.AddToScheme(skrScheme))

		g := cloudresourcesv1beta1.GroupVersion.Group
		v := cloudresourcesv1beta1.GroupVersion.Version

		objList := []struct {
			title    string
			obj      *unstructured.Unstructured
			expected types.FeatureName
		}{
			{"AwsNfsVolume", newUnstructuredWithGVK(g, v, "AwsNfsVolume"), types.FeatureNfs},
			{"AwsNfsVolumeBackup", newUnstructuredWithGVK(g, v, "AwsNfsVolumeBackup"), types.FeatureNfsBackup},
			{"AwsVpcPeering", newUnstructuredWithGVK(g, v, "AwsVpcPeering"), types.FeaturePeering},
			{"AzureVpcPeering", newUnstructuredWithGVK(g, v, "AzureVpcPeering"), types.FeaturePeering},
			{"CloudResources", newUnstructuredWithGVK(g, v, "CloudResources"), ""},
			{"GcpNfsVolumeBackup", newUnstructuredWithGVK(g, v, "GcpNfsVolumeBackup"), types.FeatureNfsBackup},
			{"GcpNfsVolumeRestore", newUnstructuredWithGVK(g, v, "GcpNfsVolumeRestore"), types.FeatureNfsBackup},
			{"GcpNfsVolume", newUnstructuredWithGVK(g, v, "GcpNfsVolume"), types.FeatureNfs},
			{"GcpRedisInstance", newUnstructuredWithGVK(g, v, "GcpRedisInstance"), types.FeatureRedis},
			{"GcpVpcPeering", newUnstructuredWithGVK(g, v, "GcpVpcPeering"), types.FeaturePeering},
			{"IpRange", newUnstructuredWithGVK(g, v, "IpRange"), ""},
		}

		for _, info := range objList {
			t.Run(info.title, func(t *testing.T) {
				actual := ObjectToFeature(info.obj, skrScheme)
				assert.Equal(t, info.expected, actual)
			})
		}
	})

	t.Run("ObjectToFeature from CRD Unstructured", func(t *testing.T) {
		skrScheme := runtime.NewScheme()
		utilruntime.Must(scheme.AddToScheme(skrScheme))
		utilruntime.Must(cloudresourcesv1beta1.AddToScheme(skrScheme))
		utilruntime.Must(apiextensions.AddToScheme(skrScheme))

		g := cloudresourcesv1beta1.GroupVersion.Group

		baseCrdUnstructured := &unstructured.Unstructured{Object: map[string]interface{}{}}
		baseCrdUnstructured.SetAPIVersion("apiextensions.k8s.io/v1")
		baseCrdUnstructured.SetKind("CustomResourceDefinition")

		objList := []struct {
			title    string
			obj      *unstructured.Unstructured
			expected types.FeatureName
		}{
			{"AwsNfsVolume", crdUnstructuredWithKindGroup(t, baseCrdUnstructured, "AwsNfsVolume", g), types.FeatureNfs},
			{"AwsNfsVolumeBackup", crdUnstructuredWithKindGroup(t, baseCrdUnstructured, "AwsNfsVolumeBackup", g), types.FeatureNfsBackup},
			{"AwsVpcPeering", crdUnstructuredWithKindGroup(t, baseCrdUnstructured, "AwsVpcPeering", g), types.FeaturePeering},
			{"AzureVpcPeering", crdUnstructuredWithKindGroup(t, baseCrdUnstructured, "AzureVpcPeering", g), types.FeaturePeering},
			{"CloudResources", crdUnstructuredWithKindGroup(t, baseCrdUnstructured, "CloudResources", g), ""},
			{"GcpNfsVolumeBackup", crdUnstructuredWithKindGroup(t, baseCrdUnstructured, "GcpNfsVolumeBackup", g), types.FeatureNfsBackup},
			{"GcpNfsVolumeRestore", crdUnstructuredWithKindGroup(t, baseCrdUnstructured, "GcpNfsVolumeRestore", g), types.FeatureNfsBackup},
			{"GcpNfsVolume", crdUnstructuredWithKindGroup(t, baseCrdUnstructured, "GcpNfsVolume", g), types.FeatureNfs},
			{"GcpRedisInstance", crdUnstructuredWithKindGroup(t, baseCrdUnstructured, "GcpRedisInstance", g), types.FeatureRedis},
			{"GcpVpcPeering", crdUnstructuredWithKindGroup(t, baseCrdUnstructured, "GcpVpcPeering", g), types.FeaturePeering},
			{"IpRange", crdUnstructuredWithKindGroup(t, baseCrdUnstructured, "IpRange", g), ""},
		}

		for _, info := range objList {
			t.Run(info.title, func(t *testing.T) {
				actual := ObjectToFeature(info.obj, skrScheme)
				assert.Equal(t, info.expected, actual)
			})
		}
	})

	t.Run("ObjectToFeature from Busola CM Unstructured", func(t *testing.T) {
		skrScheme := runtime.NewScheme()
		utilruntime.Must(scheme.AddToScheme(skrScheme))
		utilruntime.Must(cloudresourcesv1beta1.AddToScheme(skrScheme))
		utilruntime.Must(apiextensions.AddToScheme(skrScheme))

		baseCrdUnstructured := &unstructured.Unstructured{Object: map[string]interface{}{}}
		baseCrdUnstructured.SetAPIVersion("apiextensions.k8s.io/v1")
		baseCrdUnstructured.SetKind("CustomResourceDefinition")

		objList := []struct {
			title    string
			obj      *unstructured.Unstructured
			expected types.FeatureName
		}{
			{"AwsNfsVolume", busolaCmUnstructuredKindGroup(t, "AwsNfsVolume"), types.FeatureNfs},
			{"AwsNfsVolumeBackup", busolaCmUnstructuredKindGroup(t, "AwsNfsVolumeBackup"), types.FeatureNfsBackup},
			{"AwsVpcPeering", busolaCmUnstructuredKindGroup(t, "AwsVpcPeering"), types.FeaturePeering},
			{"AzureVpcPeering", busolaCmUnstructuredKindGroup(t, "AzureVpcPeering"), types.FeaturePeering},
			{"CloudResources", busolaCmUnstructuredKindGroup(t, "CloudResources"), ""},
			{"GcpNfsVolumeBackup", busolaCmUnstructuredKindGroup(t, "GcpNfsVolumeBackup"), types.FeatureNfsBackup},
			{"GcpNfsVolumeRestore", busolaCmUnstructuredKindGroup(t, "GcpNfsVolumeRestore"), types.FeatureNfsBackup},
			{"GcpNfsVolume", busolaCmUnstructuredKindGroup(t, "GcpNfsVolume"), types.FeatureNfs},
			{"GcpRedisInstance", busolaCmUnstructuredKindGroup(t, "GcpRedisInstance"), types.FeatureRedis},
			{"GcpVpcPeering", busolaCmUnstructuredKindGroup(t, "GcpVpcPeering"), types.FeaturePeering},
			{"IpRange", busolaCmUnstructuredKindGroup(t, "IpRange"), ""},
		}

		for _, info := range objList {
			t.Run(info.title, func(t *testing.T) {
				actual := ObjectToFeature(info.obj, skrScheme)
				assert.Equal(t, info.expected, actual)
			})
		}
	})

}

func newUnstructuredWithGVK(g, v, k string) *unstructured.Unstructured {
	x := &unstructured.Unstructured{}
	x.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   g,
		Version: v,
		Kind:    k,
	})
	return x
}

func crdUnstructuredWithKindGroup(t *testing.T, baseCrd *unstructured.Unstructured, k, g string) *unstructured.Unstructured {
	crd := baseCrd.DeepCopyObject().(*unstructured.Unstructured)
	err := unstructured.SetNestedField(crd.Object, k, "spec", "names", "kind")
	assert.NoError(t, err)
	err = unstructured.SetNestedField(crd.Object, g, "spec", "group")
	assert.NoError(t, err)
	return crd
}

func crdTypedWithKindGroup(_ *testing.T, baseCrd *apiextensions.CustomResourceDefinition, k, g string) *apiextensions.CustomResourceDefinition {
	crd := baseCrd.DeepCopyObject().(*apiextensions.CustomResourceDefinition)
	crd.Spec.Names.Kind = k
	crd.Spec.Group = g
	return crd
}

func busolaCmTypedKindGroup(t *testing.T, k string) *corev1.ConfigMap {
	data := map[string]interface{}{}
	err := unstructured.SetNestedField(data, "cloud-resources.kyma-project.io", "resource", "group")
	assert.NoError(t, err)
	err = unstructured.SetNestedField(data, k, "resource", "kind")
	assert.NoError(t, err)

	b, err := yaml.Marshal(data)
	assert.NoError(t, err)

	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"busola.io/extension": "resource",
			},
		},
		Data: map[string]string{
			"general": string(b),
		},
	}
}

func busolaCmUnstructuredKindGroup(t *testing.T, k string) *unstructured.Unstructured {
	data := map[string]interface{}{}
	err := unstructured.SetNestedField(data, "cloud-resources.kyma-project.io", "resource", "group")
	assert.NoError(t, err)
	err = unstructured.SetNestedField(data, k, "resource", "kind")
	assert.NoError(t, err)

	b, err := yaml.Marshal(data)
	assert.NoError(t, err)

	u := &unstructured.Unstructured{Object: map[string]interface{}{}}
	u.SetAPIVersion("v1")
	u.SetKind("ConfigMap")
	u.SetLabels(map[string]string{
		"busola.io/extension": "resource",
	})
	err = unstructured.SetNestedMap(u.Object, map[string]interface{}{
		"general": string(b),
	}, "data")
	assert.NoError(t, err)

	return u
}
