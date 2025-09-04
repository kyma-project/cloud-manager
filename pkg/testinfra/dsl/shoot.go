package dsl

import (
	"context"
	"fmt"

	gardenertypes "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	awsgardener "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/gardener"
	azuregardener "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/gardener"
	sapgardener "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/gardener"
	"github.com/kyma-project/cloud-manager/pkg/testinfra"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	DefaultGardenNamespace = "garden-kyma" // must be same as infra.Garden().Namespace()
)

func CreateGardenerCredentials(ctx context.Context, infra testinfra.Infra) error {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: infra.KCP().Namespace(),
			Name:      "gardener-credentials",
		},
	}
	err := infra.KCP().Client().Get(ctx, client.ObjectKeyFromObject(secret), secret)
	if client.IgnoreNotFound(err) != nil {
		return fmt.Errorf("error getting gardener-credentials secret: %w", err)
	}
	if apierrors.IsNotFound(err) {
		b, err := kubeconfigToBytes(restConfigToKubeconfig(infra.Garden().Cfg()))
		if err != nil {
			return fmt.Errorf("error getting garden kubeconfig bytes: %w", err)
		}
		secret.Data = map[string][]byte{
			"kubeconfig": b,
		}

		err = infra.KCP().Client().Create(ctx, secret)
		if client.IgnoreAlreadyExists(err) != nil {
			return fmt.Errorf("error creating gardener-credentials secret: %w", err)
		}
	}

	return nil
}

func CreateShootAws(ctx context.Context, infra testinfra.Infra, shoot *gardenertypes.Shoot, opts ...ObjAction) error {
	// KCP Gardener-credentials secret
	if err := CreateGardenerCredentials(ctx, infra); err != nil {
		return err
	}

	// Garden resources
	if shoot == nil {
		shoot = &gardenertypes.Shoot{}
	}
	actions := NewObjActions(opts...).
		Append(
			WithNamespace(DefaultGardenNamespace),
		)

	// Shoot
	{
		actions.ApplyOnObject(shoot)
		shoot.Spec = gardenertypes.ShootSpec{
			CloudProfileName: ptr.To("aws"),
			Region:           "eu-west-1",
			Networking: &gardenertypes.Networking{
				IPFamilies: []gardenertypes.IPFamily{gardenertypes.IPFamilyIPv4},
				Nodes:      ptr.To("10.180.0.0/16"),
				Pods:       ptr.To("100.64.0.0/12"),
				Services:   ptr.To("100.104.0.0/13"),
			},
			Provider: gardenertypes.Provider{
				Type: string(cloudcontrolv1beta1.ProviderAws),
				InfrastructureConfig: &runtime.RawExtension{
					Object: &awsgardener.InfrastructureConfig{
						TypeMeta: metav1.TypeMeta{
							Kind:       "InfrastructureConfig",
							APIVersion: "aws.provider.extensions.gardener.cloud/v1alpha1",
						},
						Networks: awsgardener.Networks{
							VPC: awsgardener.VPC{
								CIDR: ptr.To("10.180.0.0/16"),
							},
							Zones: []awsgardener.Zone{
								{
									Name:     "eu-west-1a",
									Internal: "10.180.48.0/20",
									Public:   "10.180.32.0/20",
									Workers:  "10.180.0.0/19",
								},
								{
									Name:     "eu-west-1b",
									Internal: "10.180.112.0/20",
									Public:   "10.180.96.0/20",
									Workers:  "10.180.64.0/19",
								},
								{
									Name:     "eu-west-1c",
									Internal: "10.180.176.0/20",
									Public:   "10.180.160.0/20",
									Workers:  "10.180.128.0/19",
								},
							},
						},
					},
				},
			},
			SecretBindingName: ptr.To(shoot.Name),
		}

		err := infra.Garden().Client().Create(ctx, shoot)
		if err != nil {
			return fmt.Errorf("error creating Shoot: %w", err)
		}
	}

	// SecretBinding
	{
		secretBinding := &gardenertypes.SecretBinding{}
		actions.ApplyOnObject(secretBinding)
		secretBinding.Provider = &gardenertypes.SecretBindingProvider{
			Type: string(cloudcontrolv1beta1.ProviderAws),
		}
		secretBinding.SecretRef = corev1.SecretReference{
			Name:      shoot.Name,
			Namespace: shoot.Namespace,
		}
		err := infra.Garden().Client().Create(ctx, secretBinding)
		if err != nil {
			return fmt.Errorf("error creating SecretBinding: %w", err)
		}
	}

	// Secret
	{
		secret := &corev1.Secret{}
		actions.ApplyOnObject(secret)
		secret.StringData = map[string]string{
			"accessKeyID":     "accessKeyID",
			"secretAccessKey": "secretAccessKey",
		}

		err := infra.Garden().Client().Create(ctx, secret)
		if err != nil {
			return fmt.Errorf("error creating garden secret: %w", err)
		}
	}

	return nil
}

