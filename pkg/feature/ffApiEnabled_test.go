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

// TestApiEnabled asserts the absolute wantEnabled values for every case when
// loaded from testdata/apiEnabled.yaml. Bootstrap cloudresources cases also
// carry an absolute assert.True anchor so that the isBootstrapKindGroup
// short-circuit is verified independently of config.
func TestApiEnabled(t *testing.T) {
	ctx := t.Context()
	restoreProvider(t)

	err := Initialize(ctx, logr.Discard(), WithFile("testdata/apiEnabled.yaml"))
	assert.NoError(t, err)

	sch := commonscheme.SkrScheme

	cases := []struct {
		t             string
		c             context.Context
		wantEnabled   bool
		mustBeEnabled bool
	}{
		// CloudResources — bootstrap kind-group must always be enabled
		{
			t:             "cloudresources is enabled regardless of landscape and feature",
			c:             ContextBuilderFromCtx(ctx).KindsFromObject(&cloudresourcesv1beta1.CloudResources{}, sch).Build(ctx),
			wantEnabled:   true,
			mustBeEnabled: true,
		},
		{
			t:             "cloudresources is enabled on dev",
			c:             ContextBuilderFromCtx(ctx).Landscape("dev").KindsFromObject(&cloudresourcesv1beta1.CloudResources{}, sch).Build(ctx),
			wantEnabled:   true,
			mustBeEnabled: true,
		},
		{
			t:             "cloudresources is enabled on stage",
			c:             ContextBuilderFromCtx(ctx).Landscape("stage").KindsFromObject(&cloudresourcesv1beta1.CloudResources{}, sch).Build(ctx),
			wantEnabled:   true,
			mustBeEnabled: true,
		},
		{
			t:             "cloudresources is enabled on prod",
			c:             ContextBuilderFromCtx(ctx).Landscape("prod").KindsFromObject(&cloudresourcesv1beta1.CloudResources{}, sch).Build(ctx),
			wantEnabled:   true,
			mustBeEnabled: true,
		},
		{
			t:             "cloudresources is enabled on prod even for disabled feature",
			c:             ContextBuilderFromCtx(ctx).Landscape("prod").Feature(types.FeatureNfsBackup).KindsFromObject(&cloudresourcesv1beta1.CloudResources{}, sch).Build(ctx),
			wantEnabled:   true,
			mustBeEnabled: true,
		},
		{
			t:             "cloudresources is enabled on trial",
			c:             ContextBuilderFromCtx(ctx).BrokerPlan("trial").Landscape("prod").Feature(types.FeatureNfsBackup).KindsFromObject(&cloudresourcesv1beta1.CloudResources{}, sch).Build(ctx),
			wantEnabled:   true,
			mustBeEnabled: true,
		},
		// NFS ====================================================
		{
			t:           "nfs feature is enabled on dev",
			c:           ContextBuilderFromCtx(ctx).Landscape(types.LandscapeDev).Feature(types.FeatureNfs).Build(ctx),
			wantEnabled: true,
		},
		{
			t:           "nfs feature is enabled on stage",
			c:           ContextBuilderFromCtx(ctx).Landscape(types.LandscapeStage).Feature(types.FeatureNfs).Build(ctx),
			wantEnabled: true,
		},
		{
			t:           "nfs feature is enabled on prod",
			c:           ContextBuilderFromCtx(ctx).Landscape(types.LandscapeProd).Feature(types.FeatureNfs).BrokerPlan("aws").Build(ctx),
			wantEnabled: true,
		},
		// NFS BACKUP ====================================================
		{
			t:           "nfsBackup feature is enabled on dev",
			c:           ContextBuilderFromCtx(ctx).Landscape(types.LandscapeDev).Feature(types.FeatureNfsBackup).Build(ctx),
			wantEnabled: true,
		},
		{
			t:           "nfsBackup feature is disabled on stage",
			c:           ContextBuilderFromCtx(ctx).Landscape(types.LandscapeStage).Feature(types.FeatureNfsBackup).Build(ctx),
			wantEnabled: false,
		},
		{
			t:           "nfsBackup feature is disabled on prod",
			c:           ContextBuilderFromCtx(ctx).Landscape(types.LandscapeProd).Feature(types.FeatureNfsBackup).Build(ctx),
			wantEnabled: false,
		},
		// PEERING ====================================================
		{
			t:           "peering feature is enabled on dev",
			c:           ContextBuilderFromCtx(ctx).Landscape(types.LandscapeDev).Feature(types.FeaturePeering).Build(ctx),
			wantEnabled: true,
		},
		{
			t:           "peering feature is disabled on stage",
			c:           ContextBuilderFromCtx(ctx).Landscape(types.LandscapeStage).Feature(types.FeaturePeering).Build(ctx),
			wantEnabled: false,
		},
		{
			t:           "peering feature is disabled on prod",
			c:           ContextBuilderFromCtx(ctx).Landscape(types.LandscapeProd).Feature(types.FeaturePeering).Build(ctx),
			wantEnabled: false,
		},
		// REDIS ====================================================
		{
			t:           "redis feature is enabled on dev",
			c:           ContextBuilderFromCtx(ctx).Landscape(types.LandscapeDev).Feature(types.FeatureRedis).Build(ctx),
			wantEnabled: true,
		},
		{
			t:           "redis feature is disabled on stage",
			c:           ContextBuilderFromCtx(ctx).Landscape(types.LandscapeStage).Feature(types.FeatureRedis).Build(ctx),
			wantEnabled: false,
		},
		{
			t:           "redis feature is disabled on prod",
			c:           ContextBuilderFromCtx(ctx).Landscape(types.LandscapeProd).Feature(types.FeatureRedis).Build(ctx),
			wantEnabled: false,
		},
		// TRIAL ====================================================
		{
			t:           "nfs feature is disabled on trial",
			c:           ContextBuilderFromCtx(ctx).BrokerPlan("trial").Feature(types.FeatureNfs).Build(ctx),
			wantEnabled: false,
		},
		{
			t:           "nfsbackup feature is disabled on trial",
			c:           ContextBuilderFromCtx(ctx).BrokerPlan("trial").Feature(types.FeatureNfsBackup).Build(ctx),
			wantEnabled: false,
		},
		{
			t:           "peering feature is disabled on trial",
			c:           ContextBuilderFromCtx(ctx).BrokerPlan("trial").Feature(types.FeaturePeering).Build(ctx),
			wantEnabled: false,
		},
		{
			t:           "redis feature is disabled on trial",
			c:           ContextBuilderFromCtx(ctx).BrokerPlan("trial").Feature(types.FeatureRedis).Build(ctx),
			wantEnabled: false,
		},
		// GlobalAccount exemption ====================================================
		{
			t:           "globalAccount exemption on stage enables nfsBackup",
			c:           ContextBuilderFromCtx(ctx).Landscape("stage").GlobalAccount("6034adc9-3441-494c-b438-a0eba4c5045a").Feature(types.FeatureNfsBackup).Build(ctx),
			wantEnabled: true,
		},
		{
			// Trial rule is declared before the globalAccount exemption, so GoFF
			// first-match fires the trial disable: trial wins over the exemption.
			t:           "trial overrides the globalAccount exemption on stage",
			c:           ContextBuilderFromCtx(ctx).BrokerPlan("trial").Landscape("stage").GlobalAccount("6034adc9-3441-494c-b438-a0eba4c5045a").Feature(types.FeatureNfsBackup).Build(ctx),
			wantEnabled: false,
		},
	}

	for _, cs := range cases {
		t.Run(cs.t, func(t *testing.T) {
			assert.Equal(t, cs.wantEnabled, ApiEnabled.Value(cs.c),
				"ApiEnabled must be %v for case: %s", cs.wantEnabled, cs.t)
			if cs.mustBeEnabled {
				assert.True(t, ApiEnabled.Value(cs.c),
					"bootstrap cloudresources must always be ApiEnabled for case: %s", cs.t)
			}
		})
	}
}

