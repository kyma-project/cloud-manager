package e2e

import (
	"context"
	"fmt"

	e2econfig "github.com/kyma-project/cloud-manager/e2e/config"
	"github.com/kyma-project/cloud-manager/pkg/common/bootstrap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
)

type ClusterFactory interface {
	CreateCluster(ctx context.Context) (Cluster, error)
}

// KCP ==============================================

func NewKcpClusterFactory(kcpRestConfig *rest.Config) ClusterFactory {
	if kcpRestConfig == nil {
		kcpRestConfig = ctrl.GetConfigOrDie()
	}
	return &kcpClusterFactory{
		kcpRestConfig: kcpRestConfig,
	}
}

type kcpClusterFactory struct{
	kcpRestConfig *rest.Config
}

func (f *kcpClusterFactory) CreateCluster(ctx context.Context) (Cluster, error) {
	clstr, err := cluster.New(f.kcpRestConfig, func(clusterOptions *cluster.Options) {
		clusterOptions.Scheme = bootstrap.KcpScheme
		clusterOptions.Client = client.Options{
			Cache: &client.CacheOptions{
				Unstructured: true,
			},
		}
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create KCP cluster: %w", err)
	}

	return NewCluster("kcp", clstr), nil
}

// Garden ========================================

func NewGardenClusterFactory(kcpClient client.Client) ClusterFactory {
	return &gardenClusterFactory{
		kcpClient: kcpClient,
	}
}

type gardenClusterFactory struct {
	kcpClient client.Client
}

func (f *gardenClusterFactory) CreateCluster(ctx context.Context) (Cluster, error) {
	secret := &corev1.Secret{}
	err := f.kcpClient.Get(ctx, types.NamespacedName{
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

	restConfig, err := clientcmd.RESTConfigFromKubeConfig(kubeBytes)
	if err != nil {
		return nil, fmt.Errorf("error creating gardener rest client config: %w", err)
	}

	clstr, err := cluster.New(restConfig, func(clusterOptions *cluster.Options) {
		// restrict to single namespace
		// https://book.kubebuilder.io/cronjob-tutorial/empty-main.html
		// https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/cache#MultiNamespacedCacheBuilder
		clusterOptions.Cache.DefaultNamespaces = map[string]cache.Config{
			e2econfig.Config.GardenNamespace: {},
		}
		clusterOptions.Scheme = bootstrap.GardenScheme
		clusterOptions.Client = client.Options{
			Cache: &client.CacheOptions{
				Unstructured: true,
			},
		}
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Garden cluster: %w", err)
	}

	return NewCluster("garden", clstr), nil
}
