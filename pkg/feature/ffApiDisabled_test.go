package feature

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
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

	err := Initialize(ctx, WithFile("testdata/apiDisabled.yaml"), WithEvaluateAllFlagsState())
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
		{
			t: "cloudresources is enabled regardless of landscape and feature",
			c: ContextBuilderFromCtx(ctx).KindsFromObject(&cloudresourcesv1beta1.CloudResources{}, sch).Build(ctx),
			v: false,
		},
		// NFS ====================================================
		{
			t: "nfs feature is enabled on dev",
			c: ContextBuilderFromCtx(ctx).Landscape(LandscapeDev).Feature(FeatureNfs).Build(ctx),
			v: false,
		},
		{
			t: "nfs feature is enabled on stage",
			c: ContextBuilderFromCtx(ctx).Landscape(LandscapeStage).Feature(FeatureNfs).Build(ctx),
			v: false,
		},
		{
			t: "nfs feature is enabled on prod",
			c: ContextBuilderFromCtx(ctx).Landscape(LandscapeProd).Feature(FeatureNfs).BrokerPlan("aws").Build(ctx),
			v: false,
		},
		// NFS BACKUP ====================================================
		{
			t: "nfsBackup feature is enabled on dev",
			c: ContextBuilderFromCtx(ctx).Landscape(LandscapeDev).Feature(FeatureNfsBackup).Build(ctx),
			v: false,
		},
		{
			t: "nfsBackup feature is disabled on stage",
			c: ContextBuilderFromCtx(ctx).Landscape(LandscapeStage).Feature(FeatureNfsBackup).Build(ctx),
			v: true,
		},
		{
			t: "nfsBackup feature is disabled on prod",
			c: ContextBuilderFromCtx(ctx).Landscape(LandscapeProd).Feature(FeatureNfsBackup).Build(ctx),
			v: true,
		},
		// PEERING ====================================================
		{
			t: "peering feature is enabled on dev",
			c: ContextBuilderFromCtx(ctx).Landscape(LandscapeDev).Feature(FeaturePeering).Build(ctx),
			v: false,
		},
		{
			t: "peering feature is disabled on stage",
			c: ContextBuilderFromCtx(ctx).Landscape(LandscapeStage).Feature(FeaturePeering).Build(ctx),
			v: true,
		},
		{
			t: "peering feature is disabled on prod",
			c: ContextBuilderFromCtx(ctx).Landscape(LandscapeProd).Feature(FeaturePeering).Build(ctx),
			v: true,
		},
		// TRIAL ====================================================
		{
			t: "nfs feature is disabled on trial",
			c: ContextBuilderFromCtx(ctx).BrokerPlan("trial").Feature(FeatureNfs).Build(ctx),
			v: true,
		},
		{
			t: "nfsbackup feature is disabled on trial",
			c: ContextBuilderFromCtx(ctx).BrokerPlan("trial").Feature(FeatureNfsBackup).Build(ctx),
			v: true,
		},
		{
			t: "peering feature is disabled on trial",
			c: ContextBuilderFromCtx(ctx).BrokerPlan("trial").Feature(FeaturePeering).Build(ctx),
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
