package config

import (
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/config"
	awsconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/config"
	azureconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/config"
	config2 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/config"
	sapconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/config"
	scopeconfig "github.com/kyma-project/cloud-manager/pkg/kcp/scope/config"
	vpcpeeringconfig "github.com/kyma-project/cloud-manager/pkg/kcp/vpcpeering/config"
	"github.com/kyma-project/cloud-manager/pkg/quota"
	skrruntimeconfig "github.com/kyma-project/cloud-manager/pkg/skr/runtime/config"
)

func CreateNewConfigAndLoad() config.Config {
	env := abstractions.NewOSEnvironment()
	configDir := env.Get("CONFIG_DIR")
	if len(configDir) < 1 {
		configDir = "./config/config"
	}
	cfg := config.NewConfig(env)
	cfg.BaseDir(configDir)

	LoadConfigInstance(cfg)

	return cfg
}

func LoadConfigInstance(cfg config.Config) {
	awsconfig.InitConfig(cfg)
	azureconfig.InitConfig(cfg)
	sapconfig.InitConfig(cfg)
	quota.InitConfig(cfg)
	skrruntimeconfig.InitConfig(cfg)
	scopeconfig.InitConfig(cfg)
	config2.InitConfig(cfg)
	vpcpeeringconfig.InitConfig(cfg)

	cfg.Read()
}