// TestApiEnabledEmbeddedGAParity asserts the authoritative enabled-set for the embedded
// ff_ga.yaml GA allowlist. Each case declares the expected ApiEnabled.Value directly.
// foundation and unknown cases are discriminating for the always-on hardcode: they must be
// absolutely true regardless of config.
func TestApiEnabledEmbeddedGAParity(t *testing.T) {
	ctx := t.Context()
	restoreProvider(t)

	err := Initialize(ctx, logr.Discard(), WithEmbeddedRetriever())
	assert.NoError(t, err)

	cases := []struct {
		name          string
		c             context.Context
		wantEnabled   bool
		mustBeEnabled bool
	}{
		// GA-enabled features (provider-scoped: test on gcp which lists all three)
		{
			name:        "peering is enabled on gcp",
			c:           ContextBuilderFromCtx(ctx).Feature(types.FeaturePeering).Provider("gcp").Build(ctx),
			wantEnabled: true,
		},
		{
			name:        "nfs is enabled on gcp",
			c:           ContextBuilderFromCtx(ctx).Feature(types.FeatureNfs).Provider("gcp").Build(ctx),
			wantEnabled: true,
		},
		{
			name:        "redis is enabled on gcp",
			c:           ContextBuilderFromCtx(ctx).Feature(types.FeatureRedis).Provider("gcp").Build(ctx),
			wantEnabled: true,
		},
		// Disable exceptions within enabled set
		{
			name:        "nfs on openstack is disabled",
			c:           ContextBuilderFromCtx(ctx).Feature(types.FeatureNfs).Provider("openstack").Build(ctx),
			wantEnabled: false,
		},
		{
			name:        "redis on alicloud is disabled",
			c:           ContextBuilderFromCtx(ctx).Feature(types.FeatureRedis).Provider("alicloud").Build(ctx),
			wantEnabled: false,
		},
		// Features fully disabled in GA
		{
			name:        "nfsBackup is disabled",
			c:           ContextBuilderFromCtx(ctx).Feature(types.FeatureNfsBackup).Build(ctx),
			wantEnabled: false,
		},
		{
			name:        "rediscluster is disabled",
			c:           ContextBuilderFromCtx(ctx).Feature(types.FeatureRedisCluster).Build(ctx),
			wantEnabled: false,
		},
		{
			name:        "rediscluster on alicloud is disabled",
			c:           ContextBuilderFromCtx(ctx).Feature(types.FeatureRedisCluster).Provider("alicloud").Build(ctx),
			wantEnabled: false,
		},
		{
			name:        "vpcdnslink is disabled",
			c:           ContextBuilderFromCtx(ctx).Feature(types.FeatureVpcDnsLink).Build(ctx),
			wantEnabled: false,
		},
		{
			name:        "waf is disabled",
			c:           ContextBuilderFromCtx(ctx).Feature(types.FeatureWAF).Build(ctx),
			wantEnabled: false,
		},
		{
			name:        "certificate is disabled",
			c:           ContextBuilderFromCtx(ctx).Feature(types.FeatureCertificate).Build(ctx),
			wantEnabled: false,
		},
		{
			name:        "azureManagedRedis is disabled",
			c:           ContextBuilderFromCtx(ctx).Feature(types.FeatureAzureManagedRedis).Build(ctx),
			wantEnabled: false,
		},
		// Trial disables everything configurable
		{
			name:        "trial disables all (no feature)",
			c:           ContextBuilderFromCtx(ctx).BrokerPlan("trial").Build(ctx),
			wantEnabled: false,
		},
		{
			name:        "trial disables nfs",
			c:           ContextBuilderFromCtx(ctx).BrokerPlan("trial").Feature(types.FeatureNfs).Build(ctx),
			wantEnabled: false,
		},
		{
			name:        "trial disables peering",
			c:           ContextBuilderFromCtx(ctx).BrokerPlan("trial").Feature(types.FeaturePeering).Build(ctx),
			wantEnabled: false,
		},
		{
			name:        "trial disables redis",
			c:           ContextBuilderFromCtx(ctx).BrokerPlan("trial").Feature(types.FeatureRedis).Build(ctx),
			wantEnabled: false,
		},
		// Foundation and unknown features are always-on (hardcoded, independent of config).
		{
			name:          "foundation feature (infrastructure) is always enabled",
			c:             ContextBuilderFromCtx(ctx).Feature(types.FeatureFoundation).Build(ctx),
			wantEnabled:   true,
			mustBeEnabled: true,
		},
		{
			name:          "unknown feature is always enabled",
			c:             ContextBuilderFromCtx(ctx).Feature(types.FeatureUnknown).Build(ctx),
			wantEnabled:   true,
			mustBeEnabled: true,
		},
		// Trial restores "disable all on trial" for foundation/unknown (they fall through to the flag).
		{
			name:        "foundation falls through to the flag on trial",
			c:           ContextBuilderFromCtx(ctx).BrokerPlan("trial").Feature(types.FeatureFoundation).Build(ctx),
			wantEnabled: false,
		},
		{
			name:        "unknown falls through to the flag on trial",
			c:           ContextBuilderFromCtx(ctx).BrokerPlan("trial").Feature(types.FeatureUnknown).Build(ctx),
			wantEnabled: false,
		},
	}

	for _, cs := range cases {
		t.Run(cs.name, func(t *testing.T) {
			assert.Equal(t, cs.wantEnabled, ApiEnabled.Value(cs.c),
				"ApiEnabled must be %v for case: %s", cs.wantEnabled, cs.name)
			if cs.mustBeEnabled {
				assert.True(t, ApiEnabled.Value(cs.c),
					"foundation/unknown must be absolutely enabled for case: %s", cs.name)
			}
		})
	}
}

