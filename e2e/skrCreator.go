package e2e
//
//import (
//	"context"
//	"fmt"
//	"sync"
//	"time"
//
//	"github.com/elliotchance/pie/v2"
//	authenticationv1alpha1 "github.com/gardener/gardener/pkg/apis/authentication/v1alpha1"
//	gardenertypes "github.com/gardener/gardener/pkg/apis/core/v1beta1"
//	gardenerhelper "github.com/gardener/gardener/pkg/apis/core/v1beta1/helper"
//	"github.com/go-logr/logr"
//	"github.com/hashicorp/go-multierror"
//	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
//	e2econfig "github.com/kyma-project/cloud-manager/e2e/config"
//	"github.com/kyma-project/cloud-manager/pkg/common/bootstrap"
//	"github.com/kyma-project/cloud-manager/pkg/composed"
//	"github.com/kyma-project/cloud-manager/pkg/external/infrastructuremanagerv1"
//	"github.com/kyma-project/cloud-manager/pkg/external/operatorv1beta2"
//	skrmanager "github.com/kyma-project/cloud-manager/pkg/skr/runtime/manager"
//	"github.com/kyma-project/cloud-manager/pkg/util"
//	corev1 "k8s.io/api/core/v1"
//	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
//	"k8s.io/apimachinery/pkg/types"
//	"k8s.io/apimachinery/pkg/util/wait"
//	"k8s.io/utils/ptr"
//	"sigs.k8s.io/controller-runtime/pkg/client"
//	"sigs.k8s.io/controller-runtime/pkg/cluster"
//)
//
//type CreatorSkrInput struct {
//	Subscription *e2econfig.SubscriptionInfo
//	Region       string
//	Logger       logr.Logger
//}
//
//type SkrCreator interface {
//	// CreateSkr creates new SKR for the given alias, or returns an existing one. It locks
//	// per alias, so it's blocking for the same alias, but different aliases can run in parallel
//	// Returns the SkrCluster for the given alias, true if it is created in this call, or false
//	// it was already created, and the error that might occur during creation
//	CreateSkr(ctx context.Context, alias string, in CreatorSkrInput) (SkrCluster, bool, error)
//	Has(alias string) bool
//	AllAliases() []string
//	AllClusters() []SkrCluster
//	GetByAlias(alias string) SkrCluster
//	GetByRuntimeId(runtimeId string) SkrCluster
//
//	ImportShared(ctx context.Context, runtimeId string) (SkrCluster, error)
//
//	// Remove removes SKR for the specified alias from the registry and stops it if it's started
//	Remove(alias string) error
//
//	// Stop stops and removes all SKR clusters w/out deleting them
//	Stop() error
//}
//
//type skrCreator struct {
//	m           sync.Mutex
//	partialLock map[string]*sync.Mutex
//
//	kcp    Cluster
//	garden Cluster
//
//	skrClusters map[string]SkrCluster
//}
//
//func NewSkrCreator(kcp, garden Cluster) SkrCreator {
//	return &skrCreator{
//		partialLock: make(map[string]*sync.Mutex),
//		kcp:         kcp,
//		garden:      garden,
//		skrClusters: make(map[string]SkrCluster),
//	}
//}
//
//func (c *skrCreator) getPartialLock(alias string) *sync.Mutex {
//	c.m.Lock()
//	defer c.m.Unlock()
//
//	lock, ok := c.partialLock[alias]
//	if !ok {
//		lock = &sync.Mutex{}
//		c.partialLock[alias] = lock
//	}
//
//	return lock
//}
//
//func (c *skrCreator) Stop() error {
//	c.m.Lock()
//	defer c.m.Unlock()
//
//	var result error
//
//	for _, skr := range c.skrClusters {
//		delete(c.skrClusters, skr.Alias())
//		if skr.IsStarted() {
//			if err := skr.Stop(); err != nil {
//				result = multierror.Append(result, fmt.Errorf("failed to stop skr %s: %w", skr.Alias, err))
//			}
//		}
//	}
//
//	return result
//}
//
//func (c *skrCreator) Remove(alias string) error {
//	c.m.Lock()
//	defer c.m.Unlock()
//
//	skr := c.skrClusters[alias]
//	delete(c.skrClusters, alias)
//	if skr != nil && skr.IsStarted() {
//		return skr.Stop()
//	}
//	return nil
//}
//
//func (c *skrCreator) Has(alias string) bool {
//	c.m.Lock()
//	defer c.m.Unlock()
//
//	_, ok := c.skrClusters[alias]
//	if ok {
//		return true
//	}
//	return false
//}
//
//func (c *skrCreator) GetByAlias(alias string) SkrCluster {
//	c.m.Lock()
//	defer c.m.Unlock()
//
//	result := c.skrClusters[alias]
//	if result != nil {
//		return result
//	}
//	return nil
//}
//
//func (c *skrCreator) GetByRuntimeId(runtimeId string) SkrCluster {
//	c.m.Lock()
//	defer c.m.Unlock()
//
//	for _, skr := range c.skrClusters {
//		if skr.RuntimeID() == runtimeId {
//			return skr
//		}
//	}
//	return nil
//}
//
//func (c *skrCreator) AllAliases() []string {
//	c.m.Lock()
//	defer c.m.Unlock()
//
//	result := pie.Keys(c.skrClusters)
//	return result
//}
//
//func (c *skrCreator) AllClusters() []SkrCluster {
//	c.m.Lock()
//	defer c.m.Unlock()
//
//	return pie.Values(c.skrClusters)
//}
//
//func (c *skrCreator) CreateSkr(ctx context.Context, alias string, in CreatorSkrInput) (SkrCluster, bool, error) {
//	if in.Logger.GetSink() == nil {
//		in.Logger = logr.Discard()
//	}
//
//	lock := c.getPartialLock(alias)
//
//	lock.Lock()
//	defer lock.Unlock()
//
//	GetScenarioSession(ctx).RegisterCluster(alias)
//
//	existing := c.GetByAlias(alias)
//	if existing != nil {
//		return existing, false, nil
//	}
//
//	if in.Subscription == nil {
//		return nil, true, fmt.Errorf("subscription is required")
//	}
//	if in.Region == "" {
//		in.Region = simXXX.defaultRegions[in.Subscription.Provider]
//		in.Logger.Info(fmt.Sprintf("Region defaulted to %s", in.Region))
//	}
//
//	secretBinding := &gardenertypes.SecretBinding{}
//	err := c.garden.GetAPIReader().Get(ctx, types.NamespacedName{
//		Name:      in.Subscription.Name,
//		Namespace: e2econfig.Config.GardenNamespace,
//	}, secretBinding)
//	if err != nil {
//		return nil, true, fmt.Errorf("could not load secret bindings %q: %w", in.Subscription.Name, err)
//	}
//	in.Logger.Info("SecretBinding loaded")
//
//	// kcp runtime
//
//	runtimeBuilder := simXXX.NewRuntimeBuilder().
//		WithProvider(in.Subscription.Provider, in.Region).
//		WithSecretBindingName(secretBinding.Name)
//	if err := runtimeBuilder.Validate(); err != nil {
//		return nil, true, fmt.Errorf("invalid runtime: %w", err)
//	}
//	rt := runtimeBuilder.Build()
//	err = c.kcp.GetClient().Create(ctx, rt)
//	if err != nil {
//		return nil, true, fmt.Errorf("error creating runtime: %w", err)
//	}
//	in.Logger.Info("Runtime created")
//	// garden shoot
//
//	shootBuilder := simXXX.NewShootBuilder().
//		WithRuntime(rt)
//	if err := shootBuilder.Validate(); err != nil {
//		return nil, true, fmt.Errorf("invalid shoot: %w", err)
//	}
//	shoot := shootBuilder.Build()
//
//	err = c.garden.AddResources(ctx, &ResourceDeclaration{
//		Alias:      fmt.Sprintf("shoot-%s", rt.Name),
//		Kind:       "Shoot",
//		ApiVersion: gardenertypes.SchemeGroupVersion.String(),
//		Name:       shoot.Name,
//		Namespace:  shoot.Namespace,
//	})
//	if err != nil {
//		return nil, true, fmt.Errorf("error adding shoot resource to the world: %w", err)
//	}
//
//	err = c.garden.GetClient().Create(ctx, shoot)
//	if err != nil {
//		return nil, true, fmt.Errorf("error creating shoot: %w", err)
//	}
//	in.Logger.Info("Shoot created, waiting ready...")
//
//	// wait shoot ready
//
//	err = c.waitShootReady(ctx, shoot)
//	if err != nil {
//		return nil, true, err
//	}
//	in.Logger.Info("Shoot ready")
//
//	// get shoot's kubeconfig
//
//	kubeConfig, err := c.getShootKubeconfig(ctx, shoot)
//	if err != nil {
//		return nil, true, err
//	}
//	in.Logger.Info("Admin kubeconfig obtained")
//
//	// kcp kubeconfig secret
//
//	kubeSecret := &corev1.Secret{
//		ObjectMeta: metav1.ObjectMeta{
//			Namespace: rt.Namespace,
//			Name:      "kubeconfig-" + rt.Name,
//		},
//		StringData: map[string]string{
//			"config": string(kubeConfig),
//		},
//	}
//	err = c.kcp.GetClient().Create(ctx, kubeSecret)
//	if err != nil {
//		return nil, true, fmt.Errorf("error creating kubeconfig secret: %w", err)
//	}
//	in.Logger.Info("Kubeconfig secret created")
//
//	err = c.kcp.AddResources(ctx, &ResourceDeclaration{
//		Alias:      fmt.Sprintf("kubesecret-%s", rt.Name),
//		Kind:       "Secret",
//		ApiVersion: corev1.SchemeGroupVersion.String(),
//		Name:       kubeSecret.Name,
//		Namespace:  kubeSecret.Namespace,
//	})
//	if err != nil {
//		return nil, true, fmt.Errorf("error adding shoot kubeconfig secret resource to the world: %w", err)
//	}
//
//	// kcp gardener cluster
//
//	gc := &infrastructuremanagerv1.GardenerCluster{
//		ObjectMeta: metav1.ObjectMeta{
//			Namespace: rt.Namespace,
//			Name:      rt.Name,
//			Labels: map[string]string{
//				cloudcontrolv1beta1.LabelScopeGlobalAccountId: rt.Labels[cloudcontrolv1beta1.LabelScopeGlobalAccountId],
//				cloudcontrolv1beta1.LabelScopeSubaccountId:    rt.Labels[cloudcontrolv1beta1.LabelScopeSubaccountId],
//				cloudcontrolv1beta1.LabelScopeShootName:       rt.Labels[cloudcontrolv1beta1.LabelScopeShootName],
//				cloudcontrolv1beta1.LabelScopeRegion:          rt.Labels[cloudcontrolv1beta1.LabelScopeRegion],
//				cloudcontrolv1beta1.LabelScopeBrokerPlanName:  rt.Labels[cloudcontrolv1beta1.LabelScopeBrokerPlanName],
//				cloudcontrolv1beta1.LabelScopeProvider:        rt.Labels[cloudcontrolv1beta1.LabelScopeProvider],
//				cloudcontrolv1beta1.LabelRuntimeId:            rt.Name,
//			},
//		},
//		Spec: infrastructuremanagerv1.GardenerClusterSpec{
//			Shoot: infrastructuremanagerv1.Shoot{
//				Name: shoot.Name,
//			},
//			Kubeconfig: infrastructuremanagerv1.Kubeconfig{
//				Secret: infrastructuremanagerv1.Secret{
//					Key:       "config",
//					Name:      kubeSecret.Name,
//					Namespace: kubeSecret.Namespace,
//				},
//			},
//		},
//	}
//	err = c.kcp.GetClient().Create(ctx, gc)
//	if err != nil {
//		return nil, true, fmt.Errorf("error creating gardencluster: %w", err)
//	}
//	gcSummary := &util.GardenerClusterSummary{
//		Key:       gc.Spec.Kubeconfig.Secret.Key,
//		Name:      gc.Spec.Kubeconfig.Secret.Name,
//		Namespace: gc.Spec.Kubeconfig.Secret.Namespace,
//		Shoot:     gc.Spec.Shoot.Name,
//	}
//
//	err = c.kcp.AddResources(ctx, &ResourceDeclaration{
//		Alias:      fmt.Sprintf("gardencluster-%s", rt.Name),
//		Kind:       "GardenerCluster",
//		ApiVersion: infrastructuremanagerv1.GroupVersion.String(),
//		Name:       gc.Name,
//		Namespace:  gc.Namespace,
//	})
//	if err != nil {
//		return nil, true, fmt.Errorf("error adding gardenercluster resource to the world: %w", err)
//	}
//	in.Logger.Info("GardenerCluster created")
//
//	kyma := &operatorv1beta2.Kyma{
//		ObjectMeta: metav1.ObjectMeta{
//			Namespace: rt.Namespace,
//			Name:      rt.Name,
//		},
//		Spec: operatorv1beta2.KymaSpec{
//			Channel: "regular",
//		},
//		Status: operatorv1beta2.KymaStatus{
//			State: "Ready",
//		},
//	}
//	err = c.kcp.GetClient().Create(ctx, kyma)
//	if err != nil {
//		return nil, true, fmt.Errorf("error creating kyma: %w", err)
//	}
//	err = composed.PatchObjStatus(ctx, kyma, c.kcp.GetClient())
//	if err != nil {
//		return nil, true, fmt.Errorf("error patching kyma status: %w", err)
//	}
//
//	err = c.kcp.AddResources(ctx, &ResourceDeclaration{
//		Alias:      fmt.Sprintf("kyma-%s", rt.Name),
//		Kind:       "Kyma",
//		ApiVersion: kyma.GroupVersionKind().GroupVersion().String(),
//		Name:       kyma.GetName(),
//		Namespace:  kyma.GetNamespace(),
//	})
//	if err != nil {
//		return nil, true, fmt.Errorf("error adding kyma resource to the world: %w", err)
//	}
//	in.Logger.Info("KCP Kyma created")
//
//	// start the cluster
//
//	ns := gcSummary.Namespace
//	if ns == "" {
//		ns = e2econfig.Config.KcpNamespace
//	}
//
//	skrManagerFactory := skrmanager.NewFactory(c.kcp.GetAPIReader(), ns, bootstrap.SkrScheme)
//	restConfig, err := skrManagerFactory.LoadRestConfig(ctx, gcSummary.Name, gcSummary.Key)
//	if err != nil {
//		return nil, true, fmt.Errorf("error loading skr rest config: %w", err)
//	}
//
//	clstr, err := cluster.New(restConfig, func(clusterOptions *cluster.Options) {
//		clusterOptions.Scheme = bootstrap.SkrScheme
//		clusterOptions.Client = client.Options{
//			Cache: &client.CacheOptions{
//				Unstructured: true,
//			},
//		}
//	})
//	if err != nil {
//		return nil, true, fmt.Errorf("failed to create SKR cluster: %w", err)
//	}
//
//	theCluster := NewCluster(alias, clstr)
//
//	skrCluster := NewSkrCluster(theCluster, rt)
//
//	c.m.Lock()
//	defer c.m.Unlock()
//
//	c.skrClusters[alias] = skrCluster
//
//	in.Logger.Info("Starting cluster")
//
//	err = skrCluster.Start(ctx)
//	if err != nil {
//		return nil, true, fmt.Errorf("failed to start SKR cluster %q: %w", alias, err)
//	}
//
//	err = skrCluster.EnsureKymaCR(ctx)
//	if err != nil {
//		return nil, true, fmt.Errorf("failed to create SKR Kyma: %w", err)
//	}
//
//	return skrCluster, true, nil
//}
//
//func (c *skrCreator) ImportShared(ctx context.Context, runtimeId string) (SkrCluster, error) {
//	c.m.Lock()
//	defer c.m.Unlock()
//
//	// Runtime
//
//	rt := &infrastructuremanagerv1.Runtime{}
//	err := c.kcp.GetClient().Get(ctx, client.ObjectKey{
//		Namespace: e2econfig.Config.KcpNamespace,
//		Name:      runtimeId,
//	}, rt)
//	if err != nil {
//		return nil, fmt.Errorf("failed to load runtime %q: %w", runtimeId, err)
//	}
//
//	pt, err := cloudcontrolv1beta1.ParseProviderType(rt.Spec.Shoot.Provider.Type)
//	if err != nil {
//		return nil, err
//	}
//
//	alias := SharedSkrClusterAlias(pt)
//
//	_, ok := c.skrClusters[alias]
//	if ok {
//		return nil, fmt.Errorf("the shared SKR with alias %q for runtime %q already exists", alias, rt.Name)
//	}
//
//	err = c.kcp.AddResources(ctx, &ResourceDeclaration{
//		Alias:      SharedRuntimeResourceAlias(rt.Name),
//		Kind:       "Runtime",
//		ApiVersion: infrastructuremanagerv1.GroupVersion.String(),
//		Name:       rt.Name,
//		Namespace:  rt.Namespace,
//	})
//	if err != nil {
//		return nil, fmt.Errorf("error adding runtime resource: %w", err)
//	}
//
//	// Shoot
//
//	shoot := &gardenertypes.Shoot{}
//	err = c.garden.GetClient().Get(ctx, types.NamespacedName{
//		Namespace: e2econfig.Config.GardenNamespace,
//		Name:      rt.Spec.Shoot.Name,
//	}, shoot)
//	if err != nil {
//		return nil, fmt.Errorf("failed to load shoot %q for runtime %q: %w", rt.Spec.Shoot.Name, runtimeId, err)
//	}
//
//	err = c.garden.AddResources(ctx, &ResourceDeclaration{
//		Alias:      SharedShootResourceAlias(shoot.Name),
//		Kind:       "Shoot",
//		ApiVersion: gardenertypes.SchemeGroupVersion.String(),
//		Name:       shoot.Name,
//		Namespace:  shoot.Namespace,
//	})
//	if err != nil {
//		return nil, fmt.Errorf("error adding shoot resource: %w", err)
//	}
//	// Kubeconfig
//
//	kubeConfigBytes, err := c.getShootKubeconfig(ctx, shoot)
//	if err != nil {
//		return nil, fmt.Errorf("failed to get shoot %q for runtime %q kubeconfig: %w", rt.Spec.Shoot.Name, runtimeId, err)
//	}
//
//	// GardenerCluster ====================================
//
//	gc := &infrastructuremanagerv1.GardenerCluster{}
//	err = c.kcp.GetClient().Get(ctx, types.NamespacedName{
//		Namespace: rt.Namespace,
//		Name:      rt.Name,
//	}, gc)
//	if err != nil {
//		return nil, fmt.Errorf("failed to get gardener cluster %q: %w", rt.Name, err)
//	}
//
//	err = c.kcp.AddResources(ctx, &ResourceDeclaration{
//		Alias:      SharedGardenerClusterResourceAlias(gc.Name),
//		Kind:       "GardenerCluster",
//		ApiVersion: infrastructuremanagerv1.GroupVersion.String(),
//		Name:       gc.Name,
//		Namespace:  gc.Namespace,
//	})
//	if err != nil {
//		return nil, fmt.Errorf("error adding gardener cluster resource: %w", err)
//	}
//
//	gcSummary := &util.GardenerClusterSummary{
//		Key:       gc.Spec.Kubeconfig.Secret.Key,
//		Name:      gc.Spec.Kubeconfig.Secret.Name,
//		Namespace: gc.Spec.Kubeconfig.Secret.Namespace,
//		Shoot:     gc.Spec.Shoot.Name,
//	}
//
//	// update secret with the new kubeconfig
//	kubeSecret := &corev1.Secret{}
//	err = c.kcp.GetClient().Get(ctx, types.NamespacedName{
//		Namespace: gcSummary.Namespace,
//		Name:      gcSummary.Name,
//	}, kubeSecret)
//	if client.IgnoreNotFound(err) != nil {
//		return nil, fmt.Errorf("failed to get kubeconfig secret %q: %w", gcSummary.Name, err)
//	}
//	if err == nil {
//		// update
//		kubeSecret.StringData = map[string]string{
//			gcSummary.Key: string(kubeConfigBytes),
//		}
//		err = c.kcp.GetClient().Update(ctx, kubeSecret)
//		if err != nil {
//			return nil, fmt.Errorf("failed to update kubeconfig secret %q: %w", gcSummary.Name, err)
//		}
//	} else {
//		// create
//		kubeSecret := &corev1.Secret{
//			ObjectMeta: metav1.ObjectMeta{
//				Namespace: gcSummary.Namespace,
//				Name:      gcSummary.Name,
//			},
//			StringData: map[string]string{
//				gcSummary.Key: string(kubeConfigBytes),
//			},
//		}
//		err = c.kcp.GetClient().Create(ctx, kubeSecret)
//		if err != nil {
//			return nil, fmt.Errorf("failed to create kubeconfig secret %q: %w", gcSummary.Name, err)
//		}
//	}
//
//	// start the cluster
//
//	ns := gcSummary.Namespace
//	if ns == "" {
//		ns = e2econfig.Config.KcpNamespace
//	}
//
//	skrManagerFactory := skrmanager.NewFactory(c.kcp.GetAPIReader(), ns, bootstrap.SkrScheme)
//	restConfig, err := skrManagerFactory.LoadRestConfig(ctx, gcSummary.Name, gcSummary.Key)
//	if err != nil {
//		return nil, fmt.Errorf("error loading skr rest config: %w", err)
//	}
//
//	clstr, err := cluster.New(restConfig, func(clusterOptions *cluster.Options) {
//		clusterOptions.Scheme = bootstrap.SkrScheme
//		clusterOptions.Client = client.Options{
//			Cache: &client.CacheOptions{
//				Unstructured: true,
//			},
//		}
//	})
//	if err != nil {
//		return nil, fmt.Errorf("failed to create SKR cluster: %w", err)
//	}
//
//	theCluster := NewCluster(alias, clstr)
//
//	skrCluster := NewSkrCluster(theCluster, rt)
//
//	c.skrClusters[alias] = skrCluster
//
//	err = skrCluster.Start(ctx)
//	if err != nil {
//		return nil, fmt.Errorf("failed to start SKR cluster %q: %w", alias, err)
//	}
//
//	err = skrCluster.EnsureKymaCR(ctx)
//	if err != nil {
//		return nil, fmt.Errorf("failed to create SKR %q Kyma: %w", rt.Name, err)
//	}
//
//	return skrCluster, nil
//}
//
//// private ===========================================
//
//func (c *skrCreator) waitShootReady(ctx context.Context, shoot *gardenertypes.Shoot) error {
//	ctx, cancel := context.WithTimeout(ctx, 15*time.Minute)
//	defer cancel()
//	err := wait.PollUntilContextCancel(ctx, 5*time.Second, false, func(ctx context.Context) (bool, error) {
//		err := c.garden.GetClient().Get(ctx, client.ObjectKeyFromObject(shoot), shoot)
//		if err != nil {
//			return false, err
//		}
//		if len(shoot.Status.Conditions) == 0 {
//			return false, nil
//		}
//		ctList := []gardenertypes.ConditionType{
//			gardenertypes.ShootControlPlaneHealthy,
//			gardenertypes.ShootAPIServerAvailable,
//			gardenertypes.ShootEveryNodeReady,
//			gardenertypes.ShootSystemComponentsHealthy,
//			//gardenertypes.ShootObservabilityComponentsHealthy,
//		}
//		for _, ct := range ctList {
//			cond := gardenerhelper.GetCondition(shoot.Status.Conditions, ct)
//			if cond == nil {
//				return false, nil
//			}
//			if cond.Status != gardenertypes.ConditionTrue {
//				return false, nil
//			}
//		}
//		return true, nil
//	})
//	if err != nil {
//		return fmt.Errorf("error waiting for shoot to be ready: %w", err)
//	}
//	return nil
//}
//
//func (c *skrCreator) getShootKubeconfig(ctx context.Context, shoot *gardenertypes.Shoot) ([]byte, error) {
//	adminKubeconfigRequest := &authenticationv1alpha1.AdminKubeconfigRequest{
//		Spec: authenticationv1alpha1.AdminKubeconfigRequestSpec{
//			ExpirationSeconds: ptr.To(int64(3600 * 6)), // 6 hours
//		},
//	}
//	err := c.garden.GetClient().SubResource("adminkubeconfig").Create(ctx, shoot, adminKubeconfigRequest)
//	if err != nil {
//		return nil, fmt.Errorf("error creating admin kubeconfig: %w", err)
//	}
//	return adminKubeconfigRequest.Status.Kubeconfig, nil
//}
