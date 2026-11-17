package feature

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	commonscheme "github.com/kyma-project/cloud-manager/pkg/common/scheme"
	"github.com/kyma-project/cloud-manager/pkg/feature/types"
	"github.com/stretchr/testify/assert"
)

// TestApiEnabled asserts ApiEnabled.Value == !ApiDisabled.Value for every case
// in the parity matrix. Both flags are loaded from testdata/apiEnabled.yaml
// which carries mirrored blocks (same rule order, inverted variations).
func TestApiEnabled(t *testing.T) {
	ctx := t.Context()

	err := Initialize(ctx, logr.Discard(), WithFile("testdata/apiEnabled.yaml"))
	assert.NoError(t, err)

	sch := commonscheme.SkrScheme

	cases := []struct {
		t string
		c context.Context
	}{
		// CloudResources — bootstrap kind-group must always be enabled
		{
			t: "cloudresources is enabled regardless of landscape and feature",
			c: ContextBuilderFromCtx(ctx).KindsFromObject(&cloudresourcesv1beta1.CloudResources{}, sch).Build(ctx),
		},
		{
			t: "cloudresources is enabled on dev",
			c: ContextBuilderFromCtx(ctx).Landscape("dev").KindsFromObject(&cloudresourcesv1beta1.CloudResources{}, sch).Build(ctx),
		},
		{
			t: "cloudresources is enabled on stage",
			c: ContextBuilderFromCtx(ctx).Landscape("stage").KindsFromObject(&cloudresourcesv1beta1.CloudResources{}, sch).Build(ctx),
		},
		{
			t: "cloudresources is enabled on prod",
			c: ContextBuilderFromCtx(ctx).Landscape("prod").KindsFromObject(&cloudresourcesv1beta1.CloudResources{}, sch).Build(ctx),
		},
		{
			t: "cloudresources is enabled on prod even for disabled feature",
			c: ContextBuilderFromCtx(ctx).Landscape("prod").Feature(types.FeatureNfsBackup).KindsFromObject(&cloudresourcesv1beta1.CloudResources{}, sch).Build(ctx),
		},
		{
			t: "cloudresources is enabled on trial",
			c: ContextBuilderFromCtx(ctx).BrokerPlan("trial").Landscape("prod").Feature(types.FeatureNfsBackup).KindsFromObject(&cloudresourcesv1beta1.CloudResources{}, sch).Build(ctx),
		},
		// NFS ====================================================
		{
			t: "nfs feature is enabled on dev",
			c: ContextBuilderFromCtx(ctx).Landscape(types.LandscapeDev).Feature(types.FeatureNfs).Build(ctx),
		},
		{
			t: "nfs feature is enabled on stage",
			c: ContextBuilderFromCtx(ctx).Landscape(types.LandscapeStage).Feature(types.FeatureNfs).Build(ctx),
		},
		{
			t: "nfs feature is enabled on prod",
			c: ContextBuilderFromCtx(ctx).Landscape(types.LandscapeProd).Feature(types.FeatureNfs).BrokerPlan("aws").Build(ctx),
		},
		// NFS BACKUP ====================================================
		{
			t: "nfsBackup feature is enabled on dev",
			c: ContextBuilderFromCtx(ctx).Landscape(types.LandscapeDev).Feature(types.FeatureNfsBackup).Build(ctx),
		},
		{
			t: "nfsBackup feature is disabled on stage",
			c: ContextBuilderFromCtx(ctx).Landscape(types.LandscapeStage).Feature(types.FeatureNfsBackup).Build(ctx),
		},
		{
			t: "nfsBackup feature is disabled on prod",
			c: ContextBuilderFromCtx(ctx).Landscape(types.LandscapeProd).Feature(types.FeatureNfsBackup).Build(ctx),
		},
		// PEERING ====================================================
		{
			t: "peering feature is enabled on dev",
			c: ContextBuilderFromCtx(ctx).Landscape(types.LandscapeDev).Feature(types.FeaturePeering).Build(ctx),
		},
		{
			t: "peering feature is disabled on stage",
			c: ContextBuilderFromCtx(ctx).Landscape(types.LandscapeStage).Feature(types.FeaturePeering).Build(ctx),
		},
		{
			t: "peering feature is disabled on prod",
			c: ContextBuilderFromCtx(ctx).Landscape(types.LandscapeProd).Feature(types.FeaturePeering).Build(ctx),
		},
		// REDIS ====================================================
		{
			t: "redis feature is enabled on dev",
			c: ContextBuilderFromCtx(ctx).Landscape(types.LandscapeDev).Feature(types.FeatureRedis).Build(ctx),
		},
		{
			t: "redis feature is disabled on stage",
			c: ContextBuilderFromCtx(ctx).Landscape(types.LandscapeStage).Feature(types.FeatureRedis).Build(ctx),
		},
		{
			t: "redis feature is disabled on prod",
			c: ContextBuilderFromCtx(ctx).Landscape(types.LandscapeProd).Feature(types.FeatureRedis).Build(ctx),
		},
		// TRIAL ====================================================
		{
			t: "nfs feature is disabled on trial",
			c: ContextBuilderFromCtx(ctx).BrokerPlan("trial").Feature(types.FeatureNfs).Build(ctx),
		},
		{
			t: "nfsbackup feature is disabled on trial",
			c: ContextBuilderFromCtx(ctx).BrokerPlan("trial").Feature(types.FeatureNfsBackup).Build(ctx),
		},
		{
			t: "peering feature is disabled on trial",
			c: ContextBuilderFromCtx(ctx).BrokerPlan("trial").Feature(types.FeaturePeering).Build(ctx),
		},
		{
			t: "redis feature is disabled on trial",
			c: ContextBuilderFromCtx(ctx).BrokerPlan("trial").Feature(types.FeatureRedis).Build(ctx),
		},
	}

	for _, cs := range cases {
		t.Run(cs.t, func(t *testing.T) {
			assert.Equal(t, !ApiDisabled.Value(cs.c), ApiEnabled.Value(cs.c),
				"ApiEnabled must equal !ApiDisabled for case: %s", cs.t)
		})
	}
}

