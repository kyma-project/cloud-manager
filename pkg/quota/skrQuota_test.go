package quota

import (
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/config"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
	"testing"
)

func TestSkrConfig(t *testing.T) {
	skrScheme := runtime.NewScheme()
	_ = cloudresourcesv1beta1.AddToScheme(skrScheme)

	cfg := config.NewConfig(abstractions.NewMockedEnvironment(nil))
	cfg.Path("resourceQuota.skr",
		config.SourceFile("testdata/skrQuotaConfig.yaml"),
		config.DefaultObj(DefaultSkrQuota()),
		config.Bind(SkrQuota),
	)

	// defaults in case config file is not loaded
	assert.Equal(t, 1, SkrQuota.TotalCountForObj(&cloudresourcesv1beta1.IpRange{}, skrScheme, "anyskr"))
	assert.Equal(t, 5, SkrQuota.TotalCountForObj(&cloudresourcesv1beta1.AwsNfsVolume{}, skrScheme, "anyskr"))

	// when config file is laoded
	cfg.Read()

	// values form the config file
	assert.Equal(t, 2, SkrQuota.TotalCountForObj(&cloudresourcesv1beta1.IpRange{}, skrScheme, "anyskr"))
	assert.Equal(t, 3, SkrQuota.TotalCountForObj(&cloudresourcesv1beta1.IpRange{}, skrScheme, "skr123"))

	assert.Equal(t, 10, SkrQuota.TotalCountForObj(&cloudresourcesv1beta1.AwsNfsVolume{}, skrScheme, "anyskr"))
	assert.Equal(t, 10, SkrQuota.TotalCountForObj(&cloudresourcesv1beta1.AwsNfsVolume{}, skrScheme, "skr123"))

	// when override is set for any skr
	SkrQuota.Override(&cloudresourcesv1beta1.IpRange{}, skrScheme, "", 100)
	assert.Equal(t, 100, SkrQuota.TotalCountForObj(&cloudresourcesv1beta1.IpRange{}, skrScheme, "anyskr"))
	assert.Equal(t, 3, SkrQuota.TotalCountForObj(&cloudresourcesv1beta1.IpRange{}, skrScheme, "skr123"))

	// when override is set for specific skr
	SkrQuota.Override(&cloudresourcesv1beta1.IpRange{}, skrScheme, "skr123", 200)
	assert.Equal(t, 100, SkrQuota.TotalCountForObj(&cloudresourcesv1beta1.IpRange{}, skrScheme, "anyskr"))
	assert.Equal(t, 200, SkrQuota.TotalCountForObj(&cloudresourcesv1beta1.IpRange{}, skrScheme, "skr123"))
}