func kubeconfigToBytes(clientConfig *clientcmdapi.Config) ([]byte, error) {
	return clientcmd.Write(*clientConfig)
}

func restConfigToKubeconfig(restConfig *rest.Config) *clientcmdapi.Config {
	clusters := make(map[string]*clientcmdapi.Cluster)
	clusters["default-cluster"] = &clientcmdapi.Cluster{
		Server:                   restConfig.Host,
		CertificateAuthorityData: restConfig.CAData,
	}

	contexts := make(map[string]*clientcmdapi.Context)
	contexts["default-context"] = &clientcmdapi.Context{
		Cluster:  "default-cluster",
		AuthInfo: "default-auth",
	}

	authinfos := make(map[string]*clientcmdapi.AuthInfo)
	authinfos["default-auth"] = &clientcmdapi.AuthInfo{
		ClientCertificateData: restConfig.CertData,
		ClientKeyData:         restConfig.KeyData,
	}

	clientConfig := &clientcmdapi.Config{
		Kind:           "Config",
		APIVersion:     "v1",
		Clusters:       clusters,
		Contexts:       contexts,
		CurrentContext: "default-context",
		AuthInfos:      authinfos,
	}

	return clientConfig
}

func CreateShootGcp(ctx context.Context, infra testinfra.Infra, shoot *gardenertypes.Shoot, opts ...ObjAction) error {
	// KCP Gardener-credentials secret
	{
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: infra.KCP().Namespace(),
				Name:      "gardener-credentials",
			},
		}
		err := infra.KCP().Client().Get(ctx, client.ObjectKeyFromObject(secret), secret)
		if client.IgnoreNotFound(err) != nil {
			return fmt.Errorf("error getting gardener-credentials secret: %w", err)
		}
		if apierrors.IsNotFound(err) {
			b, err := kubeconfigToBytes(restConfigToKubeconfig(infra.Garden().Cfg()))
			if err != nil {
				return fmt.Errorf("error getting garden kubeconfig bytes: %w", err)
			}
			secret.Data = map[string][]byte{
				"kubeconfig": b,
			}

			err = infra.KCP().Client().Create(ctx, secret)
			if client.IgnoreAlreadyExists(err) != nil {
				return fmt.Errorf("error creating gardener-credentials secret: %w", err)
			}
		}
	}

	// Garden resources
	if shoot == nil {
		shoot = &gardenertypes.Shoot{}
	}
	actions := NewObjActions(opts...).
		Append(
			WithNamespace(DefaultGardenNamespace),
		)

	// Shoot
	{
		actions.ApplyOnObject(shoot)
		shoot.Spec = gardenertypes.ShootSpec{
			CloudProfileName: ptr.To("gcp"),
			Region:           "eu-west-1",
			Networking:       &gardenertypes.Networking{},
			Provider: gardenertypes.Provider{
				Type: string(cloudcontrolv1beta1.ProviderGCP),
				Workers: []gardenertypes.Worker{
					{
						Zones: DefaultGcpWorkerZones,
					},
				},
			},
			SecretBindingName: ptr.To(shoot.Name),
		}

		err := infra.Garden().Client().Create(ctx, shoot)
		if err != nil {
			return fmt.Errorf("error creating Shoot: %w", err)
		}
	}

	// SecretBinding
	{
		secretBinding := &gardenertypes.SecretBinding{}
		actions.ApplyOnObject(secretBinding)
		secretBinding.Provider = &gardenertypes.SecretBindingProvider{
			Type: string(cloudcontrolv1beta1.ProviderGCP),
		}
		secretBinding.SecretRef = corev1.SecretReference{
			Name:      shoot.Name,
			Namespace: shoot.Namespace,
		}
		err := infra.Garden().Client().Create(ctx, secretBinding)
		if err != nil {
			return fmt.Errorf("error creating SecretBinding: %w", err)
		}
	}

	// Secret
	{
		secret := &corev1.Secret{}
		actions.ApplyOnObject(secret)
		secret.StringData = map[string]string{
			"serviceaccount.json": fmt.Sprintf(`{"project_id": "%s"}`, DefaultGcpProject),
		}
		err := infra.Garden().Client().Create(ctx, secret)
		if err != nil {
			return fmt.Errorf("error creating garden secret: %w", err)
		}
	}

	return nil
}

