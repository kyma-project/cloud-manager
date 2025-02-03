package testinfra

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/pkg/config"
	awsmock "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/mock"
	azuremock "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/mock"
	cceemock "github.com/kyma-project/cloud-manager/pkg/kcp/provider/ccee/mock"
	gcpmock "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/mock"
	skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime"
	"github.com/kyma-project/cloud-manager/pkg/skr/runtime/looper"
	skrmanager "github.com/kyma-project/cloud-manager/pkg/skr/runtime/manager"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"time"
)

var _ InfraEnv = &infraEnv{}

type infraEnv struct {
	i                   Infra
	kcpManager          manager.Manager
	registry            skrruntime.SkrRegistry
	activeSkrCollection skrruntime.ActiveSkrCollection
	awsMock             awsmock.Server
	gcpMock             gcpmock.Server
	azureMock           azuremock.Server
	cceeMock            cceemock.Server
	skrKymaRef          klog.ObjectRef
	skrManager          skrmanager.SkrManager
	runner              skrruntime.SkrRunner
	config              config.Config

	ctx    context.Context
	cancel context.CancelFunc
}

func (ie *infraEnv) KcpManager() manager.Manager {
	return ie.kcpManager
}

func (ie *infraEnv) Registry() skrruntime.SkrRegistry {
	return ie.registry
}

func (ie *infraEnv) ActiveSkrCollection() skrruntime.ActiveSkrCollection {
	return ie.activeSkrCollection
}

func (ie *infraEnv) AwsMock() awsmock.Server {
	return ie.awsMock
}

func (ie *infraEnv) GcpMock() gcpmock.Server {
	return ie.gcpMock
}

func (ie *infraEnv) AzureMock() azuremock.Server { return ie.azureMock }

func (ie *infraEnv) CceeMock() cceemock.Server {
	return ie.cceeMock
}

func (ie *infraEnv) SkrKymaRef() klog.ObjectRef {
	return ie.skrKymaRef
}

func (ie *infraEnv) SkrRunner() skrruntime.SkrRunner {
	return ie.runner
}

func (ie *infraEnv) Config() config.Config {
	return ie.config
}

func (ie *infraEnv) StartKcpControllers(ctx context.Context) {
	ginkgo.By("Starting controllers")
	if ctx == nil {
		ctx = context.Background()
	}
	ie.ctx, ie.cancel = context.WithCancel(ctx)
	go func() {
		defer ginkgo.GinkgoRecover()
		err := ie.kcpManager.Start(ie.ctx)
		gomega.Expect(err).NotTo(gomega.HaveOccurred(), "failed to run kcp manager")
	}()
	time.Sleep(time.Second)
}

func (ie *infraEnv) StartSkrControllers(ctx context.Context) {
	ie.kcpManager.GetLogger().Info("TestInfra: StartSkrControllers")
	var err error
	ie.skrManager, err = skrmanager.New(ie.i.SKR().Cfg(), ie.i.SKR().Scheme(), ie.skrKymaRef, ie.kcpManager.GetLogger())
	if err != nil {
		panic(fmt.Errorf("error creating SKR manager: %w", err))
	}

	ie.runner = skrruntime.NewSkrRunnerWithNoopStatusSaver(ie.registry, newKcpClusterWrap(ie.kcpManager), ie.skrKymaRef.Name)
	ie.ctx, ie.cancel = context.WithCancel(ctx)
	go func() {
		defer ginkgo.GinkgoRecover()
		err = ie.runner.Run(ie.ctx, ie.skrManager, looper.WithTimeout(10*time.Minute), looper.WithoutProvider())
		if err != nil {
			ie.skrManager.GetLogger().Error(err, "Error running SKR Runner")
		}
	}()
	time.Sleep(time.Second)
}

func (ie *infraEnv) Ctx() context.Context {
	if ie.ctx == nil {
		return context.Background()
	}
	return ie.ctx
}

func (ie *infraEnv) stopControllers() {
	if ie.cancel != nil {
		ginkgo.By("Stopping controllers")
		ie.cancel()
	}
}

// =========================================

var _ ClusterEnv = &clusterEnv{}

type clusterEnv struct {
	namespace string
}

func (ce *clusterEnv) Namespace() string {
	return ce.namespace
}

func (ce *clusterEnv) ObjKey(name string) types.NamespacedName {
	return types.NamespacedName{
		Namespace: ce.namespace,
		Name:      name,
	}
}
