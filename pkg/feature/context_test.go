package feature

import (
	"context"
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/objkind"
	"github.com/kyma-project/cloud-manager/pkg/feature/types"
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
			Feature(types.FeatureNfs).
			Build(ctx)
		ffCtxAttr := ContextFromCtx(ctx).GetCustom()
		gomega.Expect(ffCtxAttr).To(gomega.HaveKeyWithValue(types.KeyLandscape, "stage"))
		gomega.Expect(ffCtxAttr).To(gomega.HaveKeyWithValue(types.KeyFeature, types.FeatureNfs))

		ctx = ContextBuilderFromCtx(ctx).
			Feature(string(types.FeatureNfsBackup)).
			Plane(types.PlaneKcp).
			Provider("aws").
			Build(ctx)
		ffCtxAttr = ContextFromCtx(ctx).GetCustom()
		gomega.Expect(ffCtxAttr).To(gomega.HaveKeyWithValue(types.KeyFeature, types.FeatureNfsBackup))
		gomega.Expect(ffCtxAttr).To(gomega.HaveKeyWithValue(types.KeyPlane, types.PlaneKcp))
		gomega.Expect(ffCtxAttr).To(gomega.HaveKeyWithValue(types.KeyProvider, "aws"))

		ctx = ContextBuilderFromCtx(ctx).
			Feature(string(types.FeaturePeering)).
			Plane(types.PlaneSkr).
			Provider("gcp").
			BrokerPlan("trial").
			GlobalAccount("glob-123").
			SubAccount("sub-456").
			Kyma("kyma-789").
			Shoot("shoot-34567").
			ObjKindGroup("vpcpeering.cloud-control.kyma-project.io").
			Custom("foo", "bar").
			Build(ctx)
		ffCtxAttr = ContextFromCtx(ctx).GetCustom()
		gomega.Expect(ffCtxAttr).To(gomega.HaveKeyWithValue(types.KeyFeature, string(types.FeaturePeering)))
		gomega.Expect(ffCtxAttr).To(gomega.HaveKeyWithValue(types.KeyPlane, types.PlaneSkr))
		gomega.Expect(ffCtxAttr).To(gomega.HaveKeyWithValue(types.KeyProvider, "gcp"))
		gomega.Expect(ffCtxAttr).To(gomega.HaveKeyWithValue(types.KeyBrokerPlan, "trial"))
		gomega.Expect(ffCtxAttr).To(gomega.HaveKeyWithValue(types.KeyGlobalAccount, "glob-123"))
		gomega.Expect(ffCtxAttr).To(gomega.HaveKeyWithValue(types.KeySubAccount, "sub-456"))
		gomega.Expect(ffCtxAttr).To(gomega.HaveKeyWithValue(types.KeyKyma, "kyma-789"))
		gomega.Expect(ffCtxAttr).To(gomega.HaveKeyWithValue(types.KeyShoot, "shoot-34567"))
		gomega.Expect(ffCtxAttr).To(gomega.HaveKeyWithValue(types.KeyObjKindGroup, "vpcpeering.cloud-control.kyma-project.io"))
		gomega.Expect(ffCtxAttr).To(gomega.HaveKeyWithValue("foo", "bar"))
	})

	t.Run("KindsFromObject method", func(t *testing.T) {

		sch := runtime.NewScheme()
		utilruntime.Must(scheme.AddToScheme(sch))
		utilruntime.Must(cloudcontrolv1beta1.AddToScheme(sch))
		utilruntime.Must(cloudresourcesv1beta1.AddToScheme(sch))
		utilruntime.Must(apiextensions.AddToScheme(sch))

		t.Run("All KindGroups", func(t *testing.T) {
			objList := []struct {
				title    string
				obj      client.Object
				kg       string
				crdKg    string
				busolaKg string
			}{
				{"KCP IpRange", &cloudcontrolv1beta1.IpRange{}, "iprange.cloud-control.kyma-project.io", "", ""},
				{"SKR AwsNfsVolume", &cloudresourcesv1beta1.AwsNfsVolume{}, "awsnfsvolume.cloud-resources.kyma-project.io", "", ""},
				{"CRD KCP NfsInstance", objkind.NewCrdTypedV1WithKindGroup(t, "NfsInstance", "cloud-control.kyma-project.io"), "customresourcedefinition.apiextensions.k8s.io", "nfsinstance.cloud-control.kyma-project.io", ""},
				{"CRD SKR GcpNfsVolume", objkind.NewCrdTypedV1WithKindGroup(t, "GcpNfsVolume", "cloud-resources.kyma-project.io"), "customresourcedefinition.apiextensions.k8s.io", "gcpnfsvolume.cloud-resources.kyma-project.io", ""},
				{"Busola SKR IpRange", objkind.NewBusolaCmTypedKindGroup(t, "IpRange"), "configmap", "", "iprange.cloud-resources.kyma-project.io"},
			}

			for _, info := range objList {
				t.Run(info.title, func(t *testing.T) {
					ffCtx := ContextBuilderFromCtx(context.Background()).
						KindsFromObject(info.obj, sch).
						FFCtx()
					objKg := ffCtx.GetCustom()[types.KeyObjKindGroup]
					crdKg := ffCtx.GetCustom()[types.KeyCrdKindGroup]
					busolaKg := ffCtx.GetCustom()[types.KeyBusolaKindGroup]
					allKg := ffCtx.GetCustom()[types.KeyAllKindGroups]
					assert.Equal(t, info.kg, objKg)
					assert.Equal(t, info.crdKg, crdKg)
					assert.Equal(t, info.busolaKg, busolaKg)
					expectedAllKg := fmt.Sprintf("%s,%s,%s", objKg, crdKg, busolaKg)
					assert.Equal(t, expectedAllKg, allKg)
				})
			}
		})
	})

	t.Run("LoadFromKyma", func(t *testing.T) {
		kyma := util.NewKymaUnstructured()
		kyma.SetLabels(map[string]string{
			cloudcontrolv1beta1.LabelScopeBrokerPlanName:  "trial",
			cloudcontrolv1beta1.LabelScopeGlobalAccountId: "glob-123",
			cloudcontrolv1beta1.LabelScopeSubaccountId:    "sub-456",
			cloudcontrolv1beta1.LabelScopeRegion:          "us-east-1",
			cloudcontrolv1beta1.LabelScopeShootName:       "shoot-67890",
		})

		ffCtx := ContextBuilderFromCtx(context.Background()).
			LoadFromKyma(kyma).
			FFCtx()

		assert.Equal(t, "trial", ffCtx.GetCustom()[types.KeyBrokerPlan])
		assert.Equal(t, "glob-123", ffCtx.GetCustom()[types.KeyGlobalAccount])
		assert.Equal(t, "sub-456", ffCtx.GetCustom()[types.KeySubAccount])
		assert.Equal(t, "us-east-1", ffCtx.GetCustom()[types.KeyRegion])
		assert.Equal(t, "shoot-67890", ffCtx.GetCustom()[types.KeyShoot])
	})

	t.Run("LoadFromScope", func(t *testing.T) {
		scope := &cloudcontrolv1beta1.Scope{}
		scope.Spec.Provider = "aws"

		ffCtx := ContextBuilderFromCtx(context.Background()).
			LoadFromScope(scope).
			FFCtx()

		assert.Equal(t, "aws", ffCtx.GetCustom()[types.KeyProvider])
	})
}