// TestApiEnabledEmbeddedGAParity asserts ApiEnabled.Value == !ApiDisabled.Value for every
// GA feature case when both flags are loaded from the embedded ff_ga.yaml (WithEmbeddedRetriever).
// This is the oracle that verifies the apiEnabled allowlist in ff_ga.yaml reproduces
// the same enabled-set as the existing apiDisabled disable-by-exception list.
func TestApiEnabledEmbeddedGAParity(t *testing.T) {
	ctx := t.Context()

	err := Initialize(ctx, logr.Discard(), WithEmbeddedRetriever())
	assert.NoError(t, err)

	cases := []struct {
		name string
		c    context.Context
	}{
		// GA-enabled features
		{
			name: "peering is enabled",
			c:    ContextBuilderFromCtx(ctx).Feature(types.FeaturePeering).Build(ctx),
		},
		{
			name: "nfs is enabled",
			c:    ContextBuilderFromCtx(ctx).Feature(types.FeatureNfs).Build(ctx),
		},
		{
			name: "redis is enabled",
			c:    ContextBuilderFromCtx(ctx).Feature(types.FeatureRedis).Build(ctx),
		},
		// Disable exceptions within enabled set
		{
			name: "nfs on openstack is disabled",
			c:    ContextBuilderFromCtx(ctx).Feature(types.FeatureNfs).Provider("openstack").Build(ctx),
		},
		{
			name: "redis on alicloud is disabled",
			c:    ContextBuilderFromCtx(ctx).Feature(types.FeatureRedis).Provider("alicloud").Build(ctx),
		},
		// Features fully disabled in GA
		{
			name: "nfsBackup is disabled",
			c:    ContextBuilderFromCtx(ctx).Feature(types.FeatureNfsBackup).Build(ctx),
		},
		{
			name: "rediscluster is disabled",
			c:    ContextBuilderFromCtx(ctx).Feature(types.FeatureRedisCluster).Build(ctx),
		},
		{
			name: "rediscluster on alicloud is disabled",
			c:    ContextBuilderFromCtx(ctx).Feature(types.FeatureRedisCluster).Provider("alicloud").Build(ctx),
		},
		{
			name: "vpcdnslink is disabled",
			c:    ContextBuilderFromCtx(ctx).Feature(types.FeatureVpcDnsLink).Build(ctx),
		},
		{
			name: "waf is disabled",
			c:    ContextBuilderFromCtx(ctx).Feature(types.FeatureWAF).Build(ctx),
		},
		{
			name: "certificate is disabled",
			c:    ContextBuilderFromCtx(ctx).Feature(types.FeatureCertificate).Build(ctx),
		},
		{
			name: "azureManagedRedis is disabled",
			c:    ContextBuilderFromCtx(ctx).Feature(types.FeatureAzureManagedRedis).Build(ctx),
		},
		// Trial disables everything
		{
			name: "trial disables all (no feature)",
			c:    ContextBuilderFromCtx(ctx).BrokerPlan("trial").Build(ctx),
		},
		{
			name: "trial disables nfs",
			c:    ContextBuilderFromCtx(ctx).BrokerPlan("trial").Feature(types.FeatureNfs).Build(ctx),
		},
		{
			name: "trial disables peering",
			c:    ContextBuilderFromCtx(ctx).BrokerPlan("trial").Feature(types.FeaturePeering).Build(ctx),
		},
		{
			name: "trial disables redis",
			c:    ContextBuilderFromCtx(ctx).BrokerPlan("trial").Feature(types.FeatureRedis).Build(ctx),
		},
	}

	for _, cs := range cases {
		t.Run(cs.name, func(t *testing.T) {
			assert.Equal(t, !ApiDisabled.Value(cs.c), ApiEnabled.Value(cs.c),
				"ApiEnabled must equal !ApiDisabled for case: %s", cs.name)
		})
	}
}

// TestApiEnabledSecureDefault verifies that when no apiEnabled block is present
// in the config, ApiEnabled.Value falls back to false for ordinary features
// while still returning true for the bootstrap kind-group (AC2).
func TestApiEnabledSecureDefault(t *testing.T) {
	ctx := t.Context()

	err := Initialize(ctx, logr.Discard(), WithFile("testdata/apiEnabledNoBlock.yaml"))
	assert.NoError(t, err)

	sch := commonscheme.SkrScheme

	// Non-bootstrap feature: secure default is false
	nonBootstrapCtx := ContextBuilderFromCtx(ctx).Landscape(types.LandscapeDev).Feature(types.FeatureNfs).Build(ctx)
	assert.False(t, ApiEnabled.Value(nonBootstrapCtx),
		"non-bootstrap feature must be disabled when no apiEnabled block exists")

	// Bootstrap kind-group: must always be true regardless of flag config
	bootstrapCtx := ContextBuilderFromCtx(ctx).KindsFromObject(&cloudresourcesv1beta1.CloudResources{}, sch).Build(ctx)
	assert.True(t, ApiEnabled.Value(bootstrapCtx),
		"bootstrap cloudresources kind-group must always be enabled")
}
