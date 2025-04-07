package common

import (
	"fmt"
	"github.com/elliotchance/pie/v2"
	cloudcontrol1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/objkind"
	"github.com/stretchr/testify/assert"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	//"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	//apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
)

func TestObjSupportsProvider(t *testing.T) {

	const (
		AWS       = cloudcontrol1beta1.ProviderAws
		AZURE     = cloudcontrol1beta1.ProviderAzure
		GCP       = cloudcontrol1beta1.ProviderGCP
		OPENSTACK = cloudcontrol1beta1.ProviderOpenStack
	)
	scheme := runtime.NewScheme()
	assert.NoError(t, cloudresourcesv1beta1.AddToScheme(scheme))
	assert.NoError(t, clientgoscheme.AddToScheme(scheme))

	g := cloudresourcesv1beta1.GroupVersion.Group
	//v := cloudresourcesv1beta1.GroupVersion.Version

	kCrd := "CustomResourceDefinition"

	baseCrdUnstructured := &unstructured.Unstructured{Object: map[string]interface{}{}}
	baseCrdUnstructured.SetAPIVersion(apiextensionsv1.SchemeGroupVersion.WithKind(kCrd).GroupVersion().String())
	baseCrdUnstructured.SetKind(kCrd)

	//gCm := ""
	//kCm := "ConfigMap"

	testData := []struct {
		title              string
		obj                client.Object
		supportedProviders []cloudcontrol1beta1.ProviderType
	}{
		//{
		//	"IpRange typed",
		//	&cloudresourcesv1beta1.IpRange{},
		//	[]cloudcontrol1beta1.ProviderType{AWS, AZURE, GCP},
		//},
		//{
		//	"IpRange unstructured",
		//	objkind.NewUnstructuredWithGVK(g, v, "IpRange"),
		//	[]cloudcontrol1beta1.ProviderType{AWS, AZURE, GCP},
		//},
		{
			"IpRange CRD typed",
			objkind.NewCrdTypedV1WithKindGroup(t, "IpRange", g),
			[]cloudcontrol1beta1.ProviderType{AWS, AZURE, GCP},
		},
		//{
		//	"IpRange CRD unstructured",
		//	objkind.NewCrdUnstructuredWithKindGroup(t, baseCrdUnstructured, "IpRange", g),
		//	[]cloudcontrol1beta1.ProviderType{AWS, AZURE, GCP},
		//},
	}

	allProviders := []cloudcontrol1beta1.ProviderType{AWS, AZURE, GCP, OPENSTACK}

	for _, data := range testData {
		for _, provider := range allProviders {
			t.Run(fmt.Sprintf("%s %s", data.title, provider), func(t *testing.T) {
				isSupported := pie.Contains(data.supportedProviders, provider)
				actual := ObjSupportsProvider(data.obj, scheme, string(provider))
				assert.Equal(t, isSupported, actual, data.title)
			})
		}

	}
}
