package cloudresources

import (
	"context"
	"fmt"

	"cloud.google.com/go/compute/apiv1/computepb"
	"cloud.google.com/go/filestore/apiv1/filestorepb"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	skrgcpnfsvol "github.com/kyma-project/cloud-manager/pkg/skr/gcpnfsvolume"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	"github.com/kyma-project/cloud-manager/pkg/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Feature: SKR GcpNfsVolumeBackup V2", func() {

	BeforeEach(func() {
		if !feature.GcpBackupV2.Value(context.Background()) {
			Skip("Skipping v2 GcpNfsVolumeBackup tests because gcpBackupV2 feature flag is disabled!")
		}
	})

	It("Scenario: SKR GcpNfsVolumeBackup V2 is created and deleted", func() {
		gcpMock := infra.GcpMock2().NewSubscription("nfs-backup-v2")
		defer gcpMock.Delete()

		skrGcpNfsVolumeName := "207c83d0-2480-4a2d-b402-938a66d2ca9e"
		skrGcpNfsVolumeId := "b7e6f9b1-11de-40aa-8c24-1207791fc0b9"
		skrGcpNfsVolume := &cloudresourcesv1beta1.GcpNfsVolume{}
		skrIpRangeName := "3d5eef8e-b871-4147-b6a2-7a49753c8bf8"
		scope := &cloudcontrolv1beta1.Scope{}
		gcpNfsVolumeBackup := &cloudresourcesv1beta1.GcpNfsVolumeBackup{}
		gcpNfsVolumeBackupName := "8469ce3a-6529-474d-8e83-d5b8aef13362"

		kymaName := "1ab2af56-ceb5-4a93-8ad8-c846617101f2"
		skrKymaRef := util.Must(infra.ScopeProvider().GetScope(infra.Ctx(), types.NamespacedName{Name: kymaName}))

		By("Given KCP Scope exists", func() {
			Expect(infra.GivenScopeGcpExistsWithProject(skrKymaRef.Name, gcpMock.ProjectId())).NotTo(HaveOccurred())
			Eventually(func() (exists bool, err error) {
				err = infra.KCP().Client().Get(infra.Ctx(), infra.KCP().ObjKey(skrKymaRef.Name), scope)
				exists = err == nil
				return exists, client.IgnoreNotFound(err)
			}).Should(BeTrue(), "expected Scope to get created")
		})

		vpcNetworkName := scope.Spec.Scope.Gcp.VpcNetwork

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

		addressName := "test-psa-address"
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

		By("And Given SKR GcpNfsVolume exists in Ready state", func() {
			skrgcpnfsvol.Ignore.AddName(skrGcpNfsVolumeName)
			Eventually(CreateGcpNfsVolume).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrGcpNfsVolume,
					WithName(skrGcpNfsVolumeName),
					WithGcpNfsVolumeIpRange(skrIpRangeName),
				).Should(Succeed())

			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrGcpNfsVolume,
					WithConditions(SkrReadyCondition()),
					WithGcpNfsVolumeStatusLocation(scope.Spec.Region),
					WithGcpNfsVolumeStatusId(skrGcpNfsVolumeId),
				).Should(Succeed())
		})

		By("And Given GCP Filestore instance exists", func() {
			// The reconciler constructs instance name as "cm-<nfsVolumeId>"
			nfsInstanceName := fmt.Sprintf("cm-%.60s", skrGcpNfsVolumeId)
			addr, err := gcpMock.GetGlobalAddress(infra.Ctx(), &computepb.GetGlobalAddressRequest{
				Project: gcpMock.ProjectId(),
				Address: addressName,
			})
			Expect(err).ToNot(HaveOccurred())

			instanceOp, err := gcpMock.CreateFilestoreInstance(infra.Ctx(), &filestorepb.CreateInstanceRequest{
				Parent:     fmt.Sprintf("projects/%s/locations/%s", gcpMock.ProjectId(), scope.Spec.Region),
				InstanceId: nfsInstanceName,
				Instance: &filestorepb.Instance{
					Tier: filestorepb.Instance_BASIC_HDD,
					FileShares: []*filestorepb.FileShareConfig{
						{
							Name:       skrGcpNfsVolume.Spec.FileShareName,
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

		By("When GcpNfsVolumeBackup is created", func() {
			Eventually(CreateGcpNfsVolumeBackup).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeBackup,
					WithName(gcpNfsVolumeBackupName),
					WithGcpNfsVolume(skrGcpNfsVolumeName),
				).Should(Succeed())
		})

		By("Then GcpNfsVolumeBackup has status.id set", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeBackup,
					NewObjActions(),
					HavingFieldSet("status", "id"),
				).Should(Succeed(), "expected GcpNfsVolumeBackup to get status.id")
		})

		By("When GCP Backup create operation is resolved", func() {
			// Wait for the backup to have an operation identifier in status
			Eventually(func() string {
				_ = infra.SKR().Client().Get(infra.Ctx(), client.ObjectKeyFromObject(gcpNfsVolumeBackup), gcpNfsVolumeBackup)
				return gcpNfsVolumeBackup.Status.OpIdentifier
			}).ShouldNot(BeEmpty(), "expected backup to have OpIdentifier")

			opName := gcpNfsVolumeBackup.Status.OpIdentifier
			Expect(gcpMock.ResolveFilestoreBackupOperation(infra.Ctx(), opName)).To(Succeed())
		})

		By("Then GcpNfsVolumeBackup has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeBackup,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
					AssertGcpNfsVolumeBackupHasState(cloudresourcesv1beta1.StateReady),
				).Should(Succeed())
		})

		By("And Then GcpNfsVolumeBackup has .status.location set", func() {
			Expect(gcpNfsVolumeBackup.Status.Location).To(Equal(gcpNfsVolumeBackup.Spec.Location))
			Expect(len(gcpNfsVolumeBackup.Status.Location)).To(BeNumerically(">", 0))
		})

		// DELETE

		By("When GcpNfsVolumeBackup is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeBackup).
				Should(Succeed())
		})

		By("Then GCP Backup is in Deleting state", func() {
			gcpBackupPath := GcpNfsVolumeBackupPath(scope, gcpNfsVolumeBackup)
			Eventually(func() filestorepb.Backup_State {
				backup, err := gcpMock.GetFilestoreBackup(infra.Ctx(), &filestorepb.GetBackupRequest{Name: gcpBackupPath})
				if err != nil {
					return filestorepb.Backup_STATE_UNSPECIFIED
				}
				return backup.State
			}).Should(Equal(filestorepb.Backup_DELETING))
		})

		By("When GCP Backup delete operation is resolved", func() {
			// Wait for delete operation identifier.
			Eventually(func() string {
				_ = infra.SKR().Client().Get(infra.Ctx(), client.ObjectKeyFromObject(gcpNfsVolumeBackup), gcpNfsVolumeBackup)
				return gcpNfsVolumeBackup.Status.OpIdentifier
			}).ShouldNot(BeEmpty(), "expected backup to have delete OpIdentifier")

			opName := gcpNfsVolumeBackup.Status.OpIdentifier
			Expect(gcpMock.ResolveFilestoreBackupOperation(infra.Ctx(), opName)).To(Succeed())
		})

		By("Then GcpNfsVolumeBackup does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeBackup).
				Should(Succeed(), "expected GcpNfsVolumeBackup to be deleted")
		})
	})

	It("Scenario: SKR GcpNfsVolumeBackup V2 is created with empty location", func() {
		gcpMock := infra.GcpMock2().NewSubscription("nfs-backup-v2-empty-loc")
		defer gcpMock.Delete()

		skrGcpNfsVolumeName := "b5506f48-e7d5-4ef8-ab12-e0b03d933b3b"
		skrGcpNfsVolumeId := "8ad4d70a-b808-4c3a-a6d2-def64ddb31ad"
		skrGcpNfsVolume := &cloudresourcesv1beta1.GcpNfsVolume{}
		skrIpRangeName := "e2cf51c1-b381-45ca-8ea4-9031d941c0ac"
		scope := &cloudcontrolv1beta1.Scope{}
		gcpNfsVolumeBackup := &cloudresourcesv1beta1.GcpNfsVolumeBackup{}
		gcpNfsVolumeBackupName := "34b17074-7cee-4654-bb99-bedfef29aaa7"

		kymaName := "c9459714-1c4c-43ea-ad0c-6c4ef1017772"
		skrKymaRef := util.Must(infra.ScopeProvider().GetScope(infra.Ctx(), types.NamespacedName{Name: kymaName}))

		By("Given KCP Scope exists", func() {
			Expect(infra.GivenScopeGcpExistsWithProject(skrKymaRef.Name, gcpMock.ProjectId())).NotTo(HaveOccurred())
			Eventually(func() (exists bool, err error) {
				err = infra.KCP().Client().Get(infra.Ctx(), infra.KCP().ObjKey(skrKymaRef.Name), scope)
				exists = err == nil
				return exists, client.IgnoreNotFound(err)
			}).Should(BeTrue(), "expected Scope to get created")
		})

		vpcNetworkName := scope.Spec.Scope.Gcp.VpcNetwork

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

		addressName := "test-psa-address-empty-loc"
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
					Address:      ptr.To("10.252.0.0"),
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

		By("And Given SKR GcpNfsVolume exists in Ready state", func() {
			skrgcpnfsvol.Ignore.AddName(skrGcpNfsVolumeName)
			Eventually(CreateGcpNfsVolume).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrGcpNfsVolume,
					WithName(skrGcpNfsVolumeName),
					WithGcpNfsVolumeIpRange(skrIpRangeName),
				).Should(Succeed())

			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrGcpNfsVolume,
					WithConditions(SkrReadyCondition()),
					WithGcpNfsVolumeStatusLocation(scope.Spec.Region),
					WithGcpNfsVolumeStatusId(skrGcpNfsVolumeId),
				).Should(Succeed())
		})

		By("And Given GCP Filestore instance exists", func() {
			nfsInstanceName := fmt.Sprintf("cm-%.60s", skrGcpNfsVolumeId)
			addr, err := gcpMock.GetGlobalAddress(infra.Ctx(), &computepb.GetGlobalAddressRequest{
				Project: gcpMock.ProjectId(),
				Address: addressName,
			})
			Expect(err).ToNot(HaveOccurred())

			instanceOp, err := gcpMock.CreateFilestoreInstance(infra.Ctx(), &filestorepb.CreateInstanceRequest{
				Parent:     fmt.Sprintf("projects/%s/locations/%s", gcpMock.ProjectId(), scope.Spec.Region),
				InstanceId: nfsInstanceName,
				Instance: &filestorepb.Instance{
					Tier: filestorepb.Instance_BASIC_HDD,
					FileShares: []*filestorepb.FileShareConfig{
						{
							Name:       skrGcpNfsVolume.Spec.FileShareName,
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

		By("When GcpNfsVolumeBackup is created with empty location", func() {
			Eventually(CreateGcpNfsVolumeBackup).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeBackup,
					WithName(gcpNfsVolumeBackupName),
					WithGcpNfsVolume(skrGcpNfsVolumeName),
					WithGcpNfsVolumeBackupLocation(""),
				).Should(Succeed())
		})

		By("Then GcpNfsVolumeBackup has status.id set", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeBackup,
					NewObjActions(),
					HavingGcpNfsVolumeBackupStatusId(),
				).Should(Succeed(), "expected GcpNfsVolumeBackup to get status.id")
		})

		By("When GCP Backup create operation is resolved", func() {
			Eventually(func() string {
				_ = infra.SKR().Client().Get(infra.Ctx(), client.ObjectKeyFromObject(gcpNfsVolumeBackup), gcpNfsVolumeBackup)
				return gcpNfsVolumeBackup.Status.OpIdentifier
			}).ShouldNot(BeEmpty(), "expected backup to have OpIdentifier")

			opName := gcpNfsVolumeBackup.Status.OpIdentifier
			Expect(gcpMock.ResolveFilestoreBackupOperation(infra.Ctx(), opName)).To(Succeed())
		})

		By("Then GcpNfsVolumeBackup has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeBackup,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
					AssertGcpNfsVolumeBackupHasState(cloudresourcesv1beta1.StateReady),
				).Should(Succeed())
		})

		By("And Then GcpNfsVolumeBackup has .status.location set from Scope", func() {
			Expect(gcpNfsVolumeBackup.Status.Location).To(Equal(scope.Spec.Region))
		})

		// DELETE (cleanup)

		By("When GcpNfsVolumeBackup is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeBackup).
				Should(Succeed())
		})

		By("When GCP Backup delete operation is resolved", func() {
			Eventually(func() string {
				_ = infra.SKR().Client().Get(infra.Ctx(), client.ObjectKeyFromObject(gcpNfsVolumeBackup), gcpNfsVolumeBackup)
				return gcpNfsVolumeBackup.Status.OpIdentifier
			}).ShouldNot(BeEmpty(), "expected backup to have delete OpIdentifier")

			opName := gcpNfsVolumeBackup.Status.OpIdentifier
			Expect(gcpMock.ResolveFilestoreBackupOperation(infra.Ctx(), opName)).To(Succeed())
		})

		By("Then GcpNfsVolumeBackup does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeBackup).
				Should(Succeed(), "expected GcpNfsVolumeBackup to be deleted")
		})
	})

})
