package cloudresources

import (
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	"github.com/kyma-project/cloud-manager/pkg/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/api/file/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Feature: SKR GcpNfsVolumeBackupDiscovery", func() {

	It("Scenario: SKR GcpNfsVolumeBackupDiscovery is created", func() {

		scope := &cloudcontrolv1beta1.Scope{}
		By("Given KCP Scope exists", func() {
			// Given Scope exists
			Expect(
				infra.GivenScopeGcpExists(infra.SkrKymaRef().Name),
			).NotTo(HaveOccurred())
			// Load created scope
			Eventually(func() (exists bool, err error) {
				err = infra.KCP().Client().Get(infra.Ctx(), infra.KCP().ObjKey(infra.SkrKymaRef().Name), scope)
				exists = err == nil
				return exists, client.IgnoreNotFound(err)
			}).
				Should(BeTrue(), "expected Scope to get created")
		})

		By("And Given shared backups exist for shoot", func() {
			infra.GcpMock().CreateFakeBackup(&file.Backup{
				Name:               "projects/test-project/locations/us-central1-a/backups/nfs-backup-1",
				Description:        "Test NFS volume backup 1",
				State:              "READY",
				CreateTime:         "2024-10-30T10:00:00Z",
				SourceFileShare:    "nfs-share-1",
				SourceInstance:     "projects/test-project/locations/us-central1-a/instances/nfs-instance-1",
				SourceInstanceTier: "STANDARD",
				Labels: map[string]string{
					"managed-by":                                     "cloud-manager",
					"scope-name":                                     "test-scope",
					util.GcpLabelSkrVolumeName:                       "test-volume-1",
					util.GcpLabelSkrVolumeNamespace:                  "default",
					util.GcpLabelSkrBackupName:                       "test-backup-1",
					util.GcpLabelSkrBackupNamespace:                  "default",
					util.GcpLabelShootName:                           "test-shoot",
					fmt.Sprintf("cm-allow-%s", scope.Spec.ShootName): util.GcpLabelBackupAccessibleFrom,
					"cm-allow-shoot2":                                util.GcpLabelBackupAccessibleFrom,
				},
				CapacityGb:   100,
				StorageBytes: 107374182400, // 100 GB in bytes
			})
			infra.GcpMock().CreateFakeBackup(&file.Backup{
				Name:               "projects/test-project/locations/us-central1-a/backups/nfs-backup-2",
				Description:        "Test NFS volume backup 2",
				State:              "READY",
				CreateTime:         "2024-10-30T11:00:00Z",
				SourceFileShare:    "nfs-share-2",
				SourceInstance:     "projects/test-project/locations/us-central1-a/instances/nfs-instance-2",
				SourceInstanceTier: "PREMIUM",
				Labels: map[string]string{
					"managed-by":                                     "cloud-manager",
					"scope-name":                                     "test-scope",
					util.GcpLabelSkrVolumeName:                       "test-volume-2",
					util.GcpLabelSkrVolumeNamespace:                  "default",
					util.GcpLabelSkrBackupName:                       "test-backup-2",
					util.GcpLabelSkrBackupNamespace:                  "default",
					util.GcpLabelShootName:                           "test-shoot",
					fmt.Sprintf("cm-allow-%s", scope.Spec.ShootName): util.GcpLabelBackupAccessibleFrom,
					"cm-allow-shoot1":                                util.GcpLabelBackupAccessibleFrom,
				},
				CapacityGb:   200,
				StorageBytes: 214748364800, // 200 GB in bytes
			})
		})

		gcpNfsVolumeBackupDiscoveryName := "c0e295ea-4c5b-4c42-ac62-66600d37b32e"
		gcpNfsVolumeBackupDiscovery := &cloudresourcesv1beta1.GcpNfsVolumeBackupDiscovery{}

		By("When GcpNfsVolumeBackupDiscovery is created", func() {
			Eventually(CreateGcpNfsVolumeBackupDiscovery).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeBackupDiscovery,
					WithName(gcpNfsVolumeBackupDiscoveryName),
				).
				Should(Succeed())
		})

		By("Then GcpNfsVolumeBackupDiscovery is created in SKR", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeBackupDiscovery,
					NewObjActions(),
				).
				Should(Succeed())
		})

		By("And Then GcpNfsVolumeBackupDiscovery will get Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeBackupDiscovery,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
				).
				Should(Succeed())
		})

		By("And Then GcpNfsVolumeBackupDiscovery status will be Ready", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeBackupDiscovery,
					NewObjActions(),
					HavingState(cloudresourcesv1beta1.JobStateDone),
				).
				Should(Succeed())
		})

		By("And Then GcpNfsVolumeBackupDiscovery status fields should be populated", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeBackupDiscovery,
					NewObjActions(),
					AssertGcpNfsVolumeBackupDiscoveryStatusPopulated(),
					AssertGcpNfsVolumeBackupDiscoveryAvailableBackupsPopulated(),
				).
				Should(Succeed())
		})

		By("And Then GcpNfsVolumeBackupDiscovery should have non-negative backup count", func() {
			Expect(gcpNfsVolumeBackupDiscovery.Status.AvailableBackupsCount).NotTo(BeNil())
			Expect(*gcpNfsVolumeBackupDiscovery.Status.AvailableBackupsCount).To(BeNumerically(">=", 0))
		})

		// CleanUp
		infra.GcpMock().ClearAllBackups()

		Eventually(Delete).
			WithArguments(infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeBackupDiscovery).
			Should(Succeed())
	})

})
