package dsl

import (
	"context"
	"fmt"
	gardenerTypes "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	awsgardener "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/gardener"
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

func CreateShootAws(ctx context.Context, infra testinfra.Infra, shoot *gardenerTypes.Shoot, opts ...ObjAction) error {
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
		shoot = &gardenerTypes.Shoot{}
	}
	actions := NewObjActions(opts...).
		Append(
			WithNamespace(DefaultGardenNamespace),
		)

	// Shoot
	{
		actions.ApplyOnObject(shoot)
		shoot.Spec = gardenerTypes.ShootSpec{
			Region: "eu-west-1",
			Networking: &gardenerTypes.Networking{
				IPFamilies: []gardenerTypes.IPFamily{gardenerTypes.IPFamilyIPv4},
				Nodes:      ptr.To("10.180.0.0/16"),
				Pods:       ptr.To("100.64.0.0/12"),
				Services:   ptr.To("100.104.0.0/13"),
			},
			Provider: gardenerTypes.Provider{
				Type: "aws",
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
		secretBinding := &gardenerTypes.SecretBinding{}
		actions.ApplyOnObject(secretBinding)
		secretBinding.Provider = &gardenerTypes.SecretBindingProvider{
			Type: "aws",
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

func CreateShootGcp(ctx context.Context, infra testinfra.Infra, shoot *gardenerTypes.Shoot, opts ...ObjAction) error {
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
		shoot = &gardenerTypes.Shoot{}
	}
	actions := NewObjActions(opts...).
		Append(
			WithNamespace(DefaultGardenNamespace),
		)

	// Shoot
	{
		actions.ApplyOnObject(shoot)
		shoot.Spec = gardenerTypes.ShootSpec{
			Region:     "eu-west-1",
			Networking: &gardenerTypes.Networking{},
			Provider: gardenerTypes.Provider{
				Type: "gcp",
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
		secretBinding := &gardenerTypes.SecretBinding{}
		actions.ApplyOnObject(secretBinding)
		secretBinding.Provider = &gardenerTypes.SecretBindingProvider{
			Type: "gcp",
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
			"serviceaccount.json": "{" +
				"\"project_id\": \"project_id\"}",
		}
		err := infra.Garden().Client().Create(ctx, secret)
		if err != nil {
			return fmt.Errorf("error creating garden secret: %w", err)
		}
	}

	return nil
}
