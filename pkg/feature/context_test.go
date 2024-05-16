package feature

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
)

func TestContextBuilder(t *testing.T) {
	gomega.RegisterTestingT(t)

	t.Run("Standard fields replace value in consecutive calls", func(t *testing.T) {

		ctx := context.Background()

		ctx = ContextBuilderFromCtx(ctx).
			Landscape("stage").
			Feature(FeatureNfs).
			Build(ctx)
		ffCtxAttr := ContextFromCtx(ctx).GetCustom()
		gomega.Expect(ffCtxAttr).To(gomega.HaveKeyWithValue(KeyLandscape, "stage"))
		gomega.Expect(ffCtxAttr).To(gomega.HaveKeyWithValue(KeyFeature, FeatureNfs))

		ctx = ContextBuilderFromCtx(ctx).
			Feature(string(FeatureNfsBackup)).
			Plane(PlaneKcp).
			Provider("aws").
			Build(ctx)
		ffCtxAttr = ContextFromCtx(ctx).GetCustom()
		gomega.Expect(ffCtxAttr).To(gomega.HaveKeyWithValue(KeyFeature, FeatureNfsBackup))
		gomega.Expect(ffCtxAttr).To(gomega.HaveKeyWithValue(KeyPlane, PlaneKcp))
		gomega.Expect(ffCtxAttr).To(gomega.HaveKeyWithValue(KeyProvider, "aws"))

		ctx = ContextBuilderFromCtx(ctx).
			Feature(string(FeaturePeering)).
			Plane(PlaneSkr).
			Provider("gcp").
			BrokerPlan("trial").
			GlobalAccount("glob-123").
			SubAccount("sub-456").
			Kyma("kyma-789").
			Shoot("shoot-34567").
			KindGroup("vpcpeering.cloud-control.kyma-project.io").
			Custom("foo", "bar").
			Build(ctx)
		ffCtxAttr = ContextFromCtx(ctx).GetCustom()
		gomega.Expect(ffCtxAttr).To(gomega.HaveKeyWithValue(KeyFeature, string(FeaturePeering)))
		gomega.Expect(ffCtxAttr).To(gomega.HaveKeyWithValue(KeyPlane, PlaneSkr))
		gomega.Expect(ffCtxAttr).To(gomega.HaveKeyWithValue(KeyProvider, "gcp"))
		gomega.Expect(ffCtxAttr).To(gomega.HaveKeyWithValue(KeyBrokerPlan, "trial"))
		gomega.Expect(ffCtxAttr).To(gomega.HaveKeyWithValue(KeyGlobalAccount, "glob-123"))
		gomega.Expect(ffCtxAttr).To(gomega.HaveKeyWithValue(KeySubAccount, "sub-456"))
		gomega.Expect(ffCtxAttr).To(gomega.HaveKeyWithValue(KeyKyma, "kyma-789"))
		gomega.Expect(ffCtxAttr).To(gomega.HaveKeyWithValue(KeyShoot, "shoot-34567"))
		gomega.Expect(ffCtxAttr).To(gomega.HaveKeyWithValue(KeyKindGroup, "vpcpeering.cloud-control.kyma-project.io"))
		gomega.Expect(ffCtxAttr).To(gomega.HaveKeyWithValue("foo", "bar"))
	})

	t.Run("KindsFromObject method", func(t *testing.T) {

		sch := runtime.NewScheme()
		utilruntime.Must(scheme.AddToScheme(sch))
		utilruntime.Must(cloudcontrolv1beta1.AddToScheme(sch))
		utilruntime.Must(cloudresourcesv1beta1.AddToScheme(sch))
		utilruntime.Must(apiextensions.AddToScheme(sch))

		t.Run("KindGroup", func(t *testing.T) {
			baseCrdTyped := &apiextensions.CustomResourceDefinition{}
			objList := []struct {
				title    string
				obj      client.Object
				kg       string
				crdKg    string
				busolaKg string
			}{
				{"KCP IpRange", &cloudcontrolv1beta1.IpRange{}, "iprange.cloud-control.kyma-project.io", "", ""},
				{"SKR AwsNfsVolume", &cloudresourcesv1beta1.AwsNfsVolume{}, "awsnfsvolume.cloud-resources.kyma-project.io", "", ""},
				{"CRD KCP NfsInstance", crdTypedWithKindGroup(t, baseCrdTyped, "NfsInstance", "cloud-control.kyma-project.io"), "customresourcedefinition.apiextensions.k8s.io", "nfsinstance.cloud-control.kyma-project.io", ""},
				{"CRD SKR GcpNfsVolume", crdTypedWithKindGroup(t, baseCrdTyped, "GcpNfsVolume", "cloud-resources.kyma-project.io"), "customresourcedefinition.apiextensions.k8s.io", "gcpnfsvolume.cloud-resources.kyma-project.io", ""},
				{"Busola SKR IpRange", busolaCmTypedKindGroup(t, "IpRange"), "configmap", "", "iprange.cloud-resources.kyma-project.io"},
			}

			for _, info := range objList {
				ffCtx := ContextBuilderFromCtx(context.Background()).
					KindsFromObject(info.obj, sch).
					FFCtx()
				kg := ffCtx.GetCustom()["kindGroup"]
				crdKg := ffCtx.GetCustom()["crdKindGroup"]
				busolaKg := ffCtx.GetCustom()["busolaKindGroup"]
				assert.Equal(t, info.kg, kg)
				assert.Equal(t, info.crdKg, crdKg)
				assert.Equal(t, info.busolaKg, busolaKg)
			}
		})
	})

	t.Run("LoadFromKyma", func(t *testing.T) {
		kyma := util.NewKymaUnstructured()
		kyma.SetLabels(map[string]string{
			"kyma-project.io/broker-plan-name":  "trial",
			"kyma-project.io/global-account-id": "glob-123",
			"kyma-project.io/subaccount-id":     "sub-456",
			"kyma-project.io/region":            "us-east-1",
			"kyma-project.io/shoot-name":        "shoot-67890",
		})

		ffCtx := ContextBuilderFromCtx(context.Background()).
			LoadFromKyma(kyma).
			FFCtx()

		assert.Equal(t, "trial", ffCtx.GetCustom()[KeyBrokerPlan])
		assert.Equal(t, "glob-123", ffCtx.GetCustom()[KeyGlobalAccount])
		assert.Equal(t, "sub-456", ffCtx.GetCustom()[KeySubAccount])
		assert.Equal(t, "us-east-1", ffCtx.GetCustom()[KeyRegion])
		assert.Equal(t, "shoot-67890", ffCtx.GetCustom()[KeyShoot])
	})

	t.Run("LoadFromScope", func(t *testing.T) {
		scope := &cloudcontrolv1beta1.Scope{}
		scope.Spec.Provider = "aws"

		ffCtx := ContextBuilderFromCtx(context.Background()).
			LoadFromScope(scope).
			FFCtx()

		assert.Equal(t, "aws", ffCtx.GetCustom()[KeyProvider])
	})
}
