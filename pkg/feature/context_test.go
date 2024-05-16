package feature

import (
	"context"
	"github.com/onsi/gomega"
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

	t.Run("Object method", func(t *testing.T) {

	})
}
