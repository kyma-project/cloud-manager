package cloudresources

import (
	"fmt"

	"cloud.google.com/go/compute/apiv1/computepb"
	"cloud.google.com/go/filestore/apiv1/filestorepb"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	"github.com/kyma-project/cloud-manager/pkg/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Feature: SKR GcpNfsVolumeBackupDiscovery", func() {

	It("Scenario: SKR GcpNfsVolumeBackupDiscovery is created", func() {

		gcpMock := infra.GcpMock2().NewSubscription("backup-discovery")
		defer gcpMock.Delete()

		scopeName := infra.SkrKymaRef().Name
		shootName := scopeName // ShootName is always set to scope.Name in givenScopeGcpExistsWithProject
		scope := &cloudcontrolv1beta1.Scope{}

		By("Given KCP Scope exists", func() {
			Expect(infra.GivenScopeGcpExistsWithProject(scopeName, gcpMock.ProjectId())).NotTo(HaveOccurred())
			Eventually(func() (exists bool, err error) {
				err = infra.KCP().Client().Get(infra.Ctx(), infra.KCP().ObjKey(scopeName), scope)
				exists = err == nil
				return exists, client.IgnoreNotFound(err)
			}).Should(BeTrue(), "expected Scope to get created")
		})

		vpcNetworkName := scope.Spec.Scope.Gcp.VpcNetwork
		addressName := "discovery-psa-address"

		By("And Given GCP VPC network exists", func() {
			op, err := gcpMock.InsertNetwork(infra.Ctx(), &computepb.InsertNetworkRequest{
				Project: gcpMock.ProjectId(),
				NetworkResource: &computepb.Network{
					Name: ptr.To(vpcNetworkName),
				},
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(op.Wait(infra.Ctx())).To(Succeed())
		})

		By("And Given GCP PSA address range exists", func() {
			net, err := gcpMock.GetNetwork(infra.Ctx(), &computepb.GetNetworkRequest{
				Project: gcpMock.ProjectId(),
				Network: vpcNetworkName,
			})
			Expect(err).ToNot(HaveOccurred())
			op, err := gcpMock.InsertGlobalAddress(infra.Ctx(), &computepb.InsertGlobalAddressRequest{
				Project: gcpMock.ProjectId(),
				AddressResource: &computepb.Address{
					Name:         ptr.To(addressName),
					Address:      ptr.To("10.251.0.0"),
					PrefixLength: ptr.To(int32(16)),
					Network:      ptr.To(net.GetSelfLink()),
					AddressType:  ptr.To(computepb.Address_INTERNAL.String()),
					Purpose:      ptr.To(computepb.Address_VPC_PEERING.String()),
				},
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(op.Wait(infra.Ctx())).To(Succeed())
		})

		By("And Given GCP PSA connection exists", func() {
			addr, err := gcpMock.GetGlobalAddress(infra.Ctx(), &computepb.GetGlobalAddressRequest{
				Project: gcpMock.ProjectId(),
				Address: addressName,
			})
			Expect(err).ToNot(HaveOccurred())
			net, err := gcpMock.GetNetwork(infra.Ctx(), &computepb.GetNetworkRequest{
				Project: gcpMock.ProjectId(),
				Network: vpcNetworkName,
			})
			Expect(err).ToNot(HaveOccurred())
			_, err = gcpMock.CreateServiceConnection(infra.Ctx(), gcpMock.ProjectId(), net.GetName(), []string{addr.GetName()})
			Expect(err).ToNot(HaveOccurred())
		})

		instanceName := "shared-instance-1"
		fileShareName := "vol1"

		By("And Given GCP Filestore instance exists", func() {
			addr, err := gcpMock.GetGlobalAddress(infra.Ctx(), &computepb.GetGlobalAddressRequest{
				Project: gcpMock.ProjectId(),
				Address: addressName,
			})
			Expect(err).ToNot(HaveOccurred())

			instanceOp, err := gcpMock.CreateFilestoreInstance(infra.Ctx(), &filestorepb.CreateInstanceRequest{
				Parent:     fmt.Sprintf("projects/%s/locations/%s", gcpMock.ProjectId(), scope.Spec.Region),
				InstanceId: instanceName,
				Instance: &filestorepb.Instance{
					Tier: filestorepb.Instance_BASIC_HDD,
					FileShares: []*filestorepb.FileShareConfig{
						{
							Name:       fileShareName,
							CapacityGb: 1024,
						},
					},
					Networks: []*filestorepb.NetworkConfig{
						{
							Network:         fmt.Sprintf("projects/%s/global/networks/%s", gcpMock.ProjectId(), vpcNetworkName),
							ConnectMode:     filestorepb.NetworkConfig_PRIVATE_SERVICE_ACCESS,
							ReservedIpRange: addr.GetSelfLink(),
						},
					},
				},
			})
			Expect(err).ToNot(HaveOccurred())
			opName := instanceOp.Name()
			Expect(gcpMock.ResolveFilestoreOperation(infra.Ctx(), opName)).To(Succeed())
		})

		sourceInstancePath := fmt.Sprintf("projects/%s/locations/%s/instances/%s",
			gcpMock.ProjectId(), scope.Spec.Region, instanceName)

		backupName1 := "shared-backup-1"
		backupName2 := "shared-backup-2"

		By("And Given shared backups exist for shoot", func() {
			// Create backup 1
			op1, err := gcpMock.CreateFilestoreBackup(infra.Ctx(), &filestorepb.CreateBackupRequest{
				Parent:   fmt.Sprintf("projects/%s/locations/%s", gcpMock.ProjectId(), scope.Spec.Region),
				BackupId: backupName1,
				Backup: &filestorepb.Backup{
					SourceInstance:  sourceInstancePath,
					SourceFileShare: fileShareName,
					Labels: map[string]string{
						"managed-by":                          "cloud-manager",
						"scope-name":                          scopeName,
						util.GcpLabelSkrVolumeName:            "volume-1",
						util.GcpLabelSkrVolumeNamespace:       "default",
						util.GcpLabelSkrBackupName:            backupName1,
						util.GcpLabelSkrBackupNamespace:       "default",
						util.GcpLabelShootName:                shootName,
						fmt.Sprintf("cm-allow-%s", shootName): util.GcpLabelBackupAccessibleFrom,
					},
				},
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(gcpMock.ResolveFilestoreBackupOperation(infra.Ctx(), op1.Name())).To(Succeed())

			// Create backup 2
			op2, err := gcpMock.CreateFilestoreBackup(infra.Ctx(), &filestorepb.CreateBackupRequest{
				Parent:   fmt.Sprintf("projects/%s/locations/%s", gcpMock.ProjectId(), scope.Spec.Region),
				BackupId: backupName2,
				Backup: &filestorepb.Backup{
					SourceInstance:  sourceInstancePath,
					SourceFileShare: fileShareName,
					Labels: map[string]string{
						"managed-by":                          "cloud-manager",
						"scope-name":                          scopeName,
						util.GcpLabelSkrVolumeName:            "volume-2",
						util.GcpLabelSkrVolumeNamespace:       "default",
						util.GcpLabelSkrBackupName:            backupName2,
						util.GcpLabelSkrBackupNamespace:       "default",
						util.GcpLabelShootName:                shootName,
						fmt.Sprintf("cm-allow-%s", shootName): util.GcpLabelBackupAccessibleFrom,
					},
				},
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(gcpMock.ResolveFilestoreBackupOperation(infra.Ctx(), op2.Name())).To(Succeed())
		})

		gcpNfsVolumeBackupDiscovery := &cloudresourcesv1beta1.GcpNfsVolumeBackupDiscovery{}
		discoveryName := "discovery-" + scopeName[:8]

		By("When GcpNfsVolumeBackupDiscovery is created", func() {
			Expect(CreateObj(
				infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeBackupDiscovery,
				WithName(discoveryName),
			)).To(Succeed())
		})

		By("Then GcpNfsVolumeBackupDiscovery will get Ready condition", func() {
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
