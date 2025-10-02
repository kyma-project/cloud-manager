package gardener

import (
	"context"
	"errors"
	"fmt"

	gardenerapicore "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	gardenerapisecurity "github.com/gardener/gardener/pkg/apis/security/v1alpha1"
	"github.com/hashicorp/go-multierror"
	"github.com/kyma-project/cloud-manager/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type LoadGardenerCloudProviderCredentialsInput struct {
	Client      client.Client
	Namespace   string
	BindingName string
}

func (in LoadGardenerCloudProviderCredentialsInput) Validate() error {
	var result error
	if len(in.BindingName) == 0 {
		result = multierror.Append(result, errors.New("binding name is required"))
	}
	if len(in.Namespace) == 0 {
		result = multierror.Append(result, errors.New("namespace is required"))
	}
	if in.Client == nil {
		result = multierror.Append(result, errors.New("garden client is required"))
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

	credentialBinding := &gardenerapisecurity.CredentialsBinding{}
	err := in.Client.Get(ctx, types.NamespacedName{Namespace: in.Namespace, Name: in.BindingName}, credentialBinding)
	if util.IgnoreNoMatch(client.IgnoreNotFound(err)) != nil {
		return nil, fmt.Errorf("error loading credential binding: %w", err)
	}
	if err == nil {
		if credentialBinding.CredentialsRef.Kind != "Secret" {
			return nil, fmt.Errorf("unsupported CredentialsBinding credentialsRef kind: %s/%s", credentialBinding.CredentialsRef.APIVersion, credentialBinding.CredentialsRef.Kind)
		}
		out.Provider = credentialBinding.Provider.Type
		out.SecretName = credentialBinding.CredentialsRef.Name
		out.SecretNamespace = credentialBinding.CredentialsRef.Namespace
	}

	if out.SecretName == "" {
		// Fallback to SecretBinding
		//lint:ignore SA1019 support until SecretBinding is migrated to CredentialsBinding
		secretBinding := &gardenerapicore.SecretBinding{}
		err = in.Client.Get(ctx, types.NamespacedName{Namespace: in.Namespace, Name: in.BindingName}, secretBinding)
		if err != nil {
			return nil, fmt.Errorf("error loading secret binding: %w", err)
		}
		out.Provider = secretBinding.Provider.Type
		out.SecretName = secretBinding.SecretRef.Name
		out.SecretNamespace = secretBinding.SecretRef.Namespace
	}

	if out.SecretNamespace == "" {
		out.SecretNamespace = in.Namespace
	}

	secret := &corev1.Secret{}
	err = in.Client.Get(ctx, types.NamespacedName{Namespace: out.SecretNamespace, Name: out.SecretName}, secret)
	if err != nil {
		return nil, fmt.Errorf("error loading shoot secret: %w", err)
	}

	for k, v := range secret.Data {
		out.CredentialsData[k] = string(v)
	}

	return out, nil
}
