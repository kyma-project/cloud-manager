package e2e

import (
	"context"
	"fmt"
	"time"

	"github.com/elliotchance/pie/v2"
	gardenertypes "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/external/keb"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
)

type SkrCreator struct {
	world *World
}

func NewSkrCreator(world *World) *SkrCreator {
	return &SkrCreator{
		world: world,
	}
}

type CreateSkrInput struct {
	Provider     cloudcontrolv1beta1.ProviderType
	Subscription *SubscriptionInfo
	Region       string
}

type CreateSkrOutput struct {
	Runtime *keb.Runtime
	Shoot   *gardenertypes.Shoot
}

func (c *SkrCreator) CreateSkr(ctx context.Context, in CreateSkrInput) (*CreateSkrOutput, error) {
	if in.Subscription == nil {
		return nil, fmt.Errorf("subscription is required")
	}

	garden, err := c.world.ClusterProvider.Garden(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not create garden cluster: %w", err)
	}
	secretBinding := &gardenertypes.SecretBinding{}
	err = garden.Cluster.GetAPIReader().Get(ctx, types.NamespacedName{
		Name:      in.Subscription.Name,
		Namespace: Config.GardenNamespace,
	}, secretBinding)
	if err != nil {
		return nil, fmt.Errorf("could not load secret bindings %q: %w", in.Subscription.Name, err)
	}

	if in.Region == "" {
		in.Region = pie.Keys(providerRegions[in.Provider])[0]
	}

	kcp, err := c.world.ClusterProvider.KCP(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not create kcp cluster: %w", err)
	}

	runtimeBuilder := NewRuntimeBuilder().
		WithProvider(in.Provider, in.Region).
		WithSecretBindingName(secretBinding.Name)
	if err := runtimeBuilder.Validate(); err != nil {
		return nil, fmt.Errorf("invalid runtime: %w", err)
	}
	runtime := runtimeBuilder.Build()
	err = kcp.Cluster.GetClient().Create(ctx, runtime)
	if err != nil {
		return nil, fmt.Errorf("error creating runtime: %w", err)
	}

	// Shoot.core.gardener.cloud "c-f8ade" is invalid: [metadata.labels: Invalid value: "provider.extensions.gardener.cloud/":
	// name part must be non-empty, metadata.labels: Invalid value: "provider.extensions.gardener.cloud/": name part must consist
	// of alphanumeric characters, '-', '_' or '.', and must start and end with an alphanumeric character (e.g. 'MyName',  or 'my.name',
	// or '123-abc', regex used for validation is '([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9]'), spec.cloudProfile.name: Required value:
	// must specify a cloud profile, spec.provider.type: Required value: must specify a provider type, spec.kubernetes.version:
	// Required value: kubernetes version must not be empty, spec.region: Required value: must specify a region]
	shootBuilder := NewShootBuilder().
		WithRuntime(runtime)
	if err := shootBuilder.Validate(); err != nil {
		return nil, fmt.Errorf("invalid shoot: %w", err)
	}
	shoot := shootBuilder.Build()
	err = garden.Cluster.GetClient().Create(ctx, shoot)
	if err != nil {
		return nil, fmt.Errorf("error creating shoot: %w", err)
	}

	_ = wait.PollUntilContextCancel(ctx, 5*time.Second, true, func(ctx context.Context) (bool, error) {
		return true, nil
	})

	return &CreateSkrOutput{}, nil
}
