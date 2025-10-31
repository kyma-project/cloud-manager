package feature

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/pkg/common/bootstrap"

	"testing"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/feature/types"
	"github.com/stretchr/testify/assert"
)

func TestApiDisabled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := Initialize(ctx, logr.Discard(), WithFile("testdata/apiDisabled.yaml"))
	assert.NoError(t, err)

	sch := bootstrap.SkrScheme

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
		// NFS BACKUP DISCOVERY ====================================================
		{
			t: "nfsBackupDiscovery feature is enabled on dev",
			c: ContextBuilderFromCtx(ctx).Landscape(types.LandscapeDev).Feature(types.FeatureNfsBackupDiscovery).Build(ctx),
			v: false,
		},
		{
			t: "nfsBackupDiscovery feature is disabled on stage",
			c: ContextBuilderFromCtx(ctx).Landscape(types.LandscapeStage).Feature(types.FeatureNfsBackupDiscovery).Build(ctx),
			v: true,
		},
		{
			t: "nfsBackupDiscovery feature is disabled on prod",
			c: ContextBuilderFromCtx(ctx).Landscape(types.LandscapeProd).Feature(types.FeatureNfsBackupDiscovery).Build(ctx),
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
		// REDIS ====================================================
		{
			t: "redis feature is enabled on dev",
			c: ContextBuilderFromCtx(ctx).Landscape(types.LandscapeDev).Feature(types.FeatureRedis).Build(ctx),
			v: false,
		},
		{
			t: "redis feature is disabled on stage",
			c: ContextBuilderFromCtx(ctx).Landscape(types.LandscapeStage).Feature(types.FeatureRedis).Build(ctx),
			v: true,
		},
		{
			t: "redis feature is disabled on prod",
			c: ContextBuilderFromCtx(ctx).Landscape(types.LandscapeProd).Feature(types.FeatureRedis).Build(ctx),
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
		{
			t: "redis feature is disabled on trial",
			c: ContextBuilderFromCtx(ctx).BrokerPlan("trial").Feature(types.FeatureRedis).Build(ctx),
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
