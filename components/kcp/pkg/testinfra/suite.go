package testinfra

import (
	"context"
	"errors"
	"fmt"
	gardenapi "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/go-logr/logr"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/components/kcp/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/components/kcp/api/cloud-resources/v1beta1"
	cloudcontrolcontroller "github.com/kyma-project/cloud-manager/components/kcp/internal/controller/cloud-control"
	cloudresourcescontroller "github.com/kyma-project/cloud-manager/components/kcp/internal/controller/cloud-resources"
	awsmock "github.com/kyma-project/cloud-manager/components/kcp/pkg/provider/aws/mock"
	gcpiprangeclient "github.com/kyma-project/cloud-manager/components/kcp/pkg/provider/gcp/iprange/client"
	gcpFilestoreClient "github.com/kyma-project/cloud-manager/components/kcp/pkg/provider/gcp/nfsinstance/client"
	skrruntime "github.com/kyma-project/cloud-manager/components/kcp/pkg/skr/runtime"
	skrmanager "github.com/kyma-project/cloud-manager/components/kcp/pkg/skr/runtime/manager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"net/http"
	"path/filepath"
	goruntime "runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sync"
	"testing"
)

type ClusterType string

const (
	ClusterTypeKcp    = ClusterType("kcp")
	ClusterTypeSkr    = ClusterType("skr")
	ClusterTypeGarden = ClusterType("garden")
)

var onceLogger = sync.Once{}

var schemeMap map[ClusterType]*runtime.Scheme

var logger logr.Logger

func init() {
	schemeMap = map[ClusterType]*runtime.Scheme{
		ClusterTypeKcp:    runtime.NewScheme(),
		ClusterTypeSkr:    runtime.NewScheme(),
		ClusterTypeGarden: runtime.NewScheme(),
	}
	// KCP
	utilruntime.Must(clientgoscheme.AddToScheme(schemeMap[ClusterTypeKcp]))
	utilruntime.Must(cloudcontrolv1beta1.AddToScheme(schemeMap[ClusterTypeKcp]))
	// SKR
	utilruntime.Must(clientgoscheme.AddToScheme(schemeMap[ClusterTypeSkr]))
	utilruntime.Must(cloudresourcesv1beta1.AddToScheme(schemeMap[ClusterTypeSkr]))
	// Garden
	utilruntime.Must(clientgoscheme.AddToScheme(schemeMap[ClusterTypeGarden]))
	utilruntime.Must(gardenapi.AddToScheme(schemeMap[ClusterTypeGarden]))
}

type Cluster struct {
	t         *testing.T
	Name      string
	Type      ClusterType
	KymaRef   klog.ObjectRef
	Cfg       *rest.Config
	K8sClient client.Client
	TestEnv   *envtest.Environment
	Manager   manager.Manager
	Cancel    context.CancelFunc
	Ctx       context.Context
	Registry  skrruntime.SkrRegistry
	Runner    skrruntime.SkrRunner
	awsMock   awsmock.Server
}

type BaseSuite struct {
	suite.Suite
	kcpClusterName string
	clusters       map[string]*Cluster

	AwsMock awsmock.Server
}

func (s *BaseSuite) CommonSetupSuite() {
	s.AwsMock = awsmock.New()
	onceLogger.Do(func() {
		logger = zap.New(zap.UseDevMode(true))
		ctrl.SetLogger(logger)
	})
}

func (s *BaseSuite) CommonTearDownSuite() {
	for name, cluster := range s.clusters {
		if cluster.Cancel != nil {
			cluster.Cancel()
		}
		err := cluster.TestEnv.Stop()
		if err != nil {
			s.T().Errorf("Error stopping testenv %s: %s", name, err)
		}
	}
}

func (s *BaseSuite) Cluster(name string) *Cluster {
	return s.clusters[name]
}

type setupClusterOptions struct {
	kymaRef *klog.ObjectRef
}

type SetupClusterOptionFunc = func(options setupClusterOptions)

func WithKymaRef(namespace, name string) SetupClusterOptionFunc {
	return func(options setupClusterOptions) {
		options.kymaRef = &klog.ObjectRef{
			Name:      name,
			Namespace: namespace,
		}
	}
}

