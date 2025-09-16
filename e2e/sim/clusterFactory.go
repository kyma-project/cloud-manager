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
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
)

type ClientClusterFactory interface {
	CreateClientCluster(ctx context.Context, runtimeID string) (cluster.Cluster, error)
}

func NewClientClusterFactory(kcp client.Client) ClientClusterFactory {
	return &defaultClientClusterFactory{
		kcp: kcp,
	}
}

var _ ClientClusterFactory = &defaultClientClusterFactory{}

type defaultClientClusterFactory struct {
	kcp client.Client
}

func (f *defaultClientClusterFactory) CreateClientCluster(ctx context.Context, runtimeID string) (cluster.Cluster, error) {
	gc := &infrastructuremanagerv1.GardenerCluster{}
	err := f.kcp.Get(ctx, client.ObjectKey{Namespace: e2econfig.Config.KcpNamespace, Name: runtimeID}, gc)
	if client.IgnoreNotFound(err) != nil {
		return nil, fmt.Errorf("error getting GardenerCluster object: %w", err)
	}
	if apierrors.IsNotFound(err) {
		return nil, fmt.Errorf("GardenerCluster %q not found", runtimeID)
	}

	gcSummary := &util.GardenerClusterSummary{
		Key:       gc.Spec.Kubeconfig.Secret.Key,
		Name:      gc.Spec.Kubeconfig.Secret.Name,
		Namespace: gc.Spec.Kubeconfig.Secret.Namespace,
		Shoot:     gc.Spec.Shoot.Name,
	}

	skrManagerFactory := skrmanager.NewFactory(f.kcp, gcSummary.Namespace, bootstrap.SkrScheme)
	restConfig, err := skrManagerFactory.LoadRestConfig(ctx, gcSummary.Name, gcSummary.Key)
	if err != nil {
		return nil, fmt.Errorf("error loading skr rest config: %w", err)
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
