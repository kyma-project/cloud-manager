package ctrltest

import (
	"os"
	"testing"
	"time"

	"github.com/kyma-project/cloud-manager/e2e"
	e2econfig "github.com/kyma-project/cloud-manager/e2e/config"
	e2elib "github.com/kyma-project/cloud-manager/e2e/lib"
	"github.com/kyma-project/cloud-manager/e2e/lib/fixtures"
	"github.com/kyma-project/cloud-manager/e2e/sim"
	"github.com/kyma-project/cloud-manager/pkg/testinfra"
	"github.com/kyma-project/cloud-manager/pkg/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap/zapcore"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var infra testinfra.Infra

var world e2e.WorldIntf
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
	Expect(infra.Garden().GivenNamespaceExists(config.GardenNamespace)).
		NotTo(HaveOccurred(), "failed creating namespace %s in Garden", config.GardenNamespace)

	config.GardenKubeconfig = infra.Garden().KubeconfigFilePath()

	By("installing cloud profiles")
	cloudProfilesArr, err := fixtures.CloudProfiles(infra.Garden().Namespace())
	Expect(err).NotTo(HaveOccurred(), "failed to load cloud profiles fixtures")
	err = util.Apply(infra.Ctx(), infra.Garden().Client(), cloudProfilesArr)
	Expect(err).NotTo(HaveOccurred(), "failed to apply cloud profiles fixtures")

	By("starting world")

	wf := e2e.NewWorldFactory()
	w, err := wf.Create(infra.Ctx(), e2e.WorldCreateOptions{
		Config:                config,
		KcpRestConfig:         infra.KCP().Cfg(),
		CloudProfileLoader:    e2elib.NewFileCloudProfileLoader(e2elib.CloudProfilesFS, "cloudprofiles.yaml", config),
		SkrKubeconfigProvider: e2elib.NewFixedSkrKubeconfigProvider(infra.SKR().Kubeconfig()),
	})
	Expect(err).NotTo(HaveOccurred(), "failed creating the world")

	// add shoot ready controller that makes each shoot ready!!!
	Expect(sim.SetupShootReadyController(w.GardenManager())).To(Succeed())

	world = w

	// Start infra
	infra.StartKcpControllers(infra.Ctx())
	Expect(infra.KcpWaitForCacheSync(infra.Ctx())).To(Succeed())

	// it takes more time for SKR manager to stop, so giving it x2 time
	Default.SetDefaultEventuallyTimeout(8 * time.Second)
})

var _ = AfterSuite(func() {
	By("stopping world")

	err := testinfra.PrintMetrics()
	Expect(err).NotTo(HaveOccurred())

	world.Cancel()
	world.StopWaitGroup().Wait()
	err = world.RunError()
	Expect(err).NotTo(HaveOccurred())

	By("tearing down the test environment")
	err = infra.Stop()
	Expect(err).NotTo(HaveOccurred())
})

var _ = ReportAfterSuite("gherkin report", testinfra.ReportAfterSuite)
