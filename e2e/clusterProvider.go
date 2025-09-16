package e2e

import (
	"context"
	"fmt"
	"os"
	"sync"

	gardenertypes "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/hashicorp/go-multierror"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/config/crd"
	e2econfig "github.com/kyma-project/cloud-manager/e2e/config"
	"github.com/kyma-project/cloud-manager/pkg/common/bootstrap"
	"github.com/kyma-project/cloud-manager/pkg/util"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	Init(ctx context.Context) error
	Stop() error
	KCP(ctx context.Context) (Cluster, error)
	Garden(ctx context.Context) (Cluster, error)
}

var _ ClusterProvider = &defaultClusterProvider{}

func NewClusterProvider() ClusterProvider {
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

func (p *defaultClusterProvider) Init(ctx context.Context) error {
	kcp, err := p.KCP(ctx)
	if err != nil {
		return fmt.Errorf("error initializing KCP: %w", err)
	}

	// install crds
	arr, err := crd.KCP_All()
	if err != nil {
		return fmt.Errorf("error reading CRDs: %w", err)
	}
	err = util.Apply(ctx, kcp.GetClient(), arr)
	if err != nil {
		return fmt.Errorf("error installing CRDs: %w", err)
	}

	// gardener credentials
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: e2econfig.Config.KcpNamespace,
			Name:      "gardener-credentials",
		},
	}

	err = kcp.GetAPIReader().Get(ctx, client.ObjectKeyFromObject(secret), secret)
	if err == nil {
		// already exists
		err = p.setGardenNamespaceInConfig(secret.Data["kubeconfig"])
		if err != nil {
			return fmt.Errorf("failed to set garden kubeconfig: %w", err)
		}

		garden, err := p.Garden(ctx)
		if err != nil {
			return fmt.Errorf("error initializing Garden: %w", err)
		}

		shootList := &gardenertypes.ShootList{}
		err = garden.GetClient().List(ctx, shootList, client.InNamespace(e2econfig.Config.GardenNamespace))
		if err != nil {
			return fmt.Errorf("error connecting to garden: %w", err)
		}

		return nil
	}

	if e2econfig.Config.GardenKubeconfig == "" {
		return fmt.Errorf("garden kubeconfig is not set in config")
	}
	kubeBytes, err := os.ReadFile(e2econfig.Config.GardenKubeconfig)
	if err != nil {
		return fmt.Errorf("failed to read garden kubeconfig from %q: %w", e2econfig.Config.GardenKubeconfig, err)
	}

	err = p.setGardenNamespaceInConfig(kubeBytes)
	if err != nil {
		return fmt.Errorf("failed to set garden kubeconfig: %w", err)
	}

	secret = &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: e2econfig.Config.KcpNamespace,
			Name:      "gardener-credentials",
		},
		Data: map[string][]byte{
			"kubeconfig": kubeBytes,
		},
	}
	err = kcp.GetClient().Create(ctx, secret)
	if apierrors.IsAlreadyExists(err) {
		// some race condition, let's assume the secret correctly created
		return nil
	}
	if err != nil {
		return fmt.Errorf("error creating gardener credentials: %w", err)
	}

	garden, err := p.Garden(ctx)
	if err != nil {
		return fmt.Errorf("error initializing Garden: %w", err)
	}

	shootList := &gardenertypes.ShootList{}
	err = garden.GetClient().List(ctx, shootList, client.InNamespace(e2econfig.Config.KcpNamespace))
	if err != nil {
		return fmt.Errorf("error connecting to garden: %w", err)
	}

	return nil
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

	p.garden = NewCluster("garden", clstr)

	err = p.garden.Start(ctx)

	return p.garden, err
}

func (p *defaultClusterProvider) setGardenNamespaceInConfig(gardenKubeBytes []byte) error {
	config, err := clientcmd.NewClientConfigFromBytes(gardenKubeBytes)
	if err != nil {
		return fmt.Errorf("error creating gardener client config: %w", err)
	}

	rawConfig, err := config.RawConfig()
	if err != nil {
		return fmt.Errorf("error getting gardener raw client config: %w", err)
	}

	if len(rawConfig.CurrentContext) > 0 {
		e2econfig.Config.GardenNamespace = rawConfig.Contexts[rawConfig.CurrentContext].Namespace
	}

	return nil
}
