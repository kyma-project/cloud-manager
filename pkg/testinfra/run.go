package testinfra

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	goruntime "runtime"
	"time"

	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	commonconfig "github.com/kyma-project/cloud-manager/pkg/common/config"
	"github.com/kyma-project/cloud-manager/pkg/config"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	awsmock "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/mock"
	azuremock "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/mock"
	gcpmock "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/mock"
	sapmock "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/mock"
	skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime"
	"github.com/kyma-project/cloud-manager/pkg/testinfra/infraScheme"
	"github.com/kyma-project/cloud-manager/pkg/testinfra/infraTypes"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"github.com/kyma-project/cloud-manager/pkg/util/debugged"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

const (
	DefaultSkrNamespace    = "test"
	DefaultKcpNamespace    = "kcp-system"
	DefaultGardenNamespace = "garden-kyma"
)

func Start() (Infra, error) {
	projectRoot := os.Getenv("PROJECTROOT")
	if len(projectRoot) == 0 {
		return nil, errors.New("the env var PROJECTROOT must be specified and point to the dir where Makefile is")
	}
	envtestK8sVersion := os.Getenv("ENVTEST_K8S_VERSION")
	if len(envtestK8sVersion) == 0 {
		panic(errors.New("unable to resolve envtest version. Use env var ENVTEST_K8S_VERSION to specify it"))
	}

	ginkgo.By("Preparing CRDs")

	dirSkr, dirKcp, dirGarden, err := initCrds(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("error initializing CRDs: %w", err)
	}

	configDir := filepath.Join(projectRoot, "bin", "cloud-manager", "config")
	if err := os.MkdirAll(configDir, 0777); err != nil {
		return nil, fmt.Errorf("error creating config dir: %w", err)
	}

	infra := &infra{
		projectRootDir: projectRoot,
		clusters: map[infraTypes.ClusterType]*clusterInfo{
			infraTypes.ClusterTypeKcp: &clusterInfo{
				crdDirs: []string{dirKcp},
			},
			infraTypes.ClusterTypeSkr: &clusterInfo{
				crdDirs: []string{dirSkr},
			},
			infraTypes.ClusterTypeGarden: &clusterInfo{
				crdDirs: []string{dirGarden},
			},
		},
	}

	for name, cluster := range infra.clusters {
		ginkgo.By(fmt.Sprintf("Startig cluster %s", name))
		sch, ok := infraScheme.SchemeMap[name]
		if !ok {
			return nil, fmt.Errorf("missing scheme for cluster %s", name)
		}

		env, cfg, err := startCluster(cluster.crdDirs, projectRoot, envtestK8sVersion)
		if err != nil {
			return nil, fmt.Errorf("error starting cluster %s: %w", name, err)
		}

		kubeconfigFilePath := filepath.Join(configDir, fmt.Sprintf("kubeconfig-%s", name))
		b, err := kubeconfigToBytes(restConfigToKubeconfig(cfg))
		if err != nil {
			return nil, fmt.Errorf("error getting %s kubeconfig bytes: %w", name, err)
		}
		err = os.WriteFile(kubeconfigFilePath, b, 0644)
		if err != nil {
			return nil, fmt.Errorf("error saving %s kubeconfig: %w", name, err)
		}

		k8sClient, err := ctrlclient.New(cfg, ctrlclient.Options{Scheme: sch})
		if err != nil {
			return nil, fmt.Errorf("error creating client for %s: %w", name, err)
		}

		cluster.env = env
		cluster.cfg = cfg
		cluster.kubeconfig = b
		cluster.kubeconfigFilePath = kubeconfigFilePath
		cluster.scheme = sch
		cluster.client = k8sClient

		ce := &clusterEnv{}
		switch name {
		case infraTypes.ClusterTypeKcp:
			ce.namespace = DefaultKcpNamespace
		case infraTypes.ClusterTypeSkr:
			ce.namespace = DefaultSkrNamespace
		case infraTypes.ClusterTypeGarden:
			ce.namespace = DefaultGardenNamespace
		}
		cluster.ClusterEnv = ce
	}

	skip := 0
	pc, _, _, ok := goruntime.Caller(skip + 1)
	if ok {
		details := goruntime.FuncForPC(pc)
		if details != nil {
			fullName := details.Name()
			for name, clsrt := range infra.clusters {
				c := clsrt.Client()
				cctx := context.Background()
				err := c.Create(cctx, &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "cm-info",
					},
					Data: map[string]string{
						"caller": fullName,
					},
				})
				if err != nil {
					return nil, fmt.Errorf("error creating cm-info configmap in testinfra cluster %q: %w", name, err)
				}
			}
		}
	}

	ginkgo.By("All started")

	// Create ENV
	kcpMgr, err := ctrl.NewManager(infra.clusters[infraTypes.ClusterTypeKcp].cfg, ctrl.Options{
		Scheme: infra.KCP().Scheme(),
		Client: ctrlclient.Options{
			Cache: &ctrlclient.CacheOptions{
				Unstructured: true,
			},
		},
		LeaderElection: false,
		Metrics: metricsserver.Options{
			BindAddress: "0", // disabled
		},
		HealthProbeBindAddress: "0", // disable
	})
	if err != nil {
		return nil, fmt.Errorf("error creating KCP manager: %w", err)
	}
	reader := kcpMgr.GetAPIReader()
	if reader == nil {
		return nil, errors.New("KCP Manager API Reader is nil")
	}

	registry := skrruntime.NewRegistry(infra.SKR().Scheme())
	activeSkrCollection := skrruntime.NewActiveSkrCollection()

	awsMock := awsmock.New()
	awsMock.SetAccount("some-aws-account")

	infra.InfraEnv = &infraEnv{
		i:                   infra,
		kcpManager:          kcpMgr,
		registry:            registry,
		activeSkrCollection: activeSkrCollection,
		awsMock:             awsMock,
		gcpMock:             gcpmock.New(),
		azureMock:           azuremock.New(),
		sapMock:             sapmock.New(),
		skrKymaRef: klog.ObjectRef{
			Name:      "5e32a9dd-4e68-47c7-aac7-64a4880a00d7",
			Namespace: infra.KCP().Namespace(),
		},
		config: config.NewConfig(abstractions.NewOSEnvironment()),
	}

	// Create DSL
	infra.InfraDSL = &infraDSL{i: infra}
	for _, c := range infra.clusters {
		c.ClusterDSL = &clusterDSL{
			ci:  c,
			ctx: infra.Ctx,
		}
	}

	_ = os.Setenv("GARDENER_NAMESPACE", infra.Garden().Namespace())

	// github.com/onsi/gomega@v1.29.0/internal/duration_bundle.go
	if debugged.Debugged {
		ginkgo.By("Setting high GOMEGA timeouts and durations since debug build flag is set!!!")
		gomega.Default.SetDefaultEventuallyTimeout(10 * time.Minute)
		gomega.Default.SetDefaultEventuallyPollingInterval(1 * time.Second)
		// rarely used and usually not debugged, so left small, increase on demand but do not commit
		gomega.Default.SetDefaultConsistentlyDuration(5 * time.Second)
		gomega.Default.SetDefaultConsistentlyPollingInterval(1 * time.Second)
	} else {
		gomega.Default.SetDefaultEventuallyTimeout(4 * time.Second)
		gomega.Default.SetDefaultEventuallyPollingInterval(200 * time.Millisecond)
		gomega.Default.SetDefaultConsistentlyDuration(2 * time.Second)
		gomega.Default.SetDefaultConsistentlyPollingInterval(200 * time.Millisecond)
	}

	//Setup GCP env variables
	_ = os.Setenv("GCP_SA_JSON_KEY_PATH", "test")
	_ = os.Setenv("GCP_RETRY_WAIT_DURATION", "300ms")
	_ = os.Setenv("GCP_OPERATION_WAIT_DURATION", "300ms")
	_ = os.Setenv("GCP_API_TIMEOUT_DURATION", "300ms")
	_ = os.Setenv("AWS_EFS_CAPACITY_CHECK_INTERVAL", "1s")
	_ = os.Setenv("GCP_CAPACITY_CHECK_INTERVAL", "1s")

	// init config
	commonconfig.LoadConfigInstance(infra.Config())
	fmt.Printf("Starting with config:\n%s\n", infra.Config().PrintJson())

	util.SetSpeedyTimingForTests()

	feature.InitializeFromStaticConfig(abstractions.NewOSEnvironment())

	return infra, nil
}

func startCluster(crdsDirs []string, projectRoot, envtestK8sVersion string) (*envtest.Environment, *rest.Config, error) {
	env := &envtest.Environment{
		CRDDirectoryPaths:     crdsDirs,
		ErrorIfCRDPathMissing: true,
		BinaryAssetsDirectory: filepath.Join(projectRoot, "bin", "k8s",
			fmt.Sprintf("%s-%s-%s", envtestK8sVersion, goruntime.GOOS, goruntime.GOARCH)),
		ControlPlaneStartTimeout: 30 * time.Second,
	}

	cfg, err := env.Start()

	return env, cfg, err
}
