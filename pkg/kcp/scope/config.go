package scope

import "github.com/kyma-project/cloud-manager/pkg/config"

type ScopeConfigStruct struct {
	GardenerNamespace string `yaml:"gardenerNamespace,omitempty" json:"gardenerNamespace,omitempty"`
}

var ScopeConfig = &ScopeConfigStruct{}

func InitConfig(cfg config.Config) {
	cfg.Path(
		"scope",
		config.Path(
			"gardenerNamespace",
			config.DefaultScalar("garden-kyma-dev"),
			config.SourceEnv("GARDENER_NAMESPACE"),
		),
		config.SourceFile("scope.yaml"),
		config.Bind(ScopeConfig),
	)
}
