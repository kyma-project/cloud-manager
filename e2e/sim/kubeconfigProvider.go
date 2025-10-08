package sim

import (
	"context"
	"fmt"
	"time"

	authenticationv1alpha1 "github.com/gardener/gardener/pkg/apis/authentication/v1alpha1"
	gardenertypes "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	e2econfig "github.com/kyma-project/cloud-manager/e2e/config"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type KubeconfigProvider interface {
	CreateNewKubeconfig(ctx context.Context, shootName string) ([]byte, error)
	ExpiresIn() time.Duration
}

func NewKubeconfigProvider(garden client.Client, expiresIn time.Duration) KubeconfigProvider {
	return &defaultKubeconfigProvider{
		garden:    garden,
		expiresIn: expiresIn,
	}
}

type defaultKubeconfigProvider struct {
	garden    client.Client
	expiresIn time.Duration
}

func (p *defaultKubeconfigProvider) CreateNewKubeconfig(ctx context.Context, shootName string) ([]byte, error) {
	shoot := &gardenertypes.Shoot{}
	err := p.garden.Get(ctx, types.NamespacedName{
		Namespace: e2econfig.Config.GardenNamespace,
		Name:      shootName,
	}, shoot)
	if err != nil {
		return nil, fmt.Errorf("error getting shoot: %w", err)
	}

	adminKubeconfigRequest := &authenticationv1alpha1.AdminKubeconfigRequest{
		Spec: authenticationv1alpha1.AdminKubeconfigRequestSpec{
			ExpirationSeconds: ptr.To(int64(p.expiresIn.Seconds())),
		},
	}
	err = p.garden.SubResource("adminkubeconfig").Create(ctx, shoot, adminKubeconfigRequest)
	if err != nil {
		return nil, fmt.Errorf("error creating admin kubeconfig: %w", err)
	}
	return adminKubeconfigRequest.Status.Kubeconfig, nil
}

func (p *defaultKubeconfigProvider) ExpiresIn() time.Duration {
	return p.expiresIn
}
