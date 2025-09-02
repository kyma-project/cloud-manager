package e2e

import (
	"context"
	"fmt"
	"sync"

	"github.com/hashicorp/go-multierror"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/bootstrap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
)

type ProviderSkrInput struct {
	Provider cloudcontrolv1beta1.ProviderType
}

type ClusterProvider interface {
	Stop() error
	KCP(ctx context.Context) (Cluster, error)
	Garden(ctx context.Context) (Cluster, error)
}

var _ ClusterProvider = &defaultClusterProvider{}

func newClusterProvider() ClusterProvider {
	return &defaultClusterProvider{}
}

// defaultClusterProvider is a default implementation of ClusterProvider that
// * KCP loads from the KUBECONFIG environment variable
// * Garden loads from the kcp secret "gardener-credentials"
type defaultClusterProvider struct {
	mKcp    sync.Mutex
	mGarden sync.Mutex

	kcp    Cluster
	garden Cluster
}

func (p *defaultClusterProvider) Stop() error {
	p.mKcp.Lock()
	p.mGarden.Lock()
	defer p.mKcp.Unlock()
	defer p.mGarden.Unlock()

	var result error

	if p.kcp != nil && p.kcp.IsStarted() {
		if err := p.kcp.Stop(); err != nil {
			result = multierror.Append(result, err)
		}
	}
	if p.garden != nil && p.garden.IsStarted() {
		if err := p.garden.Stop(); err != nil {
			result = multierror.Append(result, err)
		}
	}

	return result
}

func (p *defaultClusterProvider) KCP(ctx context.Context) (Cluster, error) {
	p.mKcp.Lock()
	defer p.mKcp.Unlock()

	if p.kcp != nil {
		return p.kcp, nil
	}

	restConfig := ctrl.GetConfigOrDie()
	clstr, err := cluster.New(restConfig, func(clusterOptions *cluster.Options) {
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

	p.kcp = NewCluster("kcp", clstr)

	err = p.kcp.Start(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start KCP cluster: %w", err)
	}
	if !p.kcp.GetCache().WaitForCacheSync(ctx) {
		return nil, fmt.Errorf("failed to sync KCP cluster")
	}

	return p.kcp, err
}

func (p *defaultClusterProvider) Garden(ctx context.Context) (Cluster, error) {
	kcp, err := p.KCP(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get KCP cluster: %w", err)
	}

	p.mGarden.Lock()
	defer p.mGarden.Unlock()

	if p.garden != nil {
		return p.garden, nil
	}

	secret := &corev1.Secret{}
	err = kcp.GetClient().Get(ctx, types.NamespacedName{
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
			Config.GardenNamespace: {},
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

	p.garden = NewCluster("garden", clstr)

	err = p.garden.Start(ctx)

	return p.garden, err
}
