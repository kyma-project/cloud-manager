package lib

import (
	"context"
	"fmt"
	"time"

	authenticationv1alpha1 "github.com/gardener/gardener/pkg/apis/authentication/v1alpha1"
	gardenertypes "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SkrKubeconfigProvider interface {
	CreateNewKubeconfig(ctx context.Context, shootName string) ([]byte, error)
	ExpiresIn() time.Duration
}

// GARDEN SkrKubeconfigProvider =================================================

func NewGardenSkrKubeconfigProvider(gardenClient client.Client, expiresIn time.Duration, gardenNamespace string) SkrKubeconfigProvider {
	return &gardenKubeconfigProvider{
		gardenClient:    gardenClient,
		expiresIn:       expiresIn,
		gardenNamespace: gardenNamespace,
	}
}

type gardenKubeconfigProvider struct {
	gardenClient    client.Client
	expiresIn       time.Duration
	gardenNamespace string
}

func (p *gardenKubeconfigProvider) CreateNewKubeconfig(ctx context.Context, shootName string) ([]byte, error) {
	shoot := &gardenertypes.Shoot{}
	err := p.gardenClient.Get(ctx, types.NamespacedName{
		Namespace: p.gardenNamespace,
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
	err = p.gardenClient.SubResource("adminkubeconfig").Create(ctx, shoot, adminKubeconfigRequest)
	if err != nil {
		return nil, fmt.Errorf("error creating admin kubeconfig: %w", err)
	}
	return adminKubeconfigRequest.Status.Kubeconfig, nil
}

func (p *gardenKubeconfigProvider) ExpiresIn() time.Duration {
	return p.expiresIn
}

// FIXED SkrKubeconfigProvider =================================================

type SkrKubeconfigProviderWithCallCount interface {
	SkrKubeconfigProvider
	GetCallCount(shootName string) int
}

func NewFixedSkrKubeconfigProvider(kubeconfig []byte) SkrKubeconfigProvider {
	return &fixedKubeconfigProvider{
		kubeconfig: kubeconfig,
		callCount:  make(map[string]int),
	}
}

type fixedKubeconfigProvider struct {
	kubeconfig []byte
	callCount  map[string]int
}

func (p *fixedKubeconfigProvider) GetCallCount(shootName string) int {
	return p.callCount[shootName]
}

func (p *fixedKubeconfigProvider) CreateNewKubeconfig(ctx context.Context, shootName string) ([]byte, error) {
	p.callCount[shootName] = p.callCount[shootName] + 1
	return p.kubeconfig, nil
}

func (p *fixedKubeconfigProvider) ExpiresIn() time.Duration {
	return 10 * time.Hour
}
