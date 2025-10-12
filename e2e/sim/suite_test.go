package sim

import (
	"os"
	"testing"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	e2econfig "github.com/kyma-project/cloud-manager/e2e/config"
	"github.com/kyma-project/cloud-manager/e2e/sim/fixtures"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/config"
	"github.com/kyma-project/cloud-manager/pkg/testinfra"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"go.uber.org/zap/zapcore"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var infra testinfra.Infra

var simInstance Sim

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
	simInstance, err = New(infra.Ctx(), CreateOptions{
		KCP:                infra.KcpManager(),
		Garden:             infra.Garden().Client(),
		Logger:             infra.KcpManager().GetLogger(),
		CloudProfileLoader: NewCachingCloudProfileLoader(NewGardenCloudProfileLoader(infra.Garden().Client(), infra.Garden().Namespace())),
	})
	Expect(err).NotTo(HaveOccurred(), "failed creating sim")
	err = infra.KcpManager().Add(simInstance)
	Expect(err).NotTo(HaveOccurred(), "failed to add simInstance to KCP manager")

	// Start controllers
	infra.StartKcpControllers(infra.Ctx())

	infra.KcpManager().GetCache().WaitForCacheSync(infra.Ctx())

	By("creating garden namespace")
	// create garden namespace
	err = infra.Garden().Client().Create(infra.Ctx(), &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: e2econfig.Config.GardenNamespace,
		},
	})
	Expect(client.IgnoreAlreadyExists(err)).NotTo(HaveOccurred(), "failed to create garden namespace")
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")

	err := testinfra.PrintMetrics()
	Expect(err).NotTo(HaveOccurred())

	err = infra.Stop()
	Expect(err).NotTo(HaveOccurred())
})

var _ = ReportAfterSuite("gherkin report", testinfra.ReportAfterSuite)
