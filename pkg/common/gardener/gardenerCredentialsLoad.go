package gardener

import (
	"context"
	"errors"
	"fmt"

	gardenerclient "github.com/gardener/gardener/pkg/client/core/clientset/versioned/typed/core/v1beta1"
	"github.com/hashicorp/go-multierror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubernetesclient "k8s.io/client-go/kubernetes"
)

type LoadGardenerCloudProviderCredentialsInput struct {
	GardenerClient  gardenerclient.CoreV1beta1Interface
	GardenK8sClient kubernetesclient.Interface
	Namespace       string
	BindingName     string
}

func (in LoadGardenerCloudProviderCredentialsInput) Validate() error {
	var result error
	if len(in.BindingName) == 0 {
		result = multierror.Append(result, errors.New("binding name is required"))
	}
	if len(in.Namespace) == 0 {
		result = multierror.Append(result, errors.New("namespace is required"))
	}
	if in.GardenK8sClient == nil {
		result = multierror.Append(result, errors.New("gardenK8sClient is required"))
	}
	if in.GardenK8sClient == nil {
		result = multierror.Append(result, errors.New("gardenK8sClient is required"))
	}
	return result
}

type LoadGardenerCloudProviderCredentialsOutput struct {
	Provider        string
	SecretName      string
	SecretNamespace string
	CredentialsData map[string]string
}

func LoadGardenerCloudProviderCredentials(ctx context.Context, in LoadGardenerCloudProviderCredentialsInput) (*LoadGardenerCloudProviderCredentialsOutput, error) {
	if err := in.Validate(); err != nil {
		return nil, err
	}

	out := &LoadGardenerCloudProviderCredentialsOutput{
		CredentialsData: map[string]string{},
	}

	secretBinding, err := in.GardenerClient.SecretBindings(in.Namespace).Get(ctx, in.BindingName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error loading secret binding: %w", err)
	}
	out.Provider = secretBinding.Provider.Type
	out.SecretName = secretBinding.SecretRef.Name
	ns := secretBinding.SecretRef.Namespace
	if ns == "" {
		ns = in.Namespace
	}
	out.SecretNamespace = ns

	secret, err := in.GardenK8sClient.CoreV1().Secrets(out.SecretNamespace).
		Get(ctx, out.SecretName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error loading shoot secret: %w", err)
	}

	for k, v := range secret.Data {
		out.CredentialsData[k] = string(v)
	}

	return out, nil
}
