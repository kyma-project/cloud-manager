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

package azure

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"os"
	"testing"

	cloudresourcescontroller "github.com/kyma-project/cloud-manager/internal/controller/cloud-resources"
	"github.com/kyma-project/cloud-manager/pkg/testinfra"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

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

	RunSpecs(t, "SKR Controller Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	By("bootstrapping SKR test environment")
	var err error
	infra, err = testinfra.Start()
	env := abstractions.NewOSEnvironment()
	Expect(err).
		NotTo(HaveOccurred(), "failed starting infra clusters")

	Expect(infra.KCP().GivenNamespaceExists(infra.KCP().Namespace())).
		NotTo(HaveOccurred(), "failed creating namespace %s in KCP", infra.KCP().Namespace())
	Expect(infra.SKR().GivenNamespaceExists(infra.SKR().Namespace())).
		NotTo(HaveOccurred(), "failed creating namespace %s in SKR", infra.SKR().Namespace())
	Expect(infra.Garden().GivenNamespaceExists(infra.Garden().Namespace())).
		NotTo(HaveOccurred(), "failed creating namespace %s in Garden", infra.Garden().Namespace())

	// Setup controllers
	// Test Only PV Controller
	Expect(testinfra.SetupPvController(infra.Registry())).
		NotTo(HaveOccurred())
	// Test Only PVC Controller
	Expect(testinfra.SetupPVCController(infra.Registry())).
		NotTo(HaveOccurred())

	// AzureRwxBackupSchedule
	Expect(cloudresourcescontroller.SetupAzureRwxBackupScheduleReconciler(
		infra.Registry(), env)).NotTo(HaveOccurred())

	// AzureRwxVolumeRestore
	Expect(cloudresourcescontroller.SetupAzureRwxRestoreReconciler(
		infra.Registry(), infra.AzureMock().StorageProvider())).NotTo(HaveOccurred())

	// AzureRwxVolumeBackup
	// TODO: confirm if .NotTo(HaveOccurred()) is necessary
	Expect(cloudresourcescontroller.SetupAzureRwxBackupReconciler(infra.Registry(), infra.AzureMock().StorageProvider()))

	// Start controllers
	infra.StartSkrControllers(context.Background())
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")

	err := testinfra.PrintMetrics()
	Expect(err).NotTo(HaveOccurred())

	err = infra.Stop()
	Expect(err).NotTo(HaveOccurred())
})

var _ = ReportAfterSuite("gherkin report", testinfra.ReportAfterSuite)
