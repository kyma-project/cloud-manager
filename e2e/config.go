package e2e

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/config"
)

var Config = &ConfigType{}

type NetworkOwner string

const (
	NetworkOwnerKyma = NetworkOwner("kyma")
	NetworkOwnerGardener = NetworkOwner("gardener")
)

type ConfigType struct {
	GardenKubeconfig string `yaml:"gardenKubeconfig"`

	KcpNamespace  string `yaml:"kcpNamespace"`
	GardenNamespace string `yaml:"gardenNamespace"`
	SkrNamespace   string `yaml:"skrNamespace"`

	Subscriptions Subscriptions `yaml:"subscriptions"`

	NetworkOwner NetworkOwner `yaml:"networkOwner"`

	OidcClientId string `yaml:"oidcClientId"`
	OidcIssuerUrl string `yaml:"oidcIssuerUrl"`

	Administrators []string `yaml:"administrators"`
}

type Subscriptions []SubscriptionInfo

type SubscriptionInfo struct {
	Name      string                           `yaml:"name"`
	Provider  cloudcontrolv1beta1.ProviderType `yaml:"provider"`
	IsDefault bool                             `yaml:"isDefault"`
}

func (s Subscriptions) FindFirst(cb func(s SubscriptionInfo) bool) *SubscriptionInfo {
	for _, sub := range s {
		if cb(sub) {
			return &sub
		}
	}
	return nil
}

func (s Subscriptions) GetDefaultForProvider(provider cloudcontrolv1beta1.ProviderType) *SubscriptionInfo {
	result := s.FindFirst(func(s SubscriptionInfo) bool {
		return s.Provider == provider && s.IsDefault
	})
	if result != nil {
		return result
	}
	return s.FindFirst(func(s SubscriptionInfo) bool {
		return s.Provider == provider
	})
}

func InitConfig(cfg config.Config) {
	cfg.Path(
		"e2e.config",
		config.Bind(Config),
		// intention is to load the whole config from a file
		config.SourceFile("../e2e-config.yaml"),

		// below are some defaults that can be omitted from the file
		config.Path(
			"gardenKubeconfig",
			config.SourceEnv("GARDEN_KUBECONFIG"),
		),
		config.Path(
			"kcpNamespace",
			config.DefaultScalar("kcp-system"),
		),
		config.Path(
			"gardenNamespace",
			config.DefaultScalar("garden"),
		),
		config.Path(
			"skrNamespace",
			config.DefaultScalar("skr"),
		),

		config.Path(
			"networkOwner",
			config.DefaultScalar(NetworkOwnerGardener),
		),
	)
}