func CreateShootAzure(ctx context.Context, infra testinfra.Infra, shoot *gardenertypes.Shoot, opts ...ObjAction) error {
	// KCP Gardener-credentials secret
	{
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: infra.KCP().Namespace(),
				Name:      "gardener-credentials",
			},
		}
		err := infra.KCP().Client().Get(ctx, client.ObjectKeyFromObject(secret), secret)
		if client.IgnoreNotFound(err) != nil {
			return fmt.Errorf("error getting gardener-credentials secret: %w", err)
		}
		if apierrors.IsNotFound(err) {
			b, err := kubeconfigToBytes(restConfigToKubeconfig(infra.Garden().Cfg()))
			if err != nil {
				return fmt.Errorf("error getting garden kubeconfig bytes: %w", err)
			}
			secret.Data = map[string][]byte{
				"kubeconfig": b,
			}

			err = infra.KCP().Client().Create(ctx, secret)
			if client.IgnoreAlreadyExists(err) != nil {
				return fmt.Errorf("error creating gardener-credentials secret: %w", err)
			}
		}
	}

	// Garden resources
	if shoot == nil {
		shoot = &gardenertypes.Shoot{}
	}
	actions := NewObjActions(opts...).
		Append(
			WithNamespace(DefaultGardenNamespace),
		)

	// Shoot
	{
		actions.ApplyOnObject(shoot)
		shoot.Spec = gardenertypes.ShootSpec{
			CloudProfileName: ptr.To("azure"),
			Region:           "westeurope",
			Networking: &gardenertypes.Networking{
				IPFamilies: []gardenertypes.IPFamily{gardenertypes.IPFamilyIPv4},
				Nodes:      ptr.To("10.250.0.0/22"),
				Pods:       ptr.To("10.96.0.0/13"),
				Services:   ptr.To("10.104.0.0/13"),
			},
			Provider: gardenertypes.Provider{
				Type: string(cloudcontrolv1beta1.ProviderAzure),
				InfrastructureConfig: &runtime.RawExtension{
					Object: &azuregardener.InfrastructureConfig{
						TypeMeta: metav1.TypeMeta{
							Kind:       "InfrastructureConfig",
							APIVersion: "azure.provider.extensions.gardener.cloud/v1alpha1",
						},
						Networks: azuregardener.NetworkConfig{
							VNet: azuregardener.VNet{
								CIDR: ptr.To("10.250.0.0/22"),
							},
							Zones: []azuregardener.Zone{
								{
									Name: 2,
									CIDR: "10.250.0.0/25",
									NatGateway: &azuregardener.ZonedNatGatewayConfig{
										Enabled:                      true,
										IdleConnectionTimeoutMinutes: ptr.To(int32(4)),
									},
								},
								{
									Name: 3,
									CIDR: "10.250.0.128/25",
									NatGateway: &azuregardener.ZonedNatGatewayConfig{
										Enabled:                      true,
										IdleConnectionTimeoutMinutes: ptr.To(int32(4)),
									},
								},
								{
									Name: 1,
									CIDR: "10.250.1.0/25",
									NatGateway: &azuregardener.ZonedNatGatewayConfig{
										Enabled:                      true,
										IdleConnectionTimeoutMinutes: ptr.To(int32(4)),
									},
								},
							},
						},
						Zoned: true,
					},
				},
			},
			SecretBindingName: ptr.To(shoot.Name),
		}

		err := infra.Garden().Client().Create(ctx, shoot)
		if err != nil {
			return fmt.Errorf("error creating Shoot: %w", err)
		}
	}

	// SecretBinding
	{
		secretBinding := &gardenertypes.SecretBinding{}
		actions.ApplyOnObject(secretBinding)
		secretBinding.Provider = &gardenertypes.SecretBindingProvider{
			Type: string(cloudcontrolv1beta1.ProviderAzure),
		}
		secretBinding.SecretRef = corev1.SecretReference{
			Name:      shoot.Name,
			Namespace: shoot.Namespace,
		}
		err := infra.Garden().Client().Create(ctx, secretBinding)
		if err != nil {
			return fmt.Errorf("error creating SecretBinding: %w", err)
		}
	}

	// Secret
	{
		secret := &corev1.Secret{}
		actions.ApplyOnObject(secret)
		secret.StringData = map[string]string{
			"tenantID":       DefaultAzureTenantId,
			"subscriptionID": DefaultAzureSubscriptionId,
			"clientID":       "someAzureClientId",
			"clientSecret":   "someAzureClientSecret",
		}

		err := infra.Garden().Client().Create(ctx, secret)
		if err != nil {
			return fmt.Errorf("error creating garden secret: %w", err)
		}
	}

	return nil
}

