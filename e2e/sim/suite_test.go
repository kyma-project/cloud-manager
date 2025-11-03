package sim

import (
	"context"
	"os"
	"path"
	"testing"
	"time"

	e2econfig "github.com/kyma-project/cloud-manager/e2e/config"
	"github.com/kyma-project/cloud-manager/e2e/sim/fixtures"
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
var skrKubeconfigProviderInstance *fixedKubeconfigProvider
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
	skrKubeconfigProviderInstance = NewFixedSkrKubeconfigProvider(infra.SKR().Kubeconfig()).(*fixedKubeconfigProvider)
	simInstance, err = New(CreateOptions{
		Config:                config,
		StartCtx:              infra.Ctx(),
		KcpManager:            infra.KcpManager(),
		Garden:                infra.Garden().Client(),
		GardenApiReader:       infra.Garden().Client(),
		Logger:                infra.KcpManager().GetLogger(),
		CloudProfileLoader:    NewFileCloudProfileLoader(path.Join(infra.ProjectRootDir(), "e2e/sim/fixtures/cloudprofiles.yaml"), config),
		SkrKubeconfigProvider: skrKubeconfigProviderInstance,
	})
	Expect(err).NotTo(HaveOccurred(), "failed creating sim")

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
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")

	err := testinfra.PrintMetrics()
	Expect(err).NotTo(HaveOccurred())

	err = infra.Stop()
	Expect(err).NotTo(HaveOccurred())
})

var _ = ReportAfterSuite("gherkin report", testinfra.ReportAfterSuite)
