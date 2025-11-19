package sim

import (
	"context"
	"os"
	"testing"
	"time"

	e2econfig "github.com/kyma-project/cloud-manager/e2e/config"
	e2ekeb "github.com/kyma-project/cloud-manager/e2e/keb"
	e2elib "github.com/kyma-project/cloud-manager/e2e/lib"
	"github.com/kyma-project/cloud-manager/e2e/lib/fixtures"
	"github.com/kyma-project/cloud-manager/pkg/testinfra"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"go.uber.org/zap/zapcore"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/clock"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var infra testinfra.Infra

var simInstance Sim
var kebInstance e2ekeb.Keb
var skrKubeconfigProviderInstance e2elib.SkrKubeconfigProviderWithCallCount
var config *e2econfig.ConfigType

func TestSimControllers(t *testing.T) {
	if len(os.Getenv("PROJECTROOT")) == 0 {
		t.Skip("Skipping TestControllers since PROJECTROOT env var is not set. It should point to dir where Makefile is. Check `make test` for details.")
		return
	}
	RegisterFailHandler(Fail)

	RunSpecs(t, "KCP Controller Suite")

}

var _ = BeforeSuite(func() {

	config = e2econfig.Stub()

	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true), zap.ConsoleEncoder(func(config *zapcore.EncoderConfig) {
		config.EncodeTime = zapcore.RFC3339NanoTimeEncoder
	})))

	By("bootstrapping KCP test environment")
	var err error
	infra, err = testinfra.Start()
	Expect(err).
		NotTo(HaveOccurred(), "failed starting infra clusters")

	Expect(infra.KCP().GivenNamespaceExists(infra.KCP().Namespace())).
		NotTo(HaveOccurred(), "failed creating namespace %s in KCP", infra.KCP().Namespace())
	Expect(infra.SKR().GivenNamespaceExists(infra.SKR().Namespace())).
		NotTo(HaveOccurred(), "failed creating namespace %s in SKR", infra.SKR().Namespace())
	Expect(infra.Garden().GivenNamespaceExists(infra.Garden().Namespace())).
		NotTo(HaveOccurred(), "failed creating namespace %s in Garden", infra.Garden().Namespace())

	By("installing cloud profiles")
	cloudProfilesArr, err := fixtures.CloudProfiles(infra.Garden().Namespace())
	Expect(err).NotTo(HaveOccurred(), "failed to load cloud profiles fixtures")
	err = util.Apply(infra.Ctx(), infra.Garden().Client(), cloudProfilesArr)
	Expect(err).NotTo(HaveOccurred(), "failed to apply cloud profiles fixtures")

	By("starting controllers")
	// Setup controllers
	cpl := e2elib.NewFileCloudProfileLoader(e2elib.CloudProfilesFS, "cloudprofiles.yaml", config)
	skrKubeconfigProviderInstance = e2elib.NewFixedSkrKubeconfigProvider(infra.SKR().Kubeconfig()).(e2elib.SkrKubeconfigProviderWithCallCount)
	skrManagerFactory := e2ekeb.NewSkrManagerFactory(infra.KcpManager().GetClient(), clock.RealClock{}, config.KcpNamespace)
	simInstance, err = New(CreateOptions{
		Config:                config,
		StartCtx:              infra.Ctx(),
		KcpManager:            infra.KcpManager(),
		GardenClient:          infra.Garden().Client(),
		Logger:                infra.KcpManager().GetLogger(),
		CloudProfileLoader:    cpl,
		SkrKubeconfigProvider: skrKubeconfigProviderInstance,
		SkrManagerFactory:     skrManagerFactory,
	})
	Expect(err).NotTo(HaveOccurred(), "failed creating sim")

	kebInstance = e2ekeb.NewKeb(infra.KCP().Client(), infra.Garden().Client(), skrManagerFactory, cpl, skrKubeconfigProviderInstance, config)

	// add shoot ready controller that makes each shoot ready!!!
	gardenManager, err := NewGardenManager(infra.Garden().Cfg(), infra.Garden().Scheme(), infra.KcpManager().GetLogger())
	Expect(err).NotTo(HaveOccurred(), "failed creating garden manager")
	Expect(SetupShootReadyController(gardenManager)).To(Succeed())
	Expect(infra.KcpManager().Add(gardenManager)).To(Succeed())

	// Start controllers
	infra.StartKcpControllers(infra.Ctx())
	Expect(infra.KcpWaitForCacheSync(infra.Ctx())).To(Succeed())

	toCtx, toCancel := context.WithTimeout(infra.Ctx(), time.Second*10)
	defer toCancel()
	ok := gardenManager.GetCache().WaitForCacheSync(toCtx)
	Expect(toCtx.Err()).ToNot(HaveOccurred())
	Expect(ok).To(BeTrue())

	By("creating garden namespace")
	// create garden namespace
	err = infra.Garden().Client().Create(infra.Ctx(), &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: config.GardenNamespace,
		},
	})
	Expect(client.IgnoreAlreadyExists(err)).NotTo(HaveOccurred(), "failed to create garden namespace")

	Default.SetDefaultEventuallyTimeout(8 * time.Second)
	Default.SetDefaultEventuallyPollingInterval(200 * time.Millisecond)
	Default.SetDefaultConsistentlyDuration(2 * time.Second)
	Default.SetDefaultConsistentlyPollingInterval(200 * time.Millisecond)
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")

	err := testinfra.PrintMetrics()
	Expect(err).NotTo(HaveOccurred())

	err = infra.Stop()
	Expect(err).NotTo(HaveOccurred())
})

var _ = ReportAfterSuite("gherkin report", testinfra.ReportAfterSuite)
