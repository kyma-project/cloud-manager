package e2e

import (
	"context"
	"fmt"
	"maps"
	"sync"

	"github.com/kyma-project/cloud-manager/pkg/common/bootstrap"
	"github.com/kyma-project/cloud-manager/pkg/external/keb"
	skrmanager "github.com/kyma-project/cloud-manager/pkg/skr/runtime/manager"
	"github.com/kyma-project/cloud-manager/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
)

type ClusterProvider interface {
	KCP(ctx context.Context) (*Cluster, error)
	SKR(ctx context.Context, id string) (*Cluster, error)
	KnownSkrClusters() map[string]*Cluster
	Garden(ctx context.Context) (*Cluster, error)
}

var _ ClusterProvider = &defaultClusterProvider{}

// defaultClusterProvider is a default implementation of ClusterProvider that
// * KCP loads from the KUBECONFIG environment variable
// * SKR loads from the SKR_KUBECONFIG environment variable
type defaultClusterProvider struct {
	m sync.Mutex

	kcp    *Cluster
	skr    map[string]*Cluster
	garden *Cluster
}

func (p *defaultClusterProvider) KnownSkrClusters() map[string]*Cluster {
	result := make(map[string]*Cluster)
	maps.Copy(result, p.skr)
	return result
}

func (p *defaultClusterProvider) KCP(ctx context.Context) (*Cluster, error) {
	p.m.Lock()
	defer p.m.Unlock()

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

	p.kcp = NewCluster(clstr)

	err = p.kcp.Start(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start KCP cluster: %w", err)
	}
	if !p.kcp.GetCache().WaitForCacheSync(ctx) {
		return nil, fmt.Errorf("failed to sync KCP cluster")
	}

	return p.kcp, err
}

func (p *defaultClusterProvider) SKR(ctx context.Context, id string) (*Cluster, error) {
	p.m.Lock()
	defer p.m.Unlock()

	if p.skr == nil {
		p.skr = make(map[string]*Cluster)
	}

	if clstr, exists := p.skr[id]; exists {
		return clstr, nil
	}

	kcp, err := p.KCP(ctx)
	if err != nil {
		return nil, err
	}

	gardenerClusterLoaders := []func() (*unstructured.Unstructured, error){
		func() (*unstructured.Unstructured, error) {
			gc := util.NewGardenerClusterUnstructured()
			err := kcp.GetClient().Get(ctx, client.ObjectKey{Namespace: Config.KcpNamespace, Name: id}, gc)
			if client.IgnoreNotFound(err) != nil {
				return nil, fmt.Errorf("error loading GardenerCluster exact name %s: %w", id, err)
			}
			if err == nil {
				return gc, nil
			}
			return nil, nil
		},
		func() (*unstructured.Unstructured, error) {
			gardenerClusterList := util.NewGardenerClusterListUnstructured()
			err := kcp.GetClient().List(ctx, gardenerClusterList, client.InNamespace(Config.KcpNamespace), client.MatchingLabels{keb.LabelKymaRuntimeID: id})
			if err != nil {
				return nil, fmt.Errorf("error listing GardenerCluster by runtime-id label: %w", err)
			}
			if len(gardenerClusterList.Items) == 0 {
				return nil, nil
			}
			if len(gardenerClusterList.Items) > 1 {
				return nil, fmt.Errorf("multiple GardenerCluster found for id '%s' matching runtime-id label", id)
			}
			return &gardenerClusterList.Items[0], nil
		},
		func() (*unstructured.Unstructured, error) {
			gardenerClusterList := util.NewGardenerClusterListUnstructured()
			err := kcp.GetClient().List(ctx, gardenerClusterList, client.InNamespace(Config.KcpNamespace), client.MatchingLabels{keb.LabelKymaShootName: id})
			if err != nil {
				return nil, fmt.Errorf("error listing GardenerCluster by shoot-name label: %w", err)
			}
			if len(gardenerClusterList.Items) == 0 {
				return nil, nil
			}
			if len(gardenerClusterList.Items) > 1 {
				return nil, fmt.Errorf("multiple GardenerCluster found for id'%s' matching shoot-name label", id)
			}
			return &gardenerClusterList.Items[0], nil
		},
	}

	var gardenerCluster *unstructured.Unstructured
	for _, gardenerClusterLoader := range gardenerClusterLoaders {
		gardenerCluster, err = gardenerClusterLoader()
		if err != nil {
			return nil, fmt.Errorf("error loading GardenerCluster: %w", err)
		}
		if gardenerCluster != nil {
			break
		}
	}
	if gardenerCluster == nil {
		return nil, fmt.Errorf("GardenerCluster with id '%s' not found", id)
	}

	gardenerClusterSummary, err := util.ExtractGardenerClusterSummary(gardenerCluster)
	if err != nil {
		return nil, fmt.Errorf("error extracting GardenerCluster summary: %w", err)
	}
	ns := gardenerClusterSummary.Namespace
	if ns == "" {
		ns = Config.KcpNamespace
	}

	skrManagerFactory := skrmanager.NewFactory(kcp.GetAPIReader(), ns, bootstrap.SkrScheme)
	restConfig, err := skrManagerFactory.LoadRestConfig(ctx, gardenerClusterSummary.Name, gardenerClusterSummary.Key)
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

	p.skr[id] = NewCluster(clstr)

	err = p.skr[id].Start(ctx)

	return p.skr[id], err
}

func (p *defaultClusterProvider) Garden(ctx context.Context) (*Cluster, error) {
	kcp, err := p.KCP(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get KCP cluster: %w", err)
	}

	p.m.Lock()
	defer p.m.Unlock()

	if p.garden != nil {
		return p.garden, nil
	}

	secret := &corev1.Secret{}
	err = kcp.Cluster.GetClient().Get(ctx, types.NamespacedName{
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

	p.garden = NewCluster(clstr)

	err = p.garden.Start(ctx)

	return p.garden, err
}
