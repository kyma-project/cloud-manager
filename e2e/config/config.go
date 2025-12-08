package config

import (
	"fmt"
	"os"
	"path"
	"strings"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	commonscheme "github.com/kyma-project/cloud-manager/pkg/common/scheme"
	"github.com/kyma-project/cloud-manager/pkg/config"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type NetworkOwner string

const (
	NetworkOwnerKyma     = NetworkOwner("kyma")
	NetworkOwnerGardener = NetworkOwner("gardener")
)

type ConfigType struct {
	GardenKubeconfig string `yaml:"gardenKubeconfig" json:"gardenKubeconfig"`

	ShootPrefix      string            `yaml:"shootPrefix" json:"shootPrefix"`
	ShootAnnotations map[string]string `yaml:"shootAnnotations" json:"shootAnnotations"`

	KcpNamespace    string `yaml:"kcpNamespace" json:"kcpNamespace"`
	GardenNamespace string `yaml:"gardenNamespace" json:"gardenNamespace"`
	SkrNamespace    string `yaml:"skrNamespace" json:"skrNamespace"`

	Subscriptions Subscriptions `yaml:"subscriptions" json:"subscriptions"`

	NetworkOwner NetworkOwner `yaml:"networkOwner" json:"networkOwner"`

	OidcClientId  string `yaml:"oidcClientId" json:"oidcClientId"`
	OidcIssuerUrl string `yaml:"oidcIssuerUrl" json:"oidcIssuerUrl"`

	Administrators []string `yaml:"administrators" json:"administrators"`

	CloudProfiles map[string]string `yaml:"cloudProfiles" json:"cloudProfiles"`

	OverwriteGardenerCredentials bool `yaml:"overwriteGardenerCredentials" json:"overwriteGardenerCredentials"`

	ConfigDir      string `yaml:"configDir" json:"configDir"`
	CredentialsDir string `yaml:"credentialsDir" json:"credentialsDir"`
	TfWorkspaceDir string `yaml:"tfWorkspaceDir" json:"tfWorkspaceDir"`
	TfCmd          string `yaml:"tfCmd" json:"tfCmd"`

	DownloadGardenSecrets map[string]map[string]string `yaml:"downloadGardenSecrets" json:"downloadGardenSecrets"`
}

func (c *ConfigType) AfterConfigLoaded() {
	// strip all whitespaces
	c.ShootPrefix = strings.Join(strings.Fields(c.ShootPrefix), "")
	c.ShootPrefix = strings.ReplaceAll(c.ShootPrefix, "-", "")
	c.ShootPrefix = strings.ReplaceAll(c.ShootPrefix, "_", "")
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

func (c *ConfigType) CreateGardenClient() (client.Client, error) {
	kubeBytes, err := os.ReadFile(c.GardenKubeconfig)
	if err != nil {
		return nil, fmt.Errorf("error reading gardener kubeconfig: %w", err)
	}

	if err := c.SetGardenNamespaceFromKubeconfigBytes(kubeBytes); err != nil {
		return nil, fmt.Errorf("error setting gardener namespace from kubeconfig: %w", err)
	}

	restConfig, err := clientcmd.RESTConfigFromKubeConfig(kubeBytes)
	if err != nil {
		return nil, fmt.Errorf("error creating gardener rest config: %w", err)
	}

	clnt, err := client.New(restConfig, client.Options{
		Scheme: commonscheme.GardenScheme,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating gardener client: %w", err)
	}

	return clnt, nil
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

func initConfig(cfg config.Config, obj *ConfigType) {
	cfg.Path(
		"e2e.config",
		config.Bind(obj),
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

		config.Path(
			"tfCmd",
			config.DefaultScalar("tofu"),
		),
	)
}

func DefaultConfig() *ConfigType {
	env := abstractions.NewMockedEnvironment(nil)
	cfg := config.NewConfig(env)
	result := &ConfigType{}
	initConfig(cfg, result)
	cfg.Read()

	return result
}

func LoadConfig() *ConfigType {
	env := abstractions.NewOSEnvironment()
	cfg := config.NewConfig(env)
	configDir := env.Get("CONFIG_DIR")
	if configDir == "" {
		configDir = env.Get("PROJECTROOT")
	}
	if configDir == "" {
		configDir = "../../../../"
	}
	result := &ConfigType{}
	result.ConfigDir = configDir
	if !strings.HasPrefix(configDir, "/") {
		wd, err := os.Getwd()
		if err == nil {
			result.ConfigDir = path.Join(wd, configDir)
		}
	}
	result.CredentialsDir = path.Join(result.ConfigDir, "tmp")
	result.TfWorkspaceDir = path.Join(result.CredentialsDir, "tf-workspaces")
	cfg.BaseDir(configDir)
	initConfig(cfg, result)
	cfg.Read()

	return result
}

func Stub() *ConfigType {
	result := DefaultConfig()
	result.ShootPrefix = "e"
	result.OidcClientId = "79221501-5dcc-4285-9af6-d023f313918e"
	result.OidcIssuerUrl = "https://oidc.e2e.cloud-manager.kyma.local"
	result.Administrators = []string{"admin@e2e.cloud-manager.kyma.local"}
	result.Subscriptions = Subscriptions{
		{
			Name:     "aws",
			Provider: cloudcontrolv1beta1.ProviderAws,
		},
		{
			Name:     "gcp",
			Provider: cloudcontrolv1beta1.ProviderGCP,
		},
		{
			Name:     "azure",
			Provider: cloudcontrolv1beta1.ProviderAzure,
		},
		{
			Name:     "openstack",
			Provider: cloudcontrolv1beta1.ProviderOpenStack,
		},
	}
	return result
}
