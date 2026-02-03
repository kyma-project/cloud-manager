package dsl

import (
	"context"
	"fmt"

	gardenerawsapi "github.com/gardener/gardener-extension-provider-aws/pkg/apis/aws/v1alpha1"
	gardenerazureapi "github.com/gardener/gardener-extension-provider-azure/pkg/apis/azure/v1alpha1"
	gardeneropenstackapi "github.com/gardener/gardener-extension-provider-openstack/pkg/apis/openstack/v1alpha1"
	gardenerapicore "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	gardenerapisecurity "github.com/gardener/gardener/pkg/apis/security/v1alpha1"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	awsmock "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/mock"
	sapclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/client"
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

func CreateShootAws(ctx context.Context, infra testinfra.Infra, shoot *gardenerapicore.Shoot, awsCredentials awsmock.AccountCredential, opts ...ObjAction) error {
	// KCP Gardener-credentials secret
	if err := CreateGardenerCredentials(ctx, infra); err != nil {
		return err
	}

	// Garden resources
	if shoot == nil {
		shoot = &gardenerapicore.Shoot{}
	}
	actions := NewObjActions(opts...).
		Append(
			WithNamespace(DefaultGardenNamespace),
		)

	// Shoot
	{
		actions.ApplyOnObject(shoot)
		shoot.Spec = gardenerapicore.ShootSpec{
			CredentialsBindingName: ptr.To(shoot.Name),
			CloudProfile: &gardenerapicore.CloudProfileReference{
				Kind: "CloudProfile",
				Name: "aws",
			},
			Region: "eu-west-1",
			Networking: &gardenerapicore.Networking{
				IPFamilies: []gardenerapicore.IPFamily{gardenerapicore.IPFamilyIPv4},
				Nodes:      ptr.To("10.180.0.0/16"),
				Pods:       ptr.To("100.64.0.0/12"),
				Services:   ptr.To("100.104.0.0/13"),
			},
			Provider: gardenerapicore.Provider{
				Type: string(cloudcontrolv1beta1.ProviderAws),
				InfrastructureConfig: &runtime.RawExtension{
					Object: &gardenerawsapi.InfrastructureConfig{
						TypeMeta: metav1.TypeMeta{
							Kind:       "InfrastructureConfig",
							APIVersion: "aws.provider.extensions.gardener.cloud/v1alpha1",
						},
						Networks: gardenerawsapi.Networks{
							VPC: gardenerawsapi.VPC{
								CIDR: ptr.To("10.180.0.0/16"),
							},
							Zones: []gardenerawsapi.Zone{
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
		}

		err := infra.Garden().Client().Create(ctx, shoot)
		if err != nil {
			return fmt.Errorf("error creating Shoot: %w", err)
		}
	}

	credentialsBinding := &gardenerapisecurity.CredentialsBinding{}
	actions.ApplyOnObject(credentialsBinding)
	credentialsBinding.Provider.Type = string(cloudcontrolv1beta1.ProviderAws)
	credentialsBinding.CredentialsRef = corev1.ObjectReference{
		Kind:       "Secret",
		Namespace:  shoot.Namespace,
		Name:       shoot.Name,
		APIVersion: "v1",
	}
	err := infra.Garden().Client().Create(ctx, credentialsBinding)
	if err != nil {
		return fmt.Errorf("error creating CredentialsBinding: %w", err)
	}

	// Secret
	{
		secret := &corev1.Secret{}
		actions.ApplyOnObject(secret)
		secret.StringData = map[string]string{
			"accessKeyID":     awsCredentials.AccessKeyId,
			"secretAccessKey": awsCredentials.SecretAccessKey,
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

func CreateShootGcp(ctx context.Context, infra testinfra.Infra, shoot *gardenerapicore.Shoot, opts ...ObjAction) error {
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
		shoot = &gardenerapicore.Shoot{}
	}
	actions := NewObjActions(opts...).
		Append(
			WithNamespace(DefaultGardenNamespace),
		)

	// Shoot
	{
		actions.ApplyOnObject(shoot)
		shoot.Spec = gardenerapicore.ShootSpec{
			CredentialsBindingName: ptr.To(shoot.Name),
			CloudProfile: &gardenerapicore.CloudProfileReference{
				Kind: "CloudProfile",
				Name: "gcp",
			},
			Region:     "eu-west-1",
			Networking: &gardenerapicore.Networking{},
			Provider: gardenerapicore.Provider{
				Type: string(cloudcontrolv1beta1.ProviderGCP),
				Workers: []gardenerapicore.Worker{
					{
						Zones: DefaultGcpWorkerZones,
					},
				},
			},
		}

		err := infra.Garden().Client().Create(ctx, shoot)
		if err != nil {
			return fmt.Errorf("error creating Shoot: %w", err)
		}
	}

	// credentialsBinding
	{
		credentialsBinding := &gardenerapisecurity.CredentialsBinding{}
		actions.ApplyOnObject(credentialsBinding)
		credentialsBinding.Provider.Type = string(cloudcontrolv1beta1.ProviderGCP)
		credentialsBinding.CredentialsRef = corev1.ObjectReference{
			Kind:       "Secret",
			Namespace:  shoot.Namespace,
			Name:       shoot.Name,
			APIVersion: "v1",
		}
		err := infra.Garden().Client().Create(ctx, credentialsBinding)
		if err != nil {
			return fmt.Errorf("error creating CredentialsBinding: %w", err)
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

func CreateShootAzure(ctx context.Context, infra testinfra.Infra, shoot *gardenerapicore.Shoot, tenant string, subscription string, region string, opts ...ObjAction) error {
	if region == "" {
		region = "westeurope"
	}

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
		shoot = &gardenerapicore.Shoot{}
	}
	actions := NewObjActions(opts...).
		Append(
			WithNamespace(DefaultGardenNamespace),
		)

	// Shoot
	{
		actions.ApplyOnObject(shoot)
		shoot.Spec = gardenerapicore.ShootSpec{
			CredentialsBindingName: ptr.To(shoot.Name),
			CloudProfile: &gardenerapicore.CloudProfileReference{
				Kind: "CloudProfile",
				Name: "azure",
			},
			Region: region,
			Networking: &gardenerapicore.Networking{
				IPFamilies: []gardenerapicore.IPFamily{gardenerapicore.IPFamilyIPv4},
				Nodes:      ptr.To("10.250.0.0/22"),
				Pods:       ptr.To("10.96.0.0/13"),
				Services:   ptr.To("10.104.0.0/13"),
			},
			Provider: gardenerapicore.Provider{
				Type: string(cloudcontrolv1beta1.ProviderAzure),
				InfrastructureConfig: &runtime.RawExtension{
					Object: &gardenerazureapi.InfrastructureConfig{
						TypeMeta: metav1.TypeMeta{
							Kind:       "InfrastructureConfig",
							APIVersion: "azure.provider.extensions.gardener.cloud/v1alpha1",
						},
						Networks: gardenerazureapi.NetworkConfig{
							VNet: gardenerazureapi.VNet{
								CIDR: ptr.To("10.250.0.0/22"),
							},
							Zones: []gardenerazureapi.Zone{
								{
									Name: 2,
									CIDR: "10.250.0.0/25",
									NatGateway: &gardenerazureapi.ZonedNatGatewayConfig{
										Enabled:                      true,
										IdleConnectionTimeoutMinutes: ptr.To(int32(4)),
									},
								},
								{
									Name: 3,
									CIDR: "10.250.0.128/25",
									NatGateway: &gardenerazureapi.ZonedNatGatewayConfig{
										Enabled:                      true,
										IdleConnectionTimeoutMinutes: ptr.To(int32(4)),
									},
								},
								{
									Name: 1,
									CIDR: "10.250.1.0/25",
									NatGateway: &gardenerazureapi.ZonedNatGatewayConfig{
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
		}

		err := infra.Garden().Client().Create(ctx, shoot)
		if err != nil {
			return fmt.Errorf("error creating Shoot: %w", err)
		}
	}

	// credentialsBinding
	{
		credentialsBinding := &gardenerapisecurity.CredentialsBinding{}
		actions.ApplyOnObject(credentialsBinding)
		credentialsBinding.Provider.Type = string(cloudcontrolv1beta1.ProviderAzure)
		credentialsBinding.CredentialsRef = corev1.ObjectReference{
			Kind:       "Secret",
			Namespace:  shoot.Namespace,
			Name:       shoot.Name,
			APIVersion: "v1",
		}
		err := infra.Garden().Client().Create(ctx, credentialsBinding)
		if err != nil {
			return fmt.Errorf("error creating CredentialsBinding: %w", err)
		}
	}

	// Secret
	{
		secret := &corev1.Secret{}
		actions.ApplyOnObject(secret)
		secret.StringData = map[string]string{
			"tenantID":       tenant,
			"subscriptionID": subscription,
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

func CreateShootSap(ctx context.Context, infra testinfra.Infra, shoot *gardenerapicore.Shoot, pp sapclient.ProviderParams, opts ...ObjAction) error {
	// KCP Gardener-credentials secret
	if err := CreateGardenerCredentials(ctx, infra); err != nil {
		return err
	}

	// Garden resources
	if shoot == nil {
		shoot = &gardenerapicore.Shoot{}
	}
	actions := NewObjActions(opts...).
		Append(
			WithNamespace(DefaultGardenNamespace),
		)

	// Shoot
	{
		actions.ApplyOnObject(shoot)
		shoot.Spec = gardenerapicore.ShootSpec{
			CredentialsBindingName: ptr.To(shoot.Name),
			CloudProfile: &gardenerapicore.CloudProfileReference{
				Kind: "CloudProfile",
				Name: "converged-cloud-kyma",
			},
			Region: pp.RegionName,
			Networking: &gardenerapicore.Networking{
				IPFamilies: []gardenerapicore.IPFamily{gardenerapicore.IPFamilyIPv4},
				Nodes:      ptr.To("10.250.0.0/16"),
				Pods:       ptr.To("10.96.0.0/13"),
				Services:   ptr.To("10.104.0.0/13"),
			},
			Provider: gardenerapicore.Provider{
				Type: string(cloudcontrolv1beta1.ProviderOpenStack),
				InfrastructureConfig: &runtime.RawExtension{
					Object: &gardeneropenstackapi.InfrastructureConfig{
						TypeMeta: metav1.TypeMeta{
							Kind:       "InfrastructureConfig",
							APIVersion: "openstack.provider.extensions.gardener.cloud/v1alpha1",
						},
						FloatingPoolName: "FloatingIP-external-kyma-01",
						Networks: gardeneropenstackapi.Networks{
							Workers: "10.250.0.0/16",
						},
					},
				},
				Workers: []gardenerapicore.Worker{
					{
						Name:  "cpu-worker-0",
						Zones: []string{pp.RegionName + "b", pp.RegionName + "a", pp.RegionName + "c"},
					},
				},
			},
		}

		err := infra.Garden().Client().Create(ctx, shoot)
		if err != nil {
			return fmt.Errorf("error creating Shoot: %w", err)
		}
	}

	// credentialsBinding
	{
		credentialsBinding := &gardenerapisecurity.CredentialsBinding{}
		actions.ApplyOnObject(credentialsBinding)
		credentialsBinding.Provider.Type = string(cloudcontrolv1beta1.ProviderOpenStack)
		credentialsBinding.CredentialsRef = corev1.ObjectReference{
			Kind:       "Secret",
			Namespace:  shoot.Namespace,
			Name:       shoot.Name,
			APIVersion: "v1",
		}
		err := infra.Garden().Client().Create(ctx, credentialsBinding)
		if err != nil {
			return fmt.Errorf("error creating CredentialsBinding: %w", err)
		}
	}

	// Secret
	{
		secret := &corev1.Secret{}
		actions.ApplyOnObject(secret)
		secret.StringData = map[string]string{
			"domainName": pp.DomainName,
			"tenantName": pp.ProjectName,
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
)
