package config

import (
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/config"
	"k8s.io/client-go/tools/clientcmd"
)

var Config = &ConfigType{}

type NetworkOwner string

const (
	NetworkOwnerKyma     = NetworkOwner("kyma")
	NetworkOwnerGardener = NetworkOwner("gardener")
)

type ConfigType struct {
	GardenKubeconfig string `yaml:"gardenKubeconfig" json:"gardenKubeconfig"`

	KcpNamespace    string `yaml:"kcpNamespace" json:"kcpNamespace"`
	GardenNamespace string `yaml:"gardenNamespace" json:"gardenNamespace"`
	SkrNamespace    string `yaml:"skrNamespace" json:"skrNamespace"`

	Subscriptions Subscriptions `yaml:"subscriptions" json:"subscriptions"`

	NetworkOwner NetworkOwner `yaml:"networkOwner" json:"networkOwner"`

	OidcClientId  string `yaml:"oidcClientId" json:"oidcClientId"`
	OidcIssuerUrl string `yaml:"oidcIssuerUrl" json:"oidcIssuerUrl"`

	Administrators []string `yaml:"administrators" json:"administrators"`

	CloudProfiles map[string]string `yaml:"cloudProfiles" json:"cloudProfiles"`
}

func (c *ConfigType) SetGardenNamespaceFromKubeconfigBytes(gardenKubeBytes []byte) error {
	oc, err := clientcmd.NewClientConfigFromBytes(gardenKubeBytes)
	if err != nil {
		return fmt.Errorf("error creating gardener client config: %w", err)
	}

	rawConfig, err := oc.RawConfig()
	if err != nil {
		return fmt.Errorf("error getting gardener raw client config: %w", err)
	}

	if len(rawConfig.CurrentContext) > 0 {
		if rawConfig.Contexts[rawConfig.CurrentContext].Namespace != "" {
			c.GardenNamespace = rawConfig.Contexts[rawConfig.CurrentContext].Namespace
		}
	}
	if c.GardenNamespace == "" {
		c.GardenNamespace = "garden"
	}

	return nil
}

type Subscriptions []SubscriptionInfo

type SubscriptionInfo struct {
	Name      string                           `yaml:"name" json:"name"`
	Provider  cloudcontrolv1beta1.ProviderType `yaml:"provider" json:"provider"`
	IsDefault bool                             `yaml:"isDefault" json:"isDefault"`
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
		config.SourceFile("e2e-config.yaml"),

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
			config.DefaultScalar("default"),
		),

		config.Path(
			"networkOwner",
			config.DefaultScalar(NetworkOwnerGardener),
		),

		config.Path(
			"cloudProfiles",
			config.DefaultObj(map[string]string{
				"aws":       "aws",
				"azure":     "az",
				"gcp":       "gcp",
				"openstack": "converged-cloud-kyma",
			}),
		),
	)
}

func LoadConfig() config.Config {
	env := abstractions.NewOSEnvironment()
	cfg := config.NewConfig(env)
	configDir := env.Get("CONFIG_DIR")
	if configDir == "" {
		configDir = env.Get("PROJECTROOT")
	}
	if configDir == "" {
		configDir = "../../../../"
	}
	cfg.BaseDir(configDir)
	InitConfig(cfg)
	cfg.Read()

	return cfg
}
