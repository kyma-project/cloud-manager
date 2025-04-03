package config

import (
	"github.com/kyma-project/cloud-manager/pkg/config"
)

type VpcPeeringConfigStruct struct {
	NetworkTag               string `json:"networkTag,omitempty" yaml:"networkTag,omitempty"`
	RouteAsociatedCidrBlocks bool   `json:"routeAsociatedCidrBlocks,omitempty" yaml:"routeAsociatedCidrBlocks,omitempty"`
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
		config.Path(
			"routeAsociatedCidrBlocks",
			config.SourceEnv("ROUTE_ASSOCIATED_CIDR_BLOCKS"),
			config.SourceFile("ROUTE_ASSOCIATED_CIDR_BLOCKS"),
			config.DefaultScalar(true),
		),
	)
}
