package feature

import (
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNukeBackupsAzureFollowsApiEnabled is the discriminating test: the non-parity
// fixture resolves apiEnabled→false for the azure/nfsBackup context, and the standalone
// nukeBackupsAzure flag is absent (false). The gate must report false.
func TestNukeBackupsAzureFollowsApiEnabled(t *testing.T) {
	ctx := t.Context()
	restoreProvider(t)
	require.NoError(t, Initialize(ctx, logr.Discard(), WithFile("testdata/nukeBackupsAzureNonParity.yaml")))

	assert.False(t, FFNukeBackupsAzure.Value(ctx),
		"gate must follow apiEnabled (false)")
}

// TestNukeBackupsAzureEnabledByApi asserts the gate is enabled when apiEnabled resolves
// to enabled for the azure/nfsBackup context.
func TestNukeBackupsAzureEnabledByApi(t *testing.T) {
	ctx := t.Context()
	restoreProvider(t)
	require.NoError(t, Initialize(ctx, logr.Discard(), WithFile("testdata/nukeBackupsAzureApiEnabled.yaml")))

	assert.True(t, FFNukeBackupsAzure.Value(ctx),
		"gate must be enabled when apiEnabled resolves true for the azure/nfsBackup context")
}

// TestNukeBackupsAzureStandaloneOverride asserts the || standalone override still forces
// the gate on when apiEnabled is disabled but the nukeBackupsAzure flag is set.
func TestNukeBackupsAzureStandaloneOverride(t *testing.T) {
	ctx := t.Context()
	restoreProvider(t)
	require.NoError(t, Initialize(ctx, logr.Discard(), WithFile("testdata/nukeBackupsAzureStandalone.yaml")))

	assert.True(t, FFNukeBackupsAzure.Value(ctx),
		"standalone nukeBackupsAzure flag must force the gate on regardless of apiEnabled")
}
