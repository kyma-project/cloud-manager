package feature

import (
	"testing"

	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/pkg/feature/types"
	"github.com/stretchr/testify/assert"
)

func TestEmbeddedConfig(t *testing.T) {
	ctx := t.Context()
	restoreProvider(t)

	err := Initialize(ctx, logr.Discard(), WithEmbeddedRetriever())
	assert.Nil(t, err)

	assert.False(t, VpcPeeringSync.Value(ctx), "vpcPeeringSync should be false")
	assert.True(t, FFNukeBackupsAws.Value(ctx), "nukeBackupsAws should be true")
	assert.True(t, FFNukeBackupsGcp.Value(ctx), "nukeBackupsGcp should be true")
	assert.False(t, Nfs41Gcp.Value(ctx), "nfs41Gcp should be false")

	assert.False(t, ApiEnabled.Value(ContextBuilderFromCtx(ctx).Feature(types.FeatureNfsBackup).Build(ctx)),
		"apiEnabled targeting feature nfsBackup should be false")

	assert.False(t, ApiEnabled.Value(ContextBuilderFromCtx(ctx).Feature(types.FeatureNfs).Provider("openstack").Build(ctx)),
		"apiEnabled targeting feature nfs and provider openstack should be false")

	assert.False(t, ApiEnabled.Value(ContextBuilderFromCtx(ctx).Feature(types.FeatureRedisCluster).Build(ctx)),
		"apiEnabled targeting feature rediscluster should be false")

	assert.False(t, ApiEnabled.Value(ContextBuilderFromCtx(ctx).Feature(types.FeatureVpcDnsLink).Build(ctx)),
		"apiEnabled targeting feature vpcdnslink should be false")

	assert.False(t, ApiEnabled.Value(ContextBuilderFromCtx(ctx).BrokerPlan("trial").Build(ctx)),
		"apiEnabled targeting broker plan trial should be false")

	assert.False(t, ApiEnabled.Value(ContextBuilderFromCtx(ctx).Feature(types.FeatureAzureManagedRedis).Build(ctx)),
		"apiEnabled targeting feature azureManagedRedis should be false")

	assert.False(t, ApiEnabled.Value(ContextBuilderFromCtx(ctx).Feature(types.FeatureRedis).Provider("alicloud").Build(ctx)),
		"apiEnabled targeting feature redis and provider alicloud should be false")

	assert.False(t, ApiEnabled.Value(ContextBuilderFromCtx(ctx).Feature(types.FeatureRedisCluster).Provider("alicloud").Build(ctx)),
		"apiEnabled targeting feature rediscluster and provider alicloud should be false")

	assert.True(t, ApiEnabled.Value(ContextBuilderFromCtx(ctx).Feature(types.FeaturePeering).Provider("gcp").Build(ctx)),
		"apiEnabled targeting feature peering with gcp provider should be true")

	assert.True(t, ApiEnabled.Value(ContextBuilderFromCtx(ctx).Feature(types.FeatureNfs).Provider("gcp").Build(ctx)),
		"apiEnabled targeting feature nfs with gcp provider should be true")

	assert.True(t, ApiEnabled.Value(ContextBuilderFromCtx(ctx).Feature(types.FeatureRedis).Provider("gcp").Build(ctx)),
		"apiEnabled targeting feature redis with gcp provider should be true")
}
