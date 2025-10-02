package gardener

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/kyma-project/cloud-manager/pkg/common/bootstrap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
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
	Namespace  string
	RestConfig *rest.Config
	Client     client.Client
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

	clnt, err := client.New(restConfig, client.Options{
		Scheme: bootstrap.GardenScheme,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating gardener client: %w", err)
	}
	out.Client = clnt

	return out, nil
}
