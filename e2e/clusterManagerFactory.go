package e2e

import (
	"context"
	"fmt"

	commonscheme "github.com/kyma-project/cloud-manager/pkg/common/scheme"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

type ClusterManagerFactory interface {
	CreateClusterManager(ctx context.Context) (manager.Manager, error)
}

// KCP ==============================================

func NewKcpClusterManagerFactory(kcpRestConfig *rest.Config) ClusterManagerFactory {
	if kcpRestConfig == nil {
		kcpRestConfig = ctrl.GetConfigOrDie()
	}
	return &kcpClusterManagerFactory{
		kcpRestConfig: kcpRestConfig,
	}
}

type kcpClusterManagerFactory struct {
	kcpRestConfig *rest.Config
}

func (f *kcpClusterManagerFactory) CreateClusterManager(ctx context.Context) (manager.Manager, error) {
	m, err := manager.New(f.kcpRestConfig, manager.Options{
		Scheme: commonscheme.KcpScheme,
		Metrics: metricsserver.Options{
			BindAddress: "0", // disable
		},
		LeaderElection:         false, // disable
		HealthProbeBindAddress: "0",   // disable
		Client: client.Options{
			Cache: &client.CacheOptions{
				Unstructured: true,
			},
		},
		Logger: ctrl.Log.WithName("kcp"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create KCP cluster manager: %w", err)
	}

	return m, err
}

// Garden ========================================

func NewGardenClusterManagerFactory(kcpClient client.Client, gardenNamespace string) ClusterManagerFactory {
	return &gardenClusterManagerFactory{
		kcpClient:       kcpClient,
		gardenNamespace: gardenNamespace,
	}
}

type gardenClusterManagerFactory struct {
	kcpClient       client.Client
	gardenNamespace string
}

func (f *gardenClusterManagerFactory) CreateClusterManager(ctx context.Context) (manager.Manager, error) {
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

	m, err := manager.New(restConfig, manager.Options{
		Scheme: commonscheme.GardenScheme,
		Metrics: metricsserver.Options{
			BindAddress: "0", // disable
		},
		LeaderElection:         false, // disable
		HealthProbeBindAddress: "0",   // disable
		Client: client.Options{
			Cache: &client.CacheOptions{
				Unstructured: true,
			},
		},
		Cache: cache.Options{
			// restrict to single namespace
			// https://book.kubebuilder.io/cronjob-tutorial/empty-main.html
			// https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/cache#MultiNamespacedCacheBuilder
			DefaultNamespaces: map[string]cache.Config{
				f.gardenNamespace: {},
			},
		},
		Logger: ctrl.Log.WithName("garden"),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create Garden cluster manager: %w", err)
	}

	return m, nil
}
