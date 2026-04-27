package config

import "github.com/kyma-project/cloud-manager/pkg/config"

type VpcNetworkConfigType struct {
	Prefix string `yaml:"prefix,omitempty" json:"prefix,omitempty"`
}

var VpcNetworkConfig = &VpcNetworkConfigType{}

func InitConfig(cfg config.Config) {
	cfg.Path(
		"vpcNetwork",
		config.Bind(VpcNetworkConfig),
		config.SourceFile("vpcnetwork.yaml"),
		config.Path(
			"prefix",
			config.DefaultScalar("default"),
			config.SourceEnv("VPC_NETWORK_PREFIX"),
		),
	)
}
