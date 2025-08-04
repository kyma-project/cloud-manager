package e2e

import (
	"context"
	"fmt"

	gardenertypes "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/google/uuid"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/external/keb"
	"github.com/kyma-project/cloud-manager/pkg/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ SkrCreator = &skrCreatorGardenerNetwork{}

type skrCreatorGardenerNetwork struct {
	world        *World
	subscription *SubscriptionInfo
}

func (c *skrCreatorGardenerNetwork) CreateSkr(ctx context.Context, provider cloudcontrolv1beta1.ProviderType) (*SkrInfo, error) {
	garden, err := c.world.ClusterProvider.Garden(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not create garden cluster: %w", err)
	}
	sb := &gardenertypes.SecretBinding{}
	err = garden.Cluster.GetAPIReader().Get(ctx, types.NamespacedName{
		Name:      c.subscription.Name,
		Namespace: c.world.GardenNamespace(),
	}, sb)
	if err != nil {
		return nil, fmt.Errorf("could not load secret bindings %q: %w", c.subscription.Name, err)
	}

	shootId := util.RandomId(7)
	var region string
	switch provider {
	case cloudcontrolv1beta1.ProviderAzure:
		region = "westeurope"
	case cloudcontrolv1beta1.ProviderGCP:
		region = "europe-west3"
	case cloudcontrolv1beta1.ProviderAws:
		region = "us-east-1"
	case cloudcontrolv1beta1.ProviderOpenStack:
		region = "eu-de-1"
	default:
		return nil, fmt.Errorf("can not determine region for provider: %s", provider)
	}

	runtime := &keb.Runtime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      uuid.NewString(),
			Namespace: c.world.GardenNamespace(),
			Labels: map[string]string{
				cloudcontrolv1beta1.LabelScopeShootName: shootId,
				cloudcontrolv1beta1.LabelScopeRegion:    region,
			},
		},
		Spec: keb.RuntimeSpec{
			Shoot: keb.RuntimeShoot{
				Name:                shootId,
				Purpose:             "test",
				PlatformRegion:      region,
				Region:              region,
				SecretBindingName:   sb.Name,
				EnforceSeedLocation: nil,
				Kubernetes: keb.Kubernetes{
					Version: ptr.To("1.33.0"),
					KubeAPIServer: keb.APIServer{
						OidcConfig: gardenertypes.OIDCConfig{
							ClientID:       ptr.To("TODO"),
							GroupsClaim:    ptr.To("groups"),
							IssuerURL:      ptr.To("https://todo.com"),
							UsernameClaim:  ptr.To("sub"),
							UsernamePrefix: ptr.To("-"),
						},
					},
				},
				Provider: keb.Provider{
					Type: string(provider),
					Workers: []gardenertypes.Worker{},
				},
				Networking:   keb.Networking{},
				ControlPlane: nil,
			},
		},
	}
}
