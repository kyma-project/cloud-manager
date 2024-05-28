package feature

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/feature/types"
	"github.com/stretchr/testify/assert"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"testing"
)

func TestApiDisabled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := Initialize(ctx, WithFile("testdata/apiDisabled.yaml"))
	assert.NoError(t, err)

	sch := runtime.NewScheme()
	utilruntime.Must(scheme.AddToScheme(sch))
	utilruntime.Must(cloudresourcesv1beta1.AddToScheme(sch))
	utilruntime.Must(apiextensions.AddToScheme(sch))

	cases := []struct {
		t string
		c context.Context
		v bool
	}{
		// CloudResources
		{
			t: "cloudresources is enabled regardless of landscape and feature",
			c: ContextBuilderFromCtx(ctx).KindsFromObject(&cloudresourcesv1beta1.CloudResources{}, sch).Build(ctx),
			v: false,
		},
		{
			t: "cloudresources is enabled on dev",
			c: ContextBuilderFromCtx(ctx).Landscape("dev").KindsFromObject(&cloudresourcesv1beta1.CloudResources{}, sch).Build(ctx),
			v: false,
		},
		{
			t: "cloudresources is enabled on stage",
			c: ContextBuilderFromCtx(ctx).Landscape("stage").KindsFromObject(&cloudresourcesv1beta1.CloudResources{}, sch).Build(ctx),
			v: false,
		},
		{
			t: "cloudresources is enabled on prod",
			c: ContextBuilderFromCtx(ctx).Landscape("prod").KindsFromObject(&cloudresourcesv1beta1.CloudResources{}, sch).Build(ctx),
			v: false,
		},
		{
			t: "cloudresources is enabled on prod even for disabled feature",
			c: ContextBuilderFromCtx(ctx).Landscape("prod").Feature(types.FeatureNfsBackup).KindsFromObject(&cloudresourcesv1beta1.CloudResources{}, sch).Build(ctx),
			v: false,
		},
		{
			t: "cloudresources is enabled on trial",
			c: ContextBuilderFromCtx(ctx).BrokerPlan("trial").Landscape("prod").Feature(types.FeatureNfsBackup).KindsFromObject(&cloudresourcesv1beta1.CloudResources{}, sch).Build(ctx),
			v: false,
		},
		// NFS ====================================================
		{
			t: "nfs feature is enabled on dev",
			c: ContextBuilderFromCtx(ctx).Landscape(types.LandscapeDev).Feature(types.FeatureNfs).Build(ctx),
			v: false,
		},
		{
			t: "nfs feature is enabled on stage",
			c: ContextBuilderFromCtx(ctx).Landscape(types.LandscapeStage).Feature(types.FeatureNfs).Build(ctx),
			v: false,
		},
		{
			t: "nfs feature is enabled on prod",
			c: ContextBuilderFromCtx(ctx).Landscape(types.LandscapeProd).Feature(types.FeatureNfs).BrokerPlan("aws").Build(ctx),
			v: false,
		},
		// NFS BACKUP ====================================================
		{
			t: "nfsBackup feature is enabled on dev",
			c: ContextBuilderFromCtx(ctx).Landscape(types.LandscapeDev).Feature(types.FeatureNfsBackup).Build(ctx),
			v: false,
		},
		{
			t: "nfsBackup feature is disabled on stage",
			c: ContextBuilderFromCtx(ctx).Landscape(types.LandscapeStage).Feature(types.FeatureNfsBackup).Build(ctx),
			v: true,
		},
		{
			t: "nfsBackup feature is disabled on prod",
			c: ContextBuilderFromCtx(ctx).Landscape(types.LandscapeProd).Feature(types.FeatureNfsBackup).Build(ctx),
			v: true,
		},
		// PEERING ====================================================
		{
			t: "peering feature is enabled on dev",
			c: ContextBuilderFromCtx(ctx).Landscape(types.LandscapeDev).Feature(types.FeaturePeering).Build(ctx),
			v: false,
		},
		{
			t: "peering feature is disabled on stage",
			c: ContextBuilderFromCtx(ctx).Landscape(types.LandscapeStage).Feature(types.FeaturePeering).Build(ctx),
			v: true,
		},
		{
			t: "peering feature is disabled on prod",
			c: ContextBuilderFromCtx(ctx).Landscape(types.LandscapeProd).Feature(types.FeaturePeering).Build(ctx),
			v: true,
		},
		// TRIAL ====================================================
		{
			t: "nfs feature is disabled on trial",
			c: ContextBuilderFromCtx(ctx).BrokerPlan("trial").Feature(types.FeatureNfs).Build(ctx),
			v: true,
		},
		{
			t: "nfsbackup feature is disabled on trial",
			c: ContextBuilderFromCtx(ctx).BrokerPlan("trial").Feature(types.FeatureNfsBackup).Build(ctx),
			v: true,
		},
		{
			t: "peering feature is disabled on trial",
			c: ContextBuilderFromCtx(ctx).BrokerPlan("trial").Feature(types.FeaturePeering).Build(ctx),
			v: true,
		},
	}

	for _, cs := range cases {
		t.Run(cs.t, func(t *testing.T) {
			actual := ApiDisabled.Value(cs.c)
			assert.Equal(t, cs.v, actual)
		})
	}

}
