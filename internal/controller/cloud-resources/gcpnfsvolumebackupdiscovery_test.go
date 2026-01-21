package cloudresources

import (
	"fmt"

	"github.com/google/uuid"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	"github.com/kyma-project/cloud-manager/pkg/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/api/file/v1"
)

var _ = Describe("Feature: SKR GcpNfsVolumeBackupDiscovery", func() {

	It("Scenario: SKR GcpNfsVolumeBackupDiscovery is created", func() {

		name := uuid.NewString()
		scopeName := infra.SkrKymaRef().Name
		shootName := scopeName // ShootName is always set to scope.Name in GivenScopeGcpExists

		By("Given KCP Scope exists", func() {
			Expect(
				infra.GivenScopeGcpExists(scopeName),
			).NotTo(HaveOccurred())
		})

		By("And Given shared backups exist for shoot", func() {
			infra.GcpMock().CreateFakeBackup(&file.Backup{
				Name:               fmt.Sprintf("projects/kyma/locations/us-central1-a/backups/%s-backup-1", name),
				Description:        fmt.Sprintf("Test NFS volume backup 1 for %s", name),
				State:              "READY",
				CreateTime:         "2024-10-30T10:00:00Z",
				SourceFileShare:    fmt.Sprintf("%s-share-1", name),
				SourceInstance:     fmt.Sprintf("projects/kyma/locations/us-central1-a/instances/%s-instance-1", name),
				SourceInstanceTier: "STANDARD",
				Labels: map[string]string{
					"managed-by":                          "cloud-manager",
					"scope-name":                          scopeName,
					util.GcpLabelSkrVolumeName:            fmt.Sprintf("%s-volume-1", name),
					util.GcpLabelSkrVolumeNamespace:       "default",
					util.GcpLabelSkrBackupName:            fmt.Sprintf("%s-backup-1", name),
					util.GcpLabelSkrBackupNamespace:       "default",
					util.GcpLabelShootName:                shootName,
					fmt.Sprintf("cm-allow-%s", shootName): util.GcpLabelBackupAccessibleFrom,
				},
				CapacityGb:   100,
				StorageBytes: 107374182400, // 100 GB in bytes
			})
			infra.GcpMock().CreateFakeBackup(&file.Backup{
				Name:               fmt.Sprintf("projects/kyma/locations/us-central1-a/backups/%s-backup-2", name),
				Description:        fmt.Sprintf("Test NFS volume backup 2 for %s", name),
				State:              "READY",
				CreateTime:         "2024-10-30T11:00:00Z",
				SourceFileShare:    fmt.Sprintf("%s-share-2", name),
				SourceInstance:     fmt.Sprintf("projects/kyma/locations/us-central1-a/instances/%s-instance-2", name),
				SourceInstanceTier: "PREMIUM",
				Labels: map[string]string{
					"managed-by":                          "cloud-manager",
					"scope-name":                          scopeName,
					util.GcpLabelSkrVolumeName:            fmt.Sprintf("%s-volume-2", name),
					util.GcpLabelSkrVolumeNamespace:       "default",
					util.GcpLabelSkrBackupName:            fmt.Sprintf("%s-backup-2", name),
					util.GcpLabelSkrBackupNamespace:       "default",
					util.GcpLabelShootName:                shootName,
					fmt.Sprintf("cm-allow-%s", shootName): util.GcpLabelBackupAccessibleFrom,
				},
				CapacityGb:   200,
				StorageBytes: 214748364800, // 200 GB in bytes
			})
		})

		gcpNfsVolumeBackupDiscovery := &cloudresourcesv1beta1.GcpNfsVolumeBackupDiscovery{}

		By("When GcpNfsVolumeBackupDiscovery is created", func() {
			Expect(CreateObj(
				infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeBackupDiscovery,
				WithName(name),
			)).To(Succeed())
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

		By("// cleanup: When GcpNfsVolumeBackupDiscovery is deleted", func() {
			Expect(Delete(infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeBackupDiscovery)).
				To(Succeed())
		})

		By("// cleanup: Then GcpNfsVolumeBackupDiscovery does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeBackupDiscovery).
				Should(Succeed())
		})
	})

})
