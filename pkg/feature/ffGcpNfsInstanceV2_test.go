package feature

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/pkg/feature/types"
	"github.com/stretchr/testify/assert"
)

func TestGcpNfsInstanceV2(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := Initialize(ctx, logr.Discard(), WithFile("testdata/gcpNfsInstanceV2.yaml"))
	assert.NoError(t, err)

	t.Run("default value is false", func(t *testing.T) {
		c := ContextBuilderFromCtx(ctx).Build(ctx)
		v := GcpNfsInstanceV2.Value(c)
		assert.False(t, v, "gcpNfsInstanceV2 should be disabled by default")
	})

	t.Run("enabled on dev landscape", func(t *testing.T) {
		c := ContextBuilderFromCtx(ctx).Landscape(types.LandscapeDev).Build(ctx)
		v := GcpNfsInstanceV2.Value(c)
		assert.True(t, v, "gcpNfsInstanceV2 should be enabled on dev landscape")
	})

	t.Run("disabled on stage landscape", func(t *testing.T) {
		c := ContextBuilderFromCtx(ctx).Landscape(types.LandscapeStage).Build(ctx)
		v := GcpNfsInstanceV2.Value(c)
		assert.False(t, v, "gcpNfsInstanceV2 should be disabled on stage landscape")
	})

	t.Run("disabled on prod landscape", func(t *testing.T) {
		c := ContextBuilderFromCtx(ctx).Landscape(types.LandscapeProd).Build(ctx)
		v := GcpNfsInstanceV2.Value(c)
		assert.False(t, v, "gcpNfsInstanceV2 should be disabled on prod landscape")
	})
}

func TestGcpNfsInstanceV2_ProviderConfig(t *testing.T) {
	ctx := context.Background()

	t.Run("default value is false with provider config", func(t *testing.T) {
		InitializeFromStaticConfig(nil)
		v := GcpNfsInstanceV2.Value(ctx)
		assert.False(t, v, "gcpNfsInstanceV2 should be disabled by default with static config")
	})
}