// TestApiEnabledEdgeParity asserts that every feature resolves to ApiEnabled=true when
// loaded from ff_edge.yaml. On edge the apiEnabled flag has no targeting rules and a
// default-enabled rule, so all feature/provider combinations are open.
func TestApiEnabledEdgeParity(t *testing.T) {
	ctx := t.Context()
	restoreProvider(t)

	err := Initialize(ctx, logr.Discard(), WithFile("ff_edge.yaml"))
	assert.NoError(t, err)

	cases := []struct {
		name string
		c    context.Context
	}{
		{
			name: "nfs is enabled",
			c:    ContextBuilderFromCtx(ctx).Feature(types.FeatureNfs).Build(ctx),
		},
		{
			name: "nfs on openstack is enabled",
			c:    ContextBuilderFromCtx(ctx).Feature(types.FeatureNfs).Provider("openstack").Build(ctx),
		},
		{
			name: "peering is enabled",
			c:    ContextBuilderFromCtx(ctx).Feature(types.FeaturePeering).Build(ctx),
		},
		{
			name: "redis is enabled",
			c:    ContextBuilderFromCtx(ctx).Feature(types.FeatureRedis).Build(ctx),
		},
		{
			name: "redis on alicloud is enabled",
			c:    ContextBuilderFromCtx(ctx).Feature(types.FeatureRedis).Provider("alicloud").Build(ctx),
		},
		{
			name: "nfsBackup is enabled",
			c:    ContextBuilderFromCtx(ctx).Feature(types.FeatureNfsBackup).Build(ctx),
		},
		{
			name: "rediscluster is enabled",
			c:    ContextBuilderFromCtx(ctx).Feature(types.FeatureRedisCluster).Build(ctx),
		},
		{
			name: "vpcdnslink is enabled",
			c:    ContextBuilderFromCtx(ctx).Feature(types.FeatureVpcDnsLink).Build(ctx),
		},
		{
			name: "waf is enabled",
			c:    ContextBuilderFromCtx(ctx).Feature(types.FeatureWAF).Build(ctx),
		},
		{
			name: "certificate is enabled",
			c:    ContextBuilderFromCtx(ctx).Feature(types.FeatureCertificate).Build(ctx),
		},
		{
			name: "azureManagedRedis is enabled",
			c:    ContextBuilderFromCtx(ctx).Feature(types.FeatureAzureManagedRedis).Build(ctx),
		},
		{
			name: "trial is enabled on edge",
			c:    ContextBuilderFromCtx(ctx).BrokerPlan("trial").Build(ctx),
		},
		{
			name: "empty feature is enabled",
			c:    ContextBuilderFromCtx(ctx).Feature(types.FeatureFoundation).Build(ctx),
		},
		{
			name: "unknown feature is enabled",
			c:    ContextBuilderFromCtx(ctx).Feature(types.FeatureUnknown).Build(ctx),
		},
	}

	for _, cs := range cases {
		t.Run(cs.name, func(t *testing.T) {
			assert.True(t, ApiEnabled.Value(cs.c),
				"every edge case must be ApiEnabled=true for case: %s", cs.name)
		})
	}
}

