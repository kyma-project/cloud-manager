package feature

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/feature/types"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"testing"
)

func TestManifestResourceToFeature(t *testing.T) {

	t.Run("ObjectToFeature from FeatureAware", func(t *testing.T) {
		emptyScheme := runtime.NewScheme()
		objList := []struct {
			title    string
			obj      client.Object
			expected types.FeatureName
		}{
			{"AwsNfsVolume", &cloudresourcesv1beta1.AwsNfsVolume{}, types.FeatureNfs},
			{"AwsNfsVolumeBackup", &cloudresourcesv1beta1.AwsNfsVolumeBackup{}, types.FeatureNfsBackup},
			{"CloudResources", &cloudresourcesv1beta1.CloudResources{}, ""},
			{"GcpNfsVolumeBackup", &cloudresourcesv1beta1.GcpNfsVolumeBackup{}, types.FeatureNfsBackup},
			{"GcpNfsVolumeRestore", &cloudresourcesv1beta1.GcpNfsVolumeRestore{}, types.FeatureNfsBackup},
			{"GcpNfsVolume", &cloudresourcesv1beta1.GcpNfsVolume{}, types.FeatureNfs},
			{"IpRange", &cloudresourcesv1beta1.IpRange{}, ""},
		}

		for _, info := range objList {
			t.Run(info.title, func(t *testing.T) {
				actual := ObjectToFeature(info.obj, emptyScheme)
				assert.Equal(t, info.expected, actual)
			})
		}
	})

	t.Run("objectToFeaturePredetermined", func(t *testing.T) {

		t.Run("From Scheme", func(t *testing.T) {
			kcpScheme := runtime.NewScheme()
			utilruntime.Must(scheme.AddToScheme(kcpScheme))
			utilruntime.Must(cloudcontrolv1beta1.AddToScheme(kcpScheme))
			utilruntime.Must(apiextensions.AddToScheme(kcpScheme))

			skrScheme := runtime.NewScheme()
			utilruntime.Must(scheme.AddToScheme(skrScheme))
			utilruntime.Must(cloudresourcesv1beta1.AddToScheme(skrScheme))
			utilruntime.Must(apiextensions.AddToScheme(skrScheme))

			baseCrdTyped := &apiextensions.CustomResourceDefinition{}

			objList := []struct {
				title    string
				scheme   *runtime.Scheme
				obj      client.Object
				expected types.FeatureName
			}{
				{"KCP IpRange", kcpScheme, &cloudcontrolv1beta1.IpRange{}, ""},
				{"KCP NfsInstance", kcpScheme, &cloudcontrolv1beta1.NfsInstance{}, types.FeatureNfs},
				{"KCP Scope", kcpScheme, &cloudcontrolv1beta1.Scope{}, ""},
				{"KCP VpcPeering", kcpScheme, &cloudcontrolv1beta1.VpcPeering{}, types.FeaturePeering},

				{"SKR AwsNfsVolumeBackup", skrScheme, &cloudresourcesv1beta1.AwsNfsVolumeBackup{}, types.FeatureNfsBackup},
				{"SKR AwsNfsVolume", skrScheme, &cloudresourcesv1beta1.AwsNfsVolume{}, types.FeatureNfs},
				{"SKR CloudResources", skrScheme, &cloudresourcesv1beta1.CloudResources{}, ""},
				{"SKR GcpNfsVolumeBackup", skrScheme, &cloudresourcesv1beta1.GcpNfsVolumeBackup{}, types.FeatureNfsBackup},
				{"SKR GcpNfsVolumeRestore", skrScheme, &cloudresourcesv1beta1.GcpNfsVolumeRestore{}, types.FeatureNfsBackup},
				{"SKR GcpNfsVolume", skrScheme, &cloudresourcesv1beta1.GcpNfsVolume{}, types.FeatureNfs},
				{"SKR IpRange", skrScheme, &cloudresourcesv1beta1.IpRange{}, ""},

				{"CRD Typed SKR AwsNfsVolumeBackup", skrScheme, crdTypedWithKindGroup(t, baseCrdTyped, "AwsNfsVolumeBackup", "cloud-resources.kyma-project.io"), types.FeatureNfsBackup},
				{"CRD Typed SKR AwsNfsVolume", skrScheme, crdTypedWithKindGroup(t, baseCrdTyped, "AwsNfsVolume", "cloud-resources.kyma-project.io"), types.FeatureNfs},
				{"CRD Typed SKR CloudResources", skrScheme, crdTypedWithKindGroup(t, baseCrdTyped, "CloudResources", "cloud-resources.kyma-project.io"), ""},
				{"CRD Typed SKR GcpNfsVolumeBackup", skrScheme, crdTypedWithKindGroup(t, baseCrdTyped, "GcpNfsVolumeBackup", "cloud-resources.kyma-project.io"), types.FeatureNfsBackup},
				{"CRD Typed SKR GcpNfsVolumeRestore", skrScheme, crdTypedWithKindGroup(t, baseCrdTyped, "GcpNfsVolumeRestore", "cloud-resources.kyma-project.io"), types.FeatureNfsBackup},
				{"CRD Typed SKR GcpNfsVolume", skrScheme, crdTypedWithKindGroup(t, baseCrdTyped, "GcpNfsVolume", "cloud-resources.kyma-project.io"), types.FeatureNfs},
				{"CRD Typed SKR IpRange", skrScheme, crdTypedWithKindGroup(t, baseCrdTyped, "IpRange", "cloud-resources.kyma-project.io"), ""},
			}

			for _, info := range objList {
				t.Run(info.title, func(t *testing.T) {
					actual := objectToFeaturePredetermined(info.obj, info.scheme)
					assert.Equal(t, info.expected, actual)
				})
			}
		})

		t.Run("From Object Meta", func(t *testing.T) {
			baseCrdUnstructured := &unstructured.Unstructured{Object: map[string]interface{}{}}
			baseCrdUnstructured.SetAPIVersion("apiextensions.k8s.io/v1")
			baseCrdUnstructured.SetKind("CustomResourceDefinition")

			baseCrdTyped := &apiextensions.CustomResourceDefinition{
				TypeMeta: metav1.TypeMeta{
					Kind:       "CustomResourceDefinition",
					APIVersion: "apiextensions.k8s.io/v1",
				},
			}

			emptyScheme := runtime.NewScheme()

			objList := []struct {
				title    string
				scheme   *runtime.Scheme
				obj      client.Object
				expected types.FeatureName
			}{
				{"CRD Typed SKR AwsNfsVolumeBackup", emptyScheme, crdTypedWithKindGroup(t, baseCrdTyped, "AwsNfsVolumeBackup", "cloud-resources.kyma-project.io"), types.FeatureNfsBackup},
				{"CRD Typed SKR AwsNfsVolume", emptyScheme, crdTypedWithKindGroup(t, baseCrdTyped, "AwsNfsVolume", "cloud-resources.kyma-project.io"), types.FeatureNfs},
				{"CRD Typed SKR CloudResources", emptyScheme, crdTypedWithKindGroup(t, baseCrdTyped, "CloudResources", "cloud-resources.kyma-project.io"), ""},
				{"CRD Typed SKR GcpNfsVolumeBackup", emptyScheme, crdTypedWithKindGroup(t, baseCrdTyped, "GcpNfsVolumeBackup", "cloud-resources.kyma-project.io"), types.FeatureNfsBackup},
				{"CRD Typed SKR GcpNfsVolumeRestore", emptyScheme, crdTypedWithKindGroup(t, baseCrdTyped, "GcpNfsVolumeRestore", "cloud-resources.kyma-project.io"), types.FeatureNfsBackup},
				{"CRD Typed SKR GcpNfsVolume", emptyScheme, crdTypedWithKindGroup(t, baseCrdTyped, "GcpNfsVolume", "cloud-resources.kyma-project.io"), types.FeatureNfs},
				{"CRD Typed SKR IpRange", emptyScheme, crdTypedWithKindGroup(t, baseCrdTyped, "IpRange", "cloud-resources.kyma-project.io"), ""},

				{"CRD Unstructured SKR AwsNfsVolumeBackup", emptyScheme, crdUnstructuredWithKindGroup(t, baseCrdUnstructured, "AwsNfsVolumeBackup", "cloud-resources.kyma-project.io"), types.FeatureNfsBackup},
				{"CRD Unstructured SKR AwsNfsVolume", emptyScheme, crdUnstructuredWithKindGroup(t, baseCrdUnstructured, "AwsNfsVolume", "cloud-resources.kyma-project.io"), types.FeatureNfs},
				{"CRD Unstructured SKR CloudResources", emptyScheme, crdUnstructuredWithKindGroup(t, baseCrdUnstructured, "CloudResources", "cloud-resources.kyma-project.io"), ""},
				{"CRD Unstructured SKR GcpNfsVolumeBackup", emptyScheme, crdUnstructuredWithKindGroup(t, baseCrdUnstructured, "GcpNfsVolumeBackup", "cloud-resources.kyma-project.io"), types.FeatureNfsBackup},
				{"CRD Unstructured SKR GcpNfsVolumeRestore", emptyScheme, crdUnstructuredWithKindGroup(t, baseCrdUnstructured, "GcpNfsVolumeRestore", "cloud-resources.kyma-project.io"), types.FeatureNfsBackup},
				{"CRD Unstructured SKR GcpNfsVolume", emptyScheme, crdUnstructuredWithKindGroup(t, baseCrdUnstructured, "GcpNfsVolume", "cloud-resources.kyma-project.io"), types.FeatureNfs},
				{"CRD Unstructured SKR IpRange", emptyScheme, crdUnstructuredWithKindGroup(t, baseCrdUnstructured, "IpRange", "cloud-resources.kyma-project.io"), ""},
			}

			for _, info := range objList {
				t.Run(info.title, func(t *testing.T) {
					actual := objectToFeaturePredetermined(info.obj, info.scheme)
					assert.Equal(t, info.expected, actual)
				})
			}
		})

		t.Run("Busola UI", func(t *testing.T) {

			objList := []struct {
				title    string
				obj      client.Object
				expected types.FeatureName
			}{
				{"Busola Typed AwsNfsVolumeBackup", busolaCmTypedKindGroup(t, "AwsNfsVolumeBackup"), types.FeatureNfsBackup},
				{"Busola Typed AwsNfsVolume", busolaCmTypedKindGroup(t, "AwsNfsVolume"), types.FeatureNfs},
				{"Busola Typed CloudResources", busolaCmTypedKindGroup(t, "CloudResources"), ""},
				{"Busola Typed GcpNfsVolumeBackup", busolaCmTypedKindGroup(t, "GcpNfsVolumeBackup"), types.FeatureNfsBackup},
				{"Busola Typed GcpNfsVolumeRestore", busolaCmTypedKindGroup(t, "GcpNfsVolumeRestore"), types.FeatureNfsBackup},
				{"Busola Typed GcpNfsVolume", busolaCmTypedKindGroup(t, "GcpNfsVolume"), types.FeatureNfs},
				{"Busola Typed IpRange", busolaCmTypedKindGroup(t, "IpRange"), ""},

				{"Busola Unstructured AwsNfsVolumeBackup", busolaCmUnstructuredKindGroup(t, "AwsNfsVolumeBackup"), types.FeatureNfsBackup},
				{"Busola Unstructured AwsNfsVolume", busolaCmUnstructuredKindGroup(t, "AwsNfsVolume"), types.FeatureNfs},
				{"Busola Unstructured CloudResources", busolaCmUnstructuredKindGroup(t, "CloudResources"), ""},
				{"Busola Unstructured GcpNfsVolumeBackup", busolaCmUnstructuredKindGroup(t, "GcpNfsVolumeBackup"), types.FeatureNfsBackup},
				{"Busola Unstructured GcpNfsVolumeRestore", busolaCmUnstructuredKindGroup(t, "GcpNfsVolumeRestore"), types.FeatureNfsBackup},
				{"Busola Unstructured GcpNfsVolume", busolaCmUnstructuredKindGroup(t, "GcpNfsVolume"), types.FeatureNfs},
				{"Busola Unstructured IpRange", busolaCmUnstructuredKindGroup(t, "IpRange"), ""},
			}

			emptyScheme := runtime.NewScheme()

			for _, info := range objList {
				t.Run(info.title, func(t *testing.T) {
					actual := objectToFeaturePredetermined(info.obj, emptyScheme)
					assert.Equal(t, info.expected, actual)
				})
			}
		})
	})
}

func crdUnstructuredWithKindGroup(t *testing.T, baseCrd *unstructured.Unstructured, k, g string) *unstructured.Unstructured {
	crd := baseCrd.DeepCopyObject().(*unstructured.Unstructured)
	err := unstructured.SetNestedField(crd.Object, k, "spec", "names", "kind")
	assert.NoError(t, err)
	err = unstructured.SetNestedField(crd.Object, g, "spec", "group")
	assert.NoError(t, err)
	return crd
}

func crdTypedWithKindGroup(t *testing.T, baseCrd *apiextensions.CustomResourceDefinition, k, g string) *apiextensions.CustomResourceDefinition {
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
