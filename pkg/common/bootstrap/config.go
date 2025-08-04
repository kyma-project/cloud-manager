package bootstrap

import (
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/config"
	awsconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/config"
	azureconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/config"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	sapconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/config"
	"github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	vpcpeeringconfig "github.com/kyma-project/cloud-manager/pkg/kcp/vpcpeering/config"
	"github.com/kyma-project/cloud-manager/pkg/quota"
	skrruntimeconfig "github.com/kyma-project/cloud-manager/pkg/skr/runtime/config"
)

func LoadConfig() config.Config {
	env := abstractions.NewOSEnvironment()
	configDir := env.Get("CONFIG_DIR")
	if len(configDir) < 1 {
		configDir = "./config/config"
	}
	cfg := config.NewConfig(env)
	cfg.BaseDir(configDir)

	awsconfig.InitConfig(cfg)
	azureconfig.InitConfig(cfg)
	sapconfig.InitConfig(cfg)
	quota.InitConfig(cfg)
	skrruntimeconfig.InitConfig(cfg)
	scope.InitConfig(cfg)
	gcpclient.InitConfig(cfg)
	vpcpeeringconfig.InitConfig(cfg)

	cfg.Read()

	return cfg
}
