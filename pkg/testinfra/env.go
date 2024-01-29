package testinfra

import (
	"context"
	awsmock "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/mock"
	skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"time"
)

var _ InfraEnv = &infraEnv{}

type infraEnv struct {
	kcpManager manager.Manager
	registry   skrruntime.SkrRegistry
	looper     skrruntime.SkrLooper
	awsMock    awsmock.Server

	ctx    context.Context
	cancel context.CancelFunc
}

func (ie *infraEnv) KcpManager() manager.Manager {
	return ie.kcpManager
}

func (ie *infraEnv) Registry() skrruntime.SkrRegistry {
	return ie.registry
}

func (ie *infraEnv) Looper() skrruntime.SkrLooper {
	return ie.looper
}

func (ie *infraEnv) AwsMock() awsmock.Server {
	return ie.awsMock
}

func (ie *infraEnv) StartControllers(ctx context.Context) {
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
	time.Sleep(10 * time.Millisecond)
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
