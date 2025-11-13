package lib

import (
	"context"
	"testing"

	e2econfig "github.com/kyma-project/cloud-manager/e2e/config"
	"github.com/stretchr/testify/assert"
)

func TestCloudProfiles(t *testing.T) {
	cfg := e2econfig.Stub()
	loader := NewFileCloudProfileLoader(CloudProfilesFS, "cloudprofiles.yaml", cfg)
	reg, err := loader.Load(context.Background())
	assert.NoError(t, err)

	var v string
	var profile CloudProfileInfo

	profile = reg.Get("aws")
	assert.NotNil(t, profile)
	v = profile.GetKubernetesVersion()
	assert.Equal(t, "1.33.3", v)
	v = profile.GetGardenLinuxVersion()
	assert.Equal(t, "1877.4.0", v)

	profile = reg.Get("az")
	assert.NotNil(t, profile)
	v = profile.GetKubernetesVersion()
	assert.Equal(t, "1.32.7", v)
	v = profile.GetGardenLinuxVersion()
	assert.Equal(t, "1877.4.0", v)

	profile = reg.Get("gcp")
	assert.NotNil(t, profile)
	v = profile.GetKubernetesVersion()
	assert.Equal(t, "1.33.3", v)
	v = profile.GetGardenLinuxVersion()
	assert.Equal(t, "1877.4.0", v)

	profile = reg.Get("converged-cloud-kyma")
	assert.NotNil(t, profile)
	v = profile.GetKubernetesVersion()
	assert.Equal(t, "1.33.3", v)
	v = profile.GetGardenLinuxVersion()
	assert.Equal(t, "1877.4.0", v)
}
