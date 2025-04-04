package config

import (
	"github.com/kyma-project/cloud-manager/pkg/config"
)

type VpcPeeringConfigStruct struct {
	NetworkTag string `json:"networkTag,omitempty" yaml:"networkTag,omitempty"`
}

var VpcPeeringConfig = &VpcPeeringConfigStruct{}

func InitConfig(cfg config.Config) {
	cfg.Path(
		"vpcpeering.config",
		config.Bind(VpcPeeringConfig),
		config.SourceFile("vpcpeering.yaml"),
		config.Path(
			"networkTag",
			config.SourceEnv("PEERING_NETWORK_TAG"),
			config.SourceFile("PEERING_NETWORK_TAG"),
			config.DefaultScalar(true),
		),
	)
}
