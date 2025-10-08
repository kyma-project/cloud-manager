package sim

import (
	"context"
	"fmt"

	e2econfig "github.com/kyma-project/cloud-manager/e2e/config"
	"github.com/kyma-project/cloud-manager/pkg/common/bootstrap"
	"github.com/kyma-project/cloud-manager/pkg/external/infrastructuremanagerv1"
	skrmanager "github.com/kyma-project/cloud-manager/pkg/skr/runtime/manager"
	"github.com/kyma-project/cloud-manager/pkg/util"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/rest"
	"k8s.io/utils/clock"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
)

// ClientClusterFactory creates clients and clusters for SKR runtimes. It uses the KCP client to fetch the GardenerCluster
// resource and the kubeconfig secret referenced in it to create the client/cluster.
type ClientClusterFactory interface {
	CreateClient(ctx context.Context, runtimeID string) (client.Client, error)
	CreateClientCluster(ctx context.Context, runtimeID string) (cluster.Cluster, error)
}

func NewClientClusterFactory(kcp client.Client, clock clock.Clock) ClientClusterFactory {
	return &defaultClientClusterFactory{
		clock: clock,
		kcp:   kcp,
	}
}

var ErrGardenerClusterCredentialsExpired = fmt.Errorf("gardenercluster credentials expired")

var _ ClientClusterFactory = &defaultClientClusterFactory{}

type defaultClientClusterFactory struct {
	clock clock.Clock
	kcp   client.Client
}

func (f *defaultClientClusterFactory) CreateClient(ctx context.Context, runtimeID string) (client.Client, error) {
	restConfig, err := f.getRestConfig(ctx, runtimeID)
	if err != nil {
		return nil, fmt.Errorf("error getting skr rest config for client: %w", err)
	}

	clnt, err := client.New(restConfig, client.Options{
		Scheme: bootstrap.SkrScheme,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating client: %w", err)
	}

	return clnt, nil
}

func (f *defaultClientClusterFactory) CreateClientCluster(ctx context.Context, runtimeID string) (cluster.Cluster, error) {
	restConfig, err := f.getRestConfig(ctx, runtimeID)
	if err != nil {
		return nil, fmt.Errorf("error getting skr rest config for cluster: %w", err)
	}

	clstr, err := cluster.New(restConfig, func(clusterOptions *cluster.Options) {
		clusterOptions.Scheme = bootstrap.SkrScheme
		clusterOptions.Client = client.Options{
			Cache: &client.CacheOptions{
				Unstructured: true,
			},
		}
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create SKR cluster: %w", err)
	}

	return clstr, nil
}

func (f *defaultClientClusterFactory) getRestConfig(ctx context.Context, runtimeID string) (*rest.Config, error) {
	gc := &infrastructuremanagerv1.GardenerCluster{}
	err := f.kcp.Get(ctx, client.ObjectKey{Namespace: e2econfig.Config.KcpNamespace, Name: runtimeID}, gc)
	if client.IgnoreNotFound(err) != nil {
		return nil, fmt.Errorf("error getting GardenerCluster object: %w", err)
	}
	if apierrors.IsNotFound(err) {
		return nil, fmt.Errorf("GardenerCluster %q not found", runtimeID)
	}

	hasExpired, _ := IsGardenerClusterSyncNeeded(gc, f.clock)
	if hasExpired {
		return nil, ErrGardenerClusterCredentialsExpired
	}

	gcSummary := &util.GardenerClusterSummary{
		Key:       gc.Spec.Kubeconfig.Secret.Key,
		Name:      gc.Spec.Kubeconfig.Secret.Name,
		Namespace: gc.Spec.Kubeconfig.Secret.Namespace,
		Shoot:     gc.Spec.Shoot.Name,
	}

	skrManagerFactory := skrmanager.NewFactory(f.kcp, gcSummary.Namespace)
	restConfig, err := skrManagerFactory.LoadRestConfig(ctx, gcSummary.Name, gcSummary.Key)
	if err != nil {
		return nil, fmt.Errorf("error loading skr rest config: %w", err)
	}

	return restConfig, nil
}
