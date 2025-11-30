package feature

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/pkg/feature/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEmbeddedConfig(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := Initialize(ctx, logr.Discard(), WithStaticConfig())
	assert.Nil(t, err)

	assert.False(t, VpcPeeringSync.Value(ctx), "vpcPeeringSync should be false")
	assert.True(t, FFNukeBackupsAws.Value(ctx), "nukeBackupsAws should be true")
	assert.True(t, FFNukeBackupsGcp.Value(ctx), "nukeBackupsGcp should be true")
	assert.False(t, Nfs41Gcp.Value(ctx), "nfs41Gcp should be false")

	assert.True(t, ApiDisabled.Value(ContextBuilderFromCtx(ctx).Feature(types.FeatureNfsBackup).Build(ctx)),
		"apiDisabled targeting feature nfsBackup should be true")

	assert.True(t, ApiDisabled.Value(ContextBuilderFromCtx(ctx).Feature(types.FeatureNfs).Provider("openstack").Build(ctx)),
		"apiDisabled targeting feature nfs and provider openstack should be true")

	assert.True(t, ApiDisabled.Value(ContextBuilderFromCtx(ctx).Feature(types.FeatureRedisCluster).Build(ctx)),
		"apiDisabled targeting feature rediscluster should be true")

	assert.True(t, ApiDisabled.Value(ContextBuilderFromCtx(ctx).Feature(types.FeatureVpcDnsLink).Build(ctx)),
		"apiDisabled targeting feature vpcdnslink should be true")

	assert.True(t, ApiDisabled.Value(ContextBuilderFromCtx(ctx).BrokerPlan("trial").Build(ctx)),
		"apiDisabled targeting broker plan trial should be true")

	assert.False(t, ApiDisabled.Value(ContextBuilderFromCtx(ctx).Feature(types.FeaturePeering).Build(ctx)),
		"apiDisabled targeting feature peering should be false")

	assert.False(t, ApiDisabled.Value(ContextBuilderFromCtx(ctx).Feature(types.FeatureNfs).Build(ctx)),
		"apiDisabled targeting feature nfs should be false")

	assert.False(t, ApiDisabled.Value(ContextBuilderFromCtx(ctx).Feature(types.FeatureRedis).Build(ctx)),
		"apiDisabled targeting feature redis should be false")
}