// TestApiEnabledSecureDefault verifies that when no apiEnabled block is present
// in the config, ApiEnabled.Value falls back to false for ordinary features
// while still returning true for the bootstrap kind-group.
func TestApiEnabledSecureDefault(t *testing.T) {
	ctx := t.Context()
	restoreProvider(t)

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

// TestApiEnabledAlwaysOnIndependentOfConfig proves that FeatureFoundation and FeatureUnknown
// are enabled even when the config has no apiEnabled block at all. Without the hardcode,
// BoolVariation would return the default (false); the test fails if the code path is removed.
func TestApiEnabledAlwaysOnIndependentOfConfig(t *testing.T) {
	ctx := t.Context()
	restoreProvider(t)

	err := Initialize(ctx, logr.Discard(), WithFile("testdata/apiEnabledNoBlock.yaml"))
	assert.NoError(t, err)

	foundationCtx := ContextBuilderFromCtx(ctx).Feature(types.FeatureFoundation).Build(ctx)
	assert.True(t, ApiEnabled.Value(foundationCtx),
		"FeatureFoundation must be enabled even with no apiEnabled block in config")

	unknownCtx := ContextBuilderFromCtx(ctx).Feature(types.FeatureUnknown).Build(ctx)
	assert.True(t, ApiEnabled.Value(unknownCtx),
		"FeatureUnknown must be enabled even with no apiEnabled block in config")
}

// TestApiEnabledTrialGatesFoundation proves the trial exception to the always-on carve-out:
// on a trial SKR, foundation and unknown fall through to the flag (which disables all on
// trial), while off trial they stay always-on and the bootstrap kind-group stays enabled
// even on trial. The matrix parity check cannot catch this (both flags stay equal), so the
// assertions here are absolute.
func TestApiEnabledTrialGatesFoundation(t *testing.T) {
	ctx := t.Context()
	restoreProvider(t)

	err := Initialize(ctx, logr.Discard(), WithFile("testdata/apiEnabled.yaml"))
	assert.NoError(t, err)

	sch := commonscheme.SkrScheme

	foundationTrial := ContextBuilderFromCtx(ctx).BrokerPlan("trial").Feature(types.FeatureFoundation).Build(ctx)
	foundationNonTrial := ContextBuilderFromCtx(ctx).Feature(types.FeatureFoundation).Build(ctx)
	unknownTrial := ContextBuilderFromCtx(ctx).BrokerPlan("trial").Feature(types.FeatureUnknown).Build(ctx)
	bootstrapTrial := ContextBuilderFromCtx(ctx).BrokerPlan("trial").KindsFromObject(&cloudresourcesv1beta1.CloudResources{}, sch).Build(ctx)

	assert.False(t, ApiEnabled.Value(foundationTrial), "foundation must be disabled on trial (trial-gating restored)")
	assert.False(t, ApiEnabled.Value(unknownTrial), "unknown must be disabled on trial")
	assert.True(t, ApiEnabled.Value(foundationNonTrial), "foundation must stay always-on off trial")
	assert.True(t, ApiEnabled.Value(bootstrapTrial), "the bootstrap kind-group must stay enabled even on trial")
}
