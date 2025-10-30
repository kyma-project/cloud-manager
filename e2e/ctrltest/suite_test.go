package ctrltest

import (
	"fmt"
	"os"
	"path"
	"testing"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/e2e"
	e2econfig "github.com/kyma-project/cloud-manager/e2e/config"
	"github.com/kyma-project/cloud-manager/e2e/sim"
	"github.com/kyma-project/cloud-manager/e2e/sim/fixtures"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/config"
	"github.com/kyma-project/cloud-manager/pkg/testinfra"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"go.uber.org/zap/zapcore"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var infra testinfra.Infra

var world e2e.WorldIntf

func TestSimControllers(t *testing.T) {
	if len(os.Getenv("PROJECTROOT")) == 0 {
		t.Skip("Skipping TestControllers since PROJECTROOT env var is not set. It should point to dir where Makefile is. Check `make test` for details.")
		return
	}
	RegisterFailHandler(Fail)

	RunSpecs(t, "KCP Controller Suite")

}

var _ = BeforeSuite(func() {

	cfg := config.NewConfig(abstractions.NewMockedEnvironment(map[string]string{}))
	e2econfig.InitConfig(cfg)
	cfg.Read()

	e2econfig.Config.OidcClientId = "79221501-5dcc-4285-9af6-d023f313918e"
	e2econfig.Config.OidcIssuerUrl = "https://oidc.e2e.cloud-manager.kyma.local"
	e2econfig.Config.Administrators = []string{"admin@e2e.cloud-manager.kyma.local"}
	e2econfig.Config.Subscriptions = e2econfig.Subscriptions{
		{
			Name:     "aws",
			Provider: cloudcontrolv1beta1.ProviderAws,
		},
		{
			Name:     "gcp",
			Provider: cloudcontrolv1beta1.ProviderGCP,
		},
		{
			Name:     "azure",
			Provider: cloudcontrolv1beta1.ProviderAzure,
		},
		{
			Name:     "openstack",
			Provider: cloudcontrolv1beta1.ProviderOpenStack,
		},
	}

	fmt.Printf("e2econfig:\n%s\n", cfg.PrintJson())

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
	Expect(infra.Garden().GivenNamespaceExists(e2econfig.Config.GardenNamespace)).
		NotTo(HaveOccurred(), "failed creating namespace %s in Garden", e2econfig.Config.GardenNamespace)

	e2econfig.Config.GardenKubeconfig = infra.Garden().KubeconfigFilePath()

	By("installing cloud profiles")
	cloudProfilesArr, err := fixtures.CloudProfiles(infra.Garden().Namespace())
	Expect(err).NotTo(HaveOccurred(), "failed to load cloud profiles fixtures")
	err = util.Apply(infra.Ctx(), infra.Garden().Client(), cloudProfilesArr)
	Expect(err).NotTo(HaveOccurred(), "failed to apply cloud profiles fixtures")

	By("starting world")

	// add shoot ready controller that makes each shoot ready!!!
	gardenManager, err := sim.NewGardenManager(infra.Garden().Cfg(), infra.Garden().Scheme(), infra.KcpManager().GetLogger())
	Expect(err).NotTo(HaveOccurred(), "failed creating garden manager")
	Expect(sim.SetupShootReadyController(gardenManager)).To(Succeed())

	wf := e2e.NewWorldFactory()
	w, err := wf.Create(infra.Ctx(), e2e.WorldCreateOptions{
		KcpRestConfig:         infra.KCP().Cfg(),
		CloudProfileLoader:    sim.NewFileCloudProfileLoader(path.Join(infra.ProjectRootDir(), "e2e/sim/fixtures/cloudprofiles.yaml")),
		SkrKubeconfigProvider: sim.NewFixedSkrKubeconfigProvider(infra.SKR().Kubeconfig()),
		ExtraRunnables:        []manager.Runnable{gardenManager},
	})
	Expect(err).NotTo(HaveOccurred(), "failed creating the world")
	world = w

	// there's no need to start infra KCP, Garden and SKR clusters, since world has its own clusters
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
