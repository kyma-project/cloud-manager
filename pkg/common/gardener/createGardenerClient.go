package gardener

import (
	"context"
	"errors"
	"fmt"

	gardenerclient "github.com/gardener/gardener/pkg/client/core/clientset/versioned/typed/core/v1beta1"
	"github.com/hashicorp/go-multierror"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	kubernetesclient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CreateGardenerClientInput struct {
	KcpClient                 client.Reader
	GardenerFallbackNamespace string
}

func (in CreateGardenerClientInput) Validate() error {
	var result error
	if in.KcpClient == nil {
		result = multierror.Append(result, errors.New("kcp client is required"))
	}
	return result
}

type CreateGardenerClientOutput struct {
	Namespace       string
	RestConfig      *rest.Config
	GardenerClient  gardenerclient.CoreV1beta1Interface
	GardenK8sClient kubernetesclient.Interface
}

func CreateGardenerClient(ctx context.Context, in CreateGardenerClientInput) (*CreateGardenerClientOutput, error) {
	if err := in.Validate(); err != nil {
		return nil, err
	}

	secret := &corev1.Secret{}
	err := in.KcpClient.Get(ctx, types.NamespacedName{
		Namespace: "kcp-system",
		Name:      "gardener-credentials",
	}, secret)
	if err != nil {
		return nil, fmt.Errorf("error getting gardener credentials: %w", err)
	}

	kubeBytes, ok := secret.Data["kubeconfig"]
	if !ok {
		return nil, fmt.Errorf("gardener credentials missing kubeconfig key")
	}

	config, err := clientcmd.NewClientConfigFromBytes(kubeBytes)
	if err != nil {
		return nil, fmt.Errorf("error creating gardener client config: %w", err)
	}

	rawConfig, err := config.RawConfig()
	if err != nil {
		return nil, fmt.Errorf("error getting gardener raw client config: %w", err)
	}

	out := &CreateGardenerClientOutput{}

	var configContext *clientcmdapi.Context
	if len(rawConfig.CurrentContext) > 0 {
		configContext = rawConfig.Contexts[rawConfig.CurrentContext]
	} else {
		for _, c := range rawConfig.Contexts {
			configContext = c
			break
		}
	}
	if configContext != nil && len(configContext.Namespace) > 0 {
		out.Namespace = configContext.Namespace
	} else {
		out.Namespace = in.GardenerFallbackNamespace
	}

	restConfig, err := clientcmd.RESTConfigFromKubeConfig(kubeBytes)
	if err != nil {
		return nil, fmt.Errorf("error creating gardener rest client config: %w", err)
	}
	out.RestConfig = restConfig

	gClient, err := gardenerclient.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating gardener client: %w", err)
	}
	out.GardenerClient = gClient

	k8sClient, err := kubernetesclient.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating gardener k8s client: %w", err)
	}
	out.GardenK8sClient = k8sClient

	return out, nil
}
