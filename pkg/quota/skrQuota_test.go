package quota

import (
	"testing"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	commonscheme "github.com/kyma-project/cloud-manager/pkg/common/scheme"
	"github.com/kyma-project/cloud-manager/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestSkrConfig(t *testing.T) {
	cfg := config.NewConfig(abstractions.NewMockedEnvironment(nil))
	cfg.Path("resourceQuota.skr",
		config.SourceFile("testdata/skrQuotaConfig.yaml"),
		config.DefaultObj(DefaultSkrQuota()),
		config.Bind(SkrQuota),
	)

	// defaults in case config file is not loaded
	assert.Equal(t, 1, SkrQuota.TotalCountForObj(&cloudresourcesv1beta1.IpRange{}, commonscheme.SkrScheme, "anyskr"))
	assert.Equal(t, 5, SkrQuota.TotalCountForObj(&cloudresourcesv1beta1.AwsNfsVolume{}, commonscheme.SkrScheme, "anyskr"))

	// when config file is laoded
	cfg.Read()

	// values form the config file
	assert.Equal(t, 2, SkrQuota.TotalCountForObj(&cloudresourcesv1beta1.IpRange{}, commonscheme.SkrScheme, "anyskr"))
	assert.Equal(t, 3, SkrQuota.TotalCountForObj(&cloudresourcesv1beta1.IpRange{}, commonscheme.SkrScheme, "skr123"))

	assert.Equal(t, 10, SkrQuota.TotalCountForObj(&cloudresourcesv1beta1.AwsNfsVolume{}, commonscheme.SkrScheme, "anyskr"))
	assert.Equal(t, 10, SkrQuota.TotalCountForObj(&cloudresourcesv1beta1.AwsNfsVolume{}, commonscheme.SkrScheme, "skr123"))

	// when override is set for any skr
	SkrQuota.Override(&cloudresourcesv1beta1.IpRange{}, commonscheme.SkrScheme, "", 100)
	assert.Equal(t, 100, SkrQuota.TotalCountForObj(&cloudresourcesv1beta1.IpRange{}, commonscheme.SkrScheme, "anyskr"))
	assert.Equal(t, 3, SkrQuota.TotalCountForObj(&cloudresourcesv1beta1.IpRange{}, commonscheme.SkrScheme, "skr123"))

	// when override is set for specific skr
	SkrQuota.Override(&cloudresourcesv1beta1.IpRange{}, commonscheme.SkrScheme, "skr123", 200)
	assert.Equal(t, 100, SkrQuota.TotalCountForObj(&cloudresourcesv1beta1.IpRange{}, commonscheme.SkrScheme, "anyskr"))
	assert.Equal(t, 200, SkrQuota.TotalCountForObj(&cloudresourcesv1beta1.IpRange{}, commonscheme.SkrScheme, "skr123"))
}