func (s *BaseSuite) SetupCluster(name string, clusterType ClusterType, opts ...SetupClusterOptionFunc) (*Cluster, error) {
	if err := initCrds(); err != nil {
		return nil, err
	}

	options := setupClusterOptions{}
	for _, o := range opts {
		o(options)
	}

	if clusterType == ClusterTypeKcp {
		if len(s.kcpClusterName) > 0 {
			return nil, fmt.Errorf("only one KCP cluster can be created, and %s is already started", s.kcpClusterName)
		}
		s.kcpClusterName = name
	}

	if clusterType == ClusterTypeSkr {
		if options.kymaRef == nil {
			return nil, errors.New("cluster of the SKR type must have WithKymaRef option")
		}
		if len(s.kcpClusterName) == 0 {
			return nil, errors.New("an SKR cluster can se started only once a KCP cluster is already started")
		}
	}

	cluster := &Cluster{
		t:       s.T(),
		Type:    clusterType,
		Name:    name,
		awsMock: s.AwsMock,
	}

	// TestEnv
	crdMap := map[ClusterType]string{
		ClusterTypeKcp:    dirKcp,
		ClusterTypeSkr:    dirSkr,
		ClusterTypeGarden: dirGarden,
	}
	dir, ok := crdMap[clusterType]
	if !ok {
		return nil, fmt.Errorf("unknown crd dir for cluster type %s", clusterType)
	}
	cluster.TestEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{dir},
		ErrorIfCRDPathMissing: true,
		BinaryAssetsDirectory: filepath.Join("..", "..", "bin", "k8s",
			fmt.Sprintf("1.28.3-%s-%s", goruntime.GOOS, goruntime.GOARCH)),
	}

	// Cfg
	cfg, err := cluster.TestEnv.Start()
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		return nil, errors.New("testenv started with nil cfg")
	}
	cluster.Cfg = cfg

	// Client
	sch, ok := schemeMap[clusterType]
	if !ok {
		return nil, fmt.Errorf("unknown scheme for cluster type %s", clusterType)
	}
	k8sClient, err := client.New(cfg, client.Options{Scheme: sch})
	if err != nil {
		return nil, err
	}
	if k8sClient == nil {
		return nil, errors.New("got nil k8sClient")
	}
	cluster.K8sClient = k8sClient

	// Ctx
	ctx, cancel := context.WithCancel(context.TODO())
	cluster.Ctx = ctx
	cluster.Cancel = cancel

	// KymaRef
	cluster.KymaRef.Name = options.kymaRef.Name
	cluster.KymaRef.Namespace = options.kymaRef.Namespace

	// Manager controller-runtime for KCP
	if clusterType == ClusterTypeKcp {
		mngr, err := ctrl.NewManager(cfg, ctrl.Options{
			Scheme: sch,
		})
		if err != nil {
			return nil, fmt.Errorf("error creating controller-runtime manager: %w", err)
		}
		cluster.Manager = mngr
	}

	// SkrManager for SKR
	if clusterType == ClusterTypeSkr {
		mngr, err := skrmanager.New(cfg, sch, cluster.KymaRef, logger)
		if err != nil {
			return nil, fmt.Errorf("error creating SKR Manager: %w", err)
		}
		cluster.Manager = mngr
		cluster.Registry = skrruntime.NewRegistry(sch)
		kcpCluster := s.clusters[s.kcpClusterName]
		cluster.Runner = skrruntime.NewRunner(cluster.Registry, kcpCluster.Manager)
	}

	return cluster, nil
}

func (s *BaseSuite) RunClusters() {
	for _, c := range s.clusters {
		c.Run()
	}
}

func (c *Cluster) Run() {
	if c.Manager == nil {
		return
	}
	go func() {
		err := c.Manager.Start(c.Ctx)
		assert.NoError(c.t, err)
	}()
}

func (c *Cluster) SetupAllControllers(awsMock awsmock.Server) error {
	switch c.Type {
	case ClusterTypeSkr:
		return c.setupAllSkrControllers()
	case ClusterTypeKcp:
		return c.setupAllKcpControllers()
	}
	return fmt.Errorf("unable to setup cluster %s of type %s", c.Name, c.Type)
}

func (c *Cluster) setupAllKcpControllers() (err error) {
	if err = cloudcontrolcontroller.NewScopeReconciler(
		c.Manager,
		c.awsMock.ScopeGardenProvider(),
	).SetupWithManager(c.Manager); err != nil {
		return err
	}
	if err = cloudcontrolcontroller.NewIpRangeReconciler(
		c.Manager,
		c.awsMock.IpRangeSkrProvider(),
		func(ctx context.Context, httpClient *http.Client) (gcpiprangeclient.ServiceNetworkingClient, error) {
			return nil, nil
		},
		func(ctx context.Context, httpClient *http.Client) (gcpiprangeclient.ComputeClient, error) {
			return nil, nil
		},
	).SetupWithManager(c.Manager); err != nil {
		return err
	}
	if err = cloudcontrolcontroller.NewNfsInstanceReconciler(
		c.Manager,
		c.awsMock.NfsInstanceSkrProvider(),
		func(ctx context.Context, httpClient *http.Client) (gcpFilestoreClient.FilestoreClient, error) {
			return nil, nil
		},
	).SetupWithManager(c.Manager); err != nil {
		return err
	}
	if err = cloudcontrolcontroller.NewVpcPeeringReconciler(c.Manager).SetupWithManager(c.Manager); err != nil {
		return err
	}
	return
}

func (c *Cluster) setupAllSkrControllers() (err error) {
	if err = cloudresourcescontroller.SetupCloudResourcesReconciler(c.Registry); err != nil {
		return err
	}
	if err = cloudresourcescontroller.SetupIpRangeReconciler(c.Registry); err != nil {
		return err
	}
	if err = cloudresourcescontroller.SetupAwsNfsVolumeReconciler(c.Registry); err != nil {
		return err
	}

	return
}