func CreateShootSap(ctx context.Context, infra testinfra.Infra, shoot *gardenertypes.Shoot, opts ...ObjAction) error {
	// KCP Gardener-credentials secret
	if err := CreateGardenerCredentials(ctx, infra); err != nil {
		return err
	}

	// Garden resources
	if shoot == nil {
		shoot = &gardenertypes.Shoot{}
	}
	actions := NewObjActions(opts...).
		Append(
			WithNamespace(DefaultGardenNamespace),
		)

	// Shoot
	{
		actions.ApplyOnObject(shoot)
		shoot.Spec = gardenertypes.ShootSpec{
			CloudProfileName: ptr.To("converged-cloud-kyma"),
			Region:           "eu-de-1",
			Networking: &gardenertypes.Networking{
				IPFamilies: []gardenertypes.IPFamily{gardenertypes.IPFamilyIPv4},
				Nodes:      ptr.To("10.250.0.0/16"),
				Pods:       ptr.To("10.96.0.0/13"),
				Services:   ptr.To("10.104.0.0/13"),
			},
			Provider: gardenertypes.Provider{
				Type: string(cloudcontrolv1beta1.ProviderOpenStack),
				InfrastructureConfig: &runtime.RawExtension{
					Object: &sapgardener.InfrastructureConfig{
						TypeMeta: metav1.TypeMeta{
							Kind:       "InfrastructureConfig",
							APIVersion: "openstack.provider.extensions.gardener.cloud/v1alpha1",
						},
						FloatingPoolName: "FloatingIP-external-kyma-01",
						Networks: sapgardener.Networks{
							Workers: "10.250.0.0/16",
						},
					},
				},
				Workers: []gardenertypes.Worker{
					{
						Name:  "cpu-worker-0",
						Zones: []string{"eu-de-1b", "eu-de-1a", "eu-de-1c"},
					},
				},
			},
			SecretBindingName: ptr.To(shoot.Name),
		}

		err := infra.Garden().Client().Create(ctx, shoot)
		if err != nil {
			return fmt.Errorf("error creating Shoot: %w", err)
		}
	}

	// SecretBinding
	{
		secretBinding := &gardenertypes.SecretBinding{}
		actions.ApplyOnObject(secretBinding)
		secretBinding.Provider = &gardenertypes.SecretBindingProvider{
			Type: string(cloudcontrolv1beta1.ProviderOpenStack),
		}
		secretBinding.SecretRef = corev1.SecretReference{
			Name:      shoot.Name,
			Namespace: shoot.Namespace,
		}
		err := infra.Garden().Client().Create(ctx, secretBinding)
		if err != nil {
			return fmt.Errorf("error creating SecretBinding: %w", err)
		}
	}

	// Secret
	{
		secret := &corev1.Secret{}
		actions.ApplyOnObject(secret)
		secret.StringData = map[string]string{
			"domainName": DefaultSapDomain,
			"tenantName": DefaultSapTenant,
		}

		err := infra.Garden().Client().Create(ctx, secret)
		if err != nil {
			return fmt.Errorf("error creating garden secret: %w", err)
		}
	}

	return nil
}

var (
	DefaultGcpWorkerZones = []string{"europe-west1-a", "europe-west1-b", "europe-west1-c"}
)

const (
	DefaultGcpProject = "project_id"

	DefaultAzureTenantId       = "someAzureTenantId"
	DefaultAzureSubscriptionId = "someAzureSubscriptionId"

	DefaultSapDomain = "kyma"
	DefaultSapTenant = "kyma-project-01"
)
