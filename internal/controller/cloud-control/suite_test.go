/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cloudcontrol

import (
	"context"
	"os"
	"testing"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	awsnukeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/nuke/client"
	azurenukeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/nuke/client"
	"github.com/kyma-project/cloud-manager/pkg/testinfra"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"go.uber.org/zap/zapcore"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var infra testinfra.Infra

func TestControllers(t *testing.T) {
	if len(os.Getenv("PROJECTROOT")) == 0 {
		t.Skip("Skipping TestControllers since PROJECTROOT env var is not set. It should point to dir where Makefile is. Check `make test` for details.")
		return
	}
	RegisterFailHandler(Fail)

	RunSpecs(t, "KCP Controller Suite")
}

var _ = BeforeSuite(func() {
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
	Expect(infra.SKR().GivenNamespaceExists("kyma-system")).
		NotTo(HaveOccurred(), "failed creating namespace kyma-system in SKR")
	Expect(infra.Garden().GivenNamespaceExists(infra.Garden().Namespace())).
		NotTo(HaveOccurred(), "failed creating namespace %s in Garden", infra.Garden().Namespace())

	// Setup environment variables
	env := abstractions.NewMockedEnvironment(map[string]string{})

	cloudcontrolv1beta1.AutomaticNuke = false

	// Setup controllers
	// Scope
	Expect(SetupScopeReconciler(
		infra.Ctx(),
		infra.KcpManager(),
		infra.AwsMock().ScopeGardenProvider(),
		infra.ActiveSkrCollection(),
		infra.GcpMock().ServiceUsageClientProvider(),
		infra.AwsMock().ExposedDataProvider(),
		infra.AzureMock().ExposeDataProvider(),
		infra.GcpMock().ExposedDataProvider(),
		infra.SapMock().ExposedDataProvider(),
	)).NotTo(HaveOccurred())
	// Kyma
	Expect(SetupKymaReconciler(
		infra.KcpManager(),
		infra.ActiveSkrCollection(),
	)).NotTo(HaveOccurred())
	// IpRange
	Expect(SetupIpRangeReconciler(
		infra.Ctx(),
		infra.KcpManager(),
		infra.AwsMock().IpRangeSkrProvider(),
		infra.AzureMock().IpRangeProvider(),
		infra.GcpMock().ServiceNetworkingClientProvider(),
		infra.GcpMock().ComputeClientProvider(),
		infra.SapMock().IpRangeProvider(),
		env,
	)).NotTo(HaveOccurred())
	// NfsInstance
	Expect(SetupNfsInstanceReconciler(
		infra.KcpManager(),
		infra.AwsMock().NfsInstanceSkrProvider(),
		infra.GcpMock().FilestoreClientProvider(),
		infra.SapMock().NfsInstanceProvider(),
		env,
	)).NotTo(HaveOccurred())
	//VpcPeering
	Expect(SetupVpcPeeringReconciler(
		infra.KcpManager(),
		infra.AwsMock().VpcPeeringSkrProvider(),
		infra.AzureMock().VpcPeeringProvider(),
		infra.GcpMock().VpcPeeringProvider(),
		env,
	)).NotTo(HaveOccurred())
	// RedisInstance
	Expect(SetupRedisInstanceReconciler(
		infra.KcpManager(),
		infra.GcpMock().MemoryStoreProviderFake(),
		infra.AzureMock().RedisClientProvider(),
		infra.AwsMock().ElastiCacheProviderFake(),
		env,
	)).NotTo(HaveOccurred())
	// RedisCluster
	Expect(SetupRedisClusterReconciler(
		infra.KcpManager(),
		infra.AwsMock().ElastiCacheProviderFake(),
		infra.AzureMock().RedisClusterClientProvider(),
		env,
	)).NotTo(HaveOccurred())
	Expect(SetupGcpRedisClusterReconciler(
		infra.KcpManager(),
		infra.GcpMock().MemoryStoreClusterProviderFake(),
		env,
	)).NotTo(HaveOccurred())
	// Network
	Expect(SetupNetworkReconciler(
		infra.Ctx(),
		infra.KcpManager(),
		infra.AzureMock().NetworkProvider(),
	)).NotTo(HaveOccurred())
	// Nuke
	Expect(SetupNukeReconciler(
		infra.KcpManager(),
		infra.ActiveSkrCollection(),
		infra.GcpMock().FileBackupClientProvider(),
		awsnukeclient.Mock(),
		azurenukeclient.NukeProvider(infra.AzureMock().StorageProvider()),
		env,
	)).To(Succeed())
	// GcpSubnet
	Expect(SetupGcpSubnetReconciler(
		infra.Ctx(),
		infra.KcpManager(),
		infra.GcpMock().SubnetComputeClientProvider(),
		infra.GcpMock().SubnetNetworkConnectivityProvider(),
		infra.GcpMock().SubnetRegionOperationsClientProvider(),
		env,
	)).To(Succeed())
	//AzureVNetLink
	Expect(SetupAzureVNetLinkReconciler(
		infra.KcpManager(),
		infra.AzureMock().DnsZoneVNetLinkProvider(),
		infra.AzureMock().DnsResolverVNetLinkProvider(),
	)).NotTo(HaveOccurred())
	// Subscription
	Expect(SetupSubscriptionReconciler(
		infra.KcpManager(),
		infra.AwsMock().SubscriptionGardenProvider(),
	)).To(Succeed())

	// Start controllers
	infra.StartKcpControllers(context.Background())
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")

	err := testinfra.PrintMetrics()
	Expect(err).NotTo(HaveOccurred())

	err = infra.Stop()
	Expect(err).NotTo(HaveOccurred())
})

var _ = ReportAfterSuite("gherkin report", testinfra.ReportAfterSuite)
