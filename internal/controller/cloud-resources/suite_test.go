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

package cloudresources

import (
	"context"
	"os"
	"testing"

	"github.com/kyma-project/cloud-manager/pkg/migrateFinalizers"

	iprangeallocate "github.com/kyma-project/cloud-manager/pkg/kcp/iprange/allocate"

	ctrl "sigs.k8s.io/controller-runtime"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/quota"
	"github.com/kyma-project/cloud-manager/pkg/testinfra"
	testinfradsl "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var infra testinfra.Infra

var addressSpace = iprangeallocate.NewAddressSpace()

func TestControllers(t *testing.T) {
	if len(os.Getenv("PROJECTROOT")) == 0 {
		t.Skip("Skipping TestControllers since PROJECTROOT env var is not set. It should point to dir where Makefile is. Check `make test` for details.")
		return
	}
	RegisterFailHandler(Fail)

	RunSpecs(t, "SKR Controller Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	By("bootstrapping SKR test environment")
	var err error
	infra, err = testinfra.Start()
	Expect(err).
		NotTo(HaveOccurred(), "failed starting infra clusters")

	Expect(infra.KCP().GivenNamespaceExists(infra.KCP().Namespace())).
		NotTo(HaveOccurred(), "failed creating namespace %s in KCP", infra.KCP().Namespace())
	Expect(infra.KCP().GivenNamespaceExists(testinfradsl.DefaultKcpNamespace)).
		NotTo(HaveOccurred(), "failed creating namespace %s in KCP", infra.KCP().Namespace())
	Expect(infra.SKR().GivenNamespaceExists(infra.SKR().Namespace())).
		NotTo(HaveOccurred(), "failed creating namespace %s in SKR", infra.SKR().Namespace())
	Expect(infra.SKR().GivenNamespaceExists(testinfradsl.DefaultSkrNamespace)).
		NotTo(HaveOccurred(), "failed creating namespace %s in SKR", infra.SKR().Namespace())
	Expect(infra.Garden().GivenNamespaceExists(infra.Garden().Namespace())).
		NotTo(HaveOccurred(), "failed creating namespace %s in Garden", infra.Garden().Namespace())

	// Quota override
	quota.SkrQuota.Override(&cloudresourcesv1beta1.IpRange{}, infra.SKR().Scheme(), "", 1000)

	// Setup environment variables
	env := abstractions.NewMockedEnvironment(map[string]string{})

	testSetupLog := ctrl.Log.WithName("testSetup")

	// Setup controllers
	// Test Only PV Controller
	Expect(testinfra.SetupPvController(infra.Registry())).
		NotTo(HaveOccurred())
	// Test Only PVC Controller
	Expect(testinfra.SetupPVCController(infra.Registry())).
		NotTo(HaveOccurred())
	// CloudResources
	Expect(SetupCloudResourcesReconciler(infra.Registry())).
		NotTo(HaveOccurred())
	// IpRange
	Expect(SetupIpRangeReconciler(infra.Registry())).
		NotTo(HaveOccurred())
	// AwsNfsVolume
	Expect(SetupAwsNfsVolumeReconciler(infra.Registry())).
		NotTo(HaveOccurred())
	// CceeNfsVolume
	Expect(SetupCceeNfsVolumeReconciler(infra.Registry())).
		NotTo(HaveOccurred())
	// GcpNfsVolume
	Expect(SetupGcpNfsVolumeReconciler(infra.Registry())).
		NotTo(HaveOccurred())
	// GcpNfsVolumeRestore
	Expect(SetupGcpNfsVolumeRestoreReconciler(infra.Registry(), infra.GcpMock().FilerestoreClientProvider(), env, testSetupLog)).
		NotTo(HaveOccurred())
	// GcpNfsVolumeBackup
	Expect(SetupGcpNfsVolumeBackupReconciler(infra.Registry(), infra.GcpMock().FileBackupClientProvider(), env, testSetupLog)).
		NotTo(HaveOccurred())
	// GcpRedisInstance
	Expect(SetupGcpRedisInstanceReconciler(infra.Registry())).
		NotTo(HaveOccurred())
	// AwsRedisInstance
	Expect(SetupAwsRedisInstanceReconciler(infra.Registry())).
		NotTo(HaveOccurred())
	// AwsRedisCluster
	Expect(SetupAwsRedisClusterReconciler(infra.Registry())).
		NotTo(HaveOccurred())
	// AzureRedisInstance
	Expect(SetupAzureRedisInstanceReconciler(infra.Registry())).
		NotTo(HaveOccurred())
	// NfsBackupSchedule
	Expect(SetupGcpNfsBackupScheduleReconciler(infra.Registry(), env)).NotTo(HaveOccurred())

	// AzureVpcPeering
	Expect(SetupAzureVpcPeeringReconciler(infra.Registry()))

	// AwsVpcPeering
	Expect(SetupAwsVpcPeeringReconciler(infra.Registry()))

	Expect(addressSpace.Reserve("10.128.0.0/10")).NotTo(HaveOccurred())

	//GCP Vpc Peering
	Expect(SetupGcpVpcPeeringReconciler(infra.Registry())).NotTo(HaveOccurred())

	migrateFinalizers.RunMigration = false

	// Start controllers
	infra.StartSkrControllers(context.Background())
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := infra.Stop()
	Expect(err).NotTo(HaveOccurred())
})

var _ = ReportAfterSuite("gherkin report", testinfra.ReportAfterSuite)
