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
	skrgcpnfsvolbackupv1 "github.com/kyma-project/cloud-manager/pkg/skr/gcpnfsvolumebackup/v1"
	skrgcpnfsvolbackupv2 "github.com/kyma-project/cloud-manager/pkg/skr/gcpnfsvolumebackup/v2"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	"github.com/kyma-project/cloud-manager/pkg/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Feature: SKR GcpNfsVolumeRestore V2", func() {

	BeforeEach(func() {
		if !feature.GcpNfsRestoreV2.Value(context.Background()) {
			Skip("Skipping v2 GcpNfsVolumeRestore tests because gcpNfsRestoreV2 feature flag is disabled")
		}
	})

	It("Scenario: SKR GcpNfsVolumeRestore V2 is created with backup ref and completed", func() {
		gcpMock := infra.GcpMock2().NewSubscription("nfs-restore-v2")
		defer gcpMock.Delete()

		skrGcpNfsVolumeName := "3ec6e249-de2f-42fc-9c2f-5334114a1537"
		skrGcpNfsVolumeId := "a7b5c8d2-3e4f-5a6b-7c8d-9e0f1a2b3c4d"
		skrGcpNfsVolume := &cloudresourcesv1beta1.GcpNfsVolume{}
		skrIpRangeName := "da5f0c69-6e3b-4b81-a9a9-4152869f2611"
		skrGcpNfsBackupName := "3e9ae34a-b225-4dd7-8d88-ba4527d816e2"
		skrGcpNfsBackup := &cloudresourcesv1beta1.GcpNfsVolumeBackup{}
		scope := &cloudcontrolv1beta1.Scope{}
		gcpNfsVolumeRestore := &cloudresourcesv1beta1.GcpNfsVolumeRestore{}
		gcpNfsVolumeRestoreName := "9a63bc2b-055c-45c9-9128-37863cd2f00a"

		kymaName := "afc1d85e-bc28-43b8-ad36-49461d60bb77"
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

		addressName := "test-psa-address-restore"
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

		By("And Given GCP Filestore instance exists as restore destination", func() {
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
			Expect(gcpMock.ResolveFilestoreOperation(infra.Ctx(), instanceOp.Name())).To(Succeed())
		})

		By("And Given SKR GcpNfsVolumeBackup exists in Ready state", func() {
			skrgcpnfsvolbackupv1.Ignore.AddName(skrGcpNfsBackupName)
			skrgcpnfsvolbackupv2.Ignore.AddName(skrGcpNfsBackupName)
			Eventually(CreateGcpNfsVolumeBackup).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrGcpNfsBackup,
					WithName(skrGcpNfsBackupName),
					WithGcpNfsVolume(skrGcpNfsVolumeName),
					WithGcpNfsVolumeBackupValues(),
				).Should(Succeed())

			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrGcpNfsBackup,
					WithConditions(SkrReadyCondition()),
					WithGcpNfsVolumeBackupStatusLocation(scope.Spec.Region),
					WithGcpNfsVolumeBackupStatusId("test-backup-id"),
				).Should(Succeed())
		})

		By("When GcpNfsVolumeRestore is created", func() {
			Eventually(CreateGcpNfsVolumeRestore).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeRestore,
					WithName(gcpNfsVolumeRestoreName),
					WithRestoreSourceBackup(skrGcpNfsBackupName),
					WithRestoreDestinationVolume(skrGcpNfsVolumeName),
				).Should(Succeed())
		})

		By("Then GcpNfsVolumeRestore reaches InProgress state", func() {
			Eventually(func() (bool, error) {
				err := infra.SKR().Client().Get(infra.Ctx(), client.ObjectKeyFromObject(gcpNfsVolumeRestore), gcpNfsVolumeRestore)
				if err != nil {
					return false, err
				}
				return gcpNfsVolumeRestore.Status.State == cloudresourcesv1beta1.JobStateInProgress, nil
			}).Should(BeTrue(), "expected GcpNfsVolumeRestore to reach InProgress state")
		})

		By("And When the GCP restore operation completes", func() {
			opName := gcpNfsVolumeRestore.Status.OpIdentifier
			Expect(gcpMock.ResolveFilestoreRestoreOperation(infra.Ctx(), opName)).To(Succeed())
		})

		By("Then GcpNfsVolumeRestore has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeRestore,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
				).Should(Succeed(), "expected GcpNfsVolumeRestore with Ready condition")
		})

		By("And Then GcpNfsVolumeRestore has Done state", func() {
			Expect(gcpNfsVolumeRestore.Status.State).To(Equal(cloudresourcesv1beta1.JobStateDone))
		})

		// DELETE

		By("When GcpNfsVolumeRestore is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeRestore).
				Should(Succeed())
		})

		By("Then GcpNfsVolumeRestore does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeRestore).
				Should(Succeed(), "expected GcpNfsVolumeRestore to be deleted")
		})
	})

	It("Scenario: SKR GcpNfsVolumeRestore V2 is deleted when in Done state", func() {
		gcpMock := infra.GcpMock2().NewSubscription("nfs-restore-v2-del")
		defer gcpMock.Delete()

		skrGcpNfsVolumeName := "6e854f96-d730-4333-8263-a752346b4c89"
		skrGcpNfsVolumeId := "b8c6d9e3-4f5a-6b7c-8d9e-0f1a2b3c4d5e"
		skrGcpNfsVolume := &cloudresourcesv1beta1.GcpNfsVolume{}
		skrIpRangeName := "920ea8af-c458-4c55-9c6b-6112dfe0ae20"
		skrGcpNfsBackupName := "5ab6d98d-77a0-4747-a30f-ac8d716ffd08"
		skrGcpNfsBackup := &cloudresourcesv1beta1.GcpNfsVolumeBackup{}
		scope := &cloudcontrolv1beta1.Scope{}
		gcpNfsVolumeRestore := &cloudresourcesv1beta1.GcpNfsVolumeRestore{}
		gcpNfsVolumeRestoreName := "9f65425d-c7b4-4139-b916-4c7e091f28c0"

		kymaName := "d2a0281d-2520-4f5b-8137-b553d2334d81"
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

		addressName := "test-psa-address-restore-del"
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

		By("And Given GCP Filestore instance exists as restore destination", func() {
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
			Expect(gcpMock.ResolveFilestoreOperation(infra.Ctx(), instanceOp.Name())).To(Succeed())
		})

		By("And Given SKR GcpNfsVolumeBackup exists in Ready state", func() {
			skrgcpnfsvolbackupv1.Ignore.AddName(skrGcpNfsBackupName)
			skrgcpnfsvolbackupv2.Ignore.AddName(skrGcpNfsBackupName)
			Eventually(CreateGcpNfsVolumeBackup).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrGcpNfsBackup,
					WithName(skrGcpNfsBackupName),
					WithGcpNfsVolume(skrGcpNfsVolumeName),
					WithGcpNfsVolumeBackupValues(),
				).Should(Succeed())

			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrGcpNfsBackup,
					WithConditions(SkrReadyCondition()),
					WithGcpNfsVolumeBackupStatusLocation(scope.Spec.Region),
					WithGcpNfsVolumeBackupStatusId("test-backup-id-del"),
				).Should(Succeed())
		})

		By("And Given GcpNfsVolumeRestore is created", func() {
			Eventually(CreateGcpNfsVolumeRestore).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeRestore,
					WithName(gcpNfsVolumeRestoreName),
					WithRestoreSourceBackup(skrGcpNfsBackupName),
					WithRestoreDestinationVolume(skrGcpNfsVolumeName),
				).Should(Succeed())
		})

		By("And Given the GCP restore operation completes", func() {
			Eventually(func() (bool, error) {
				err := infra.SKR().Client().Get(infra.Ctx(), client.ObjectKeyFromObject(gcpNfsVolumeRestore), gcpNfsVolumeRestore)
				if err != nil {
					return false, err
				}
				return gcpNfsVolumeRestore.Status.State == cloudresourcesv1beta1.JobStateInProgress, nil
			}).Should(BeTrue(), "expected GcpNfsVolumeRestore to reach InProgress state")

			opName := gcpNfsVolumeRestore.Status.OpIdentifier
			Expect(gcpMock.ResolveFilestoreRestoreOperation(infra.Ctx(), opName)).To(Succeed())
		})

		By("And Given GcpNfsVolumeRestore reaches Done state", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeRestore,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
				).Should(Succeed(), "expected GcpNfsVolumeRestore to reach Ready/Done state")
		})

		By("When GcpNfsVolumeRestore is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeRestore).
				Should(Succeed())
		})

		By("Then GcpNfsVolumeRestore does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeRestore).
				Should(Succeed(), "expected GcpNfsVolumeRestore to be deleted")
		})
	})

	It("Scenario: SKR GcpNfsVolumeRestore V2 fails when GcpNfsVolume is not ready", func() {
		gcpMock := infra.GcpMock2().NewSubscription("nfs-restore-v2-err")
		defer gcpMock.Delete()

		skrGcpNfsVolumeName := "c3d310f8-b26b-4852-b2f1-b46294fdaae0"
		skrGcpNfsVolume := &cloudresourcesv1beta1.GcpNfsVolume{}
		skrIpRangeName := "dd5b1196-7419-4b19-a8d8-4b373c755c1d"
		skrGcpNfsBackupName := "3a314877-9924-4977-b91e-297a4851a1cc"
		skrGcpNfsBackup := &cloudresourcesv1beta1.GcpNfsVolumeBackup{}
		scope := &cloudcontrolv1beta1.Scope{}
		gcpNfsVolumeRestore := &cloudresourcesv1beta1.GcpNfsVolumeRestore{}
		gcpNfsVolumeRestoreName := "6e5efcc7-692e-4a1c-9638-9e8a879a3544"

		kymaName := "68fea4d0-0364-4513-8f69-d7e97f93f39b"
		skrKymaRef := util.Must(infra.ScopeProvider().GetScope(infra.Ctx(), types.NamespacedName{Name: kymaName}))

		By("Given KCP Scope exists", func() {
			Expect(infra.GivenScopeGcpExistsWithProject(skrKymaRef.Name, gcpMock.ProjectId())).NotTo(HaveOccurred())
			Eventually(func() (exists bool, err error) {
				err = infra.KCP().Client().Get(infra.Ctx(), infra.KCP().ObjKey(skrKymaRef.Name), scope)
				exists = err == nil
				return exists, client.IgnoreNotFound(err)
			}).Should(BeTrue(), "expected Scope to get created")
		})

		By("And Given SKR GcpNfsVolume exists but is NOT in Ready state", func() {
			skrgcpnfsvol.Ignore.AddName(skrGcpNfsVolumeName)
			Eventually(CreateGcpNfsVolume).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrGcpNfsVolume,
					WithName(skrGcpNfsVolumeName),
					WithGcpNfsVolumeIpRange(skrIpRangeName),
				).Should(Succeed())
			// NFS Volume is NOT set to Ready state
		})

		By("And Given SKR GcpNfsVolumeBackup exists in Ready state", func() {
			skrgcpnfsvolbackupv1.Ignore.AddName(skrGcpNfsBackupName)
			skrgcpnfsvolbackupv2.Ignore.AddName(skrGcpNfsBackupName)
			Eventually(CreateGcpNfsVolumeBackup).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrGcpNfsBackup,
					WithName(skrGcpNfsBackupName),
					WithGcpNfsVolume(skrGcpNfsVolumeName),
					WithGcpNfsVolumeBackupValues(),
				).Should(Succeed())

			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrGcpNfsBackup,
					WithConditions(SkrReadyCondition()),
				).Should(Succeed())
		})

		By("When GcpNfsVolumeRestore is created", func() {
			Eventually(CreateGcpNfsVolumeRestore).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeRestore,
					WithName(gcpNfsVolumeRestoreName),
					WithRestoreSourceBackup(skrGcpNfsBackupName),
					WithRestoreDestinationVolume(skrGcpNfsVolumeName),
				).Should(Succeed())
		})

		By("Then GcpNfsVolumeRestore has Error condition with NfsVolumeNotReady reason", func() {
			Eventually(func() (bool, error) {
				err := infra.SKR().Client().Get(infra.Ctx(), client.ObjectKeyFromObject(gcpNfsVolumeRestore), gcpNfsVolumeRestore)
				if err != nil {
					return false, err
				}
				errCond := meta.FindStatusCondition(gcpNfsVolumeRestore.Status.Conditions, cloudresourcesv1beta1.ConditionTypeError)
				if errCond == nil {
					return false, nil
				}
				return errCond.Reason == cloudresourcesv1beta1.ConditionReasonNfsVolumeNotReady, nil
			}).Should(BeTrue(), "expected GcpNfsVolumeRestore to have Error condition with NfsVolumeNotReady reason")
		})
	})
})
