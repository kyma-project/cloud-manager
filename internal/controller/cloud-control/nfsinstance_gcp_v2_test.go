package cloudcontrol

import (
	"fmt"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"
	"cloud.google.com/go/filestore/apiv1/filestorepb"
	"cloud.google.com/go/longrunning/autogen/longrunningpb"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	kcpiprange "github.com/kyma-project/cloud-manager/pkg/kcp/iprange"
	gcpnfsinstancev2client "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsinstance/v2/client"
	kcpscope "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"
)

var _ = Describe("Feature: KCP NfsInstance GCP v2", func() {

	// Skip v2 tests when GcpNfsInstanceV2 flag is not enabled
	BeforeEach(func() {
		if !feature.GcpNfsInstanceV2.Value(infra.Ctx()) {
			Skip("Skipping v2 tests when GcpNfsInstanceV2 flag is not enabled")
		}
	})

	It("Scenario: KCP GCP NfsInstance v2 is created, updated and deleted", func() {

		const kymaName = "38bd0764-ba1b-4428-9afa-7736aee31ded"
		scope := &cloudcontrolv1beta1.Scope{}

		gcpMock := infra.GcpMock2().NewSubscription("nfs-instance-v2")
		defer gcpMock.Delete()

		By("Given Scope exists", func() {
			// Tell Scope reconciler to ignore this kymaName
			kcpscope.Ignore.AddName(kymaName)

			Eventually(CreateScopeGcp2).
				WithArguments(infra.Ctx(), infra, scope, gcpMock.ProjectId(), WithName(kymaName)).
				Should(Succeed())
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

		kcpIpRangeName := "ffff14c2-0937-43cb-872f-cc5573e7c5b9"
		kcpIpRange := &cloudcontrolv1beta1.IpRange{}

		// Tell IpRange reconciler to ignore this kymaName
		kcpiprange.Ignore.AddName(kcpIpRangeName)
		By("And Given KCP IpRange exists", func() {
			Eventually(CreateKcpIpRange).
				WithArguments(
					infra.Ctx(), infra.KCP().Client(), kcpIpRange,
					WithName(kcpIpRangeName),
					WithScope(scope.Name),
				).
				Should(Succeed())
		})

		By("And Given KCP IpRange has Ready condition and Status.Id", func() {
			addr, err := gcpMock.GetGlobalAddress(infra.Ctx(), &computepb.GetGlobalAddressRequest{
				Project: gcpMock.ProjectId(),
				Address: addressName,
			})
			Expect(err).ToNot(HaveOccurred())
			kcpIpRange.Status.Id = addr.GetSelfLink()
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.KCP().Client(), kcpIpRange,
					WithKcpIpRangeStatusCidr(kcpIpRange.Spec.Cidr),
					WithConditions(KcpReadyCondition()),
				).
				Should(Succeed(), "Expected KCP IpRange to become ready")
		})

		nfsInstance := &cloudcontrolv1beta1.NfsInstance{}

		By("When NfsInstance is created", func() {
			Eventually(CreateNfsInstance).
				WithArguments(
					infra.Ctx(), infra.KCP().Client(), nfsInstance,
					WithName(kymaName),
					WithRemoteRef("skr-nfs-v2-example"),
					WithIpRange(kcpIpRangeName),
					WithScope(kymaName),
					WithNfsInstanceGcp(scope.Spec.Region),
				).
				Should(Succeed(), "failed creating NfsInstance")
		})

		By("Then GCP Filestore instance is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), nfsInstance,
					NewObjActions(),
					HavingNfsInstanceStatusId()).
				Should(Succeed(), "expected NfsInstance to get status.id")

			// Get instance from mock2 using the correct GCP instance name
			instanceName := gcpnfsinstancev2client.GetFilestoreName(gcpMock.ProjectId(), scope.Spec.Region, kymaName)
			filestoreInstance, err := gcpMock.GetFilestoreInstance(infra.Ctx(), &filestorepb.GetInstanceRequest{
				Name: instanceName,
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(filestoreInstance).NotTo(BeNil())
		})

		var createOpName string
		By("When GCP Filestore create operation is resolved", func() {
			Eventually(func() error {
				it := gcpMock.ListFilestoreOperations(infra.Ctx(), &longrunningpb.ListOperationsRequest{})
				for op, err := it.Next(); err == nil; op, err = it.Next() {
					if !op.Done && op.Name != "" {
						createOpName = op.Name
						return gcpMock.ResolveFilestoreOperation(infra.Ctx(), createOpName)
					}
				}
				return fmt.Errorf("no pending create operation found yet")
			}).Should(Succeed(), "expected to find and resolve create operation")
		})

		By("Then NfsInstance has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), nfsInstance,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
					HavingState("Ready"),
				).
				Should(Succeed(), "expected NfsInstance to have Ready state, but it didn't")
		})

		By("And Then NfsInstance has .status.host set", func() {
			Expect(len(nfsInstance.Status.Host) > 0).To(Equal(true))
		})

		By("And Then NfsInstance has .status.path set", func() {
			Expect(len(nfsInstance.Status.Path) > 0).To(Equal(true))
		})

		By("And Then NfsInstance has both capacity fields set consistently", func() {
			Expect(nfsInstance.Status.Capacity.IsZero()).To(BeFalse())
			Expect(nfsInstance.Status.CapacityGb).To(BeNumerically(">", 0))
			// Verify both fields represent the same value
			expectedQuantity := resource.MustParse(fmt.Sprintf("%dGi", nfsInstance.Status.CapacityGb))
			Expect(nfsInstance.Status.Capacity.Cmp(expectedQuantity)).To(Equal(0))
		})

		originalCapacity := nfsInstance.Status.CapacityGb

		// UPDATE

		By("When NfsInstance capacity is increased", func() {
			Eventually(UpdateNfsInstance).
				WithArguments(
					infra.Ctx(), infra.KCP().Client(), nfsInstance,
				).
				Should(Succeed())
		})

		By("Then NfsInstance still has Ready condition during update", func() {
			Expect(nfsInstance.Status.State).To(Equal(cloudcontrolv1beta1.StateReady))
		})

		By("And When GCP Filestore update operation is resolved", func() {
			Eventually(func() error {
				it := gcpMock.ListFilestoreOperations(infra.Ctx(), &longrunningpb.ListOperationsRequest{})
				for op, err := it.Next(); err == nil; op, err = it.Next() {
					if !op.Done && op.Name != "" && op.Name != createOpName {
						return gcpMock.ResolveFilestoreOperation(infra.Ctx(), op.Name)
					}
				}
				return fmt.Errorf("no pending update operation found yet")
			}).Should(Succeed(), "expected to find and resolve update operation")
		})

		By("And Then NfsInstance status reflects the updated capacity", func() {
			expectedCapacityGb := 2 * originalCapacity
			expectedCapacity := resource.MustParse(fmt.Sprintf("%dGi", expectedCapacityGb))
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), nfsInstance,
					NewObjActions(),
					HavingNfsInstanceStatusCapacityGb(expectedCapacityGb),
					HavingNfsInstanceStatusCapacity(expectedCapacity)).
				Should(Succeed())
		})

		// DELETE

		By("When NfsInstance is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), nfsInstance).
				Should(Succeed(), "failed deleting NfsInstance")
		})

		By("And When GCP Filestore delete operation is resolved", func() {
			Eventually(func() error {
				it := gcpMock.ListFilestoreOperations(infra.Ctx(), &longrunningpb.ListOperationsRequest{})
				for op, err := it.Next(); err == nil; op, err = it.Next() {
					if !op.Done && op.Name != "" {
						return gcpMock.ResolveFilestoreOperation(infra.Ctx(), op.Name)
					}
				}
				return fmt.Errorf("no pending delete operation found yet")
			}).Should(Succeed(), "expected to find and resolve delete operation")
		})

		By("Then NfsInstance does not exist", func() {
			Eventually(IsDeleted, 5*time.Second).
				WithArguments(infra.Ctx(), infra.KCP().Client(), nfsInstance).
				Should(Succeed(), "expected NfsInstance not to exist (be deleted), but it still exists")
		})
	})

	It("Scenario: KCP GCP NfsInstance v2 is created from backup", func() {

		const kymaName = "76261c61-5ff3-45e7-9642-47b574003ee7"
		scope := &cloudcontrolv1beta1.Scope{}

		gcpMock := infra.GcpMock2().NewSubscription("nfs-instance-v2-backup")
		defer gcpMock.Delete()

		By("Given Scope exists", func() {
			// Tell Scope reconciler to ignore this kymaName
			kcpscope.Ignore.AddName(kymaName)

			Eventually(CreateScopeGcp2).
				WithArguments(infra.Ctx(), infra, scope, gcpMock.ProjectId(), WithName(kymaName)).
				Should(Succeed())
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

		kcpIpRangeName := "14161cce-cd3e-49f8-9b06-18db08409440"
		kcpIpRange := &cloudcontrolv1beta1.IpRange{}

		// Tell IpRange reconciler to ignore this kymaName
		kcpiprange.Ignore.AddName(kcpIpRangeName)
		By("And Given KCP IpRange exists", func() {
			Eventually(CreateKcpIpRange).
				WithArguments(
					infra.Ctx(), infra.KCP().Client(), kcpIpRange,
					WithName(kcpIpRangeName),
					WithScope(scope.Name),
				).
				Should(Succeed())
		})

		By("And Given KCP IpRange has Ready condition and Status.Id", func() {
			addr, err := gcpMock.GetGlobalAddress(infra.Ctx(), &computepb.GetGlobalAddressRequest{
				Project: gcpMock.ProjectId(),
				Address: addressName,
			})
			Expect(err).ToNot(HaveOccurred())
			kcpIpRange.Status.Id = addr.GetSelfLink()
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.KCP().Client(), kcpIpRange,
					WithKcpIpRangeStatusCidr(kcpIpRange.Spec.Cidr),
					WithConditions(KcpReadyCondition()),
				).
				Should(Succeed(), "Expected KCP IpRange to become ready")
		})

		nfsInstance := &cloudcontrolv1beta1.NfsInstance{}
		sourceBackupPath := "projects/test-project/locations/us-central1/backups/my-backup"

		By("When NfsInstance is created with SourceBackup", func() {
			Eventually(CreateNfsInstance).
				WithArguments(
					infra.Ctx(), infra.KCP().Client(), nfsInstance,
					WithName(kymaName),
					WithRemoteRef("skr-nfs-v2-from-backup"),
					WithIpRange(kcpIpRangeName),
					WithScope(kymaName),
					WithNfsInstanceGcp(scope.Spec.Region),
					WithSourceBackup(sourceBackupPath),
				).
				Should(Succeed(), "failed creating NfsInstance with SourceBackup")
		})

		By("Then GCP Filestore instance is created with SourceBackup", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), nfsInstance,
					NewObjActions(),
					HavingNfsInstanceStatusId()).
				Should(Succeed(), "expected NfsInstance to get status.id")

			// Get instance from mock2 using the correct GCP instance name
			instanceName := gcpnfsinstancev2client.GetFilestoreName(gcpMock.ProjectId(), scope.Spec.Region, kymaName)
			filestoreInstance, err := gcpMock.GetFilestoreInstance(infra.Ctx(), &filestorepb.GetInstanceRequest{
				Name: instanceName,
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(filestoreInstance).NotTo(BeNil())
			Expect(filestoreInstance.FileShares).To(HaveLen(1))
			Expect(filestoreInstance.FileShares[0].GetSourceBackup()).To(Equal(sourceBackupPath))
		})

		var createOpName string
		By("When GCP Filestore create operation is resolved", func() {
			Eventually(func() error {
				it := gcpMock.ListFilestoreOperations(infra.Ctx(), &longrunningpb.ListOperationsRequest{})
				for op, err := it.Next(); err == nil; op, err = it.Next() {
					if !op.Done && op.Name != "" {
						createOpName = op.Name
						return gcpMock.ResolveFilestoreOperation(infra.Ctx(), createOpName)
					}
				}
				return fmt.Errorf("no pending create operation found yet")
			}).Should(Succeed(), "expected to find and resolve create operation")
		})

		By("Then NfsInstance has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), nfsInstance,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
					HavingState("Ready"),
				).
				Should(Succeed(), "expected NfsInstance to have Ready state, but it didn't")
		})

		// DELETE

		By("When NfsInstance is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), nfsInstance).
				Should(Succeed(), "failed deleting NfsInstance")
		})

		By("And When GCP Filestore delete operation is resolved", func() {
			Eventually(func() error {
				it := gcpMock.ListFilestoreOperations(infra.Ctx(), &longrunningpb.ListOperationsRequest{})
				for op, err := it.Next(); err == nil; op, err = it.Next() {
					if !op.Done && op.Name != "" && op.Name != createOpName {
						return gcpMock.ResolveFilestoreOperation(infra.Ctx(), op.Name)
					}
				}
				return fmt.Errorf("no pending delete operation found yet")
			}).Should(Succeed(), "expected to find and resolve delete operation")
		})

		By("Then NfsInstance does not exist", func() {
			Eventually(IsDeleted, 5*time.Second).
				WithArguments(infra.Ctx(), infra.KCP().Client(), nfsInstance).
				Should(Succeed(), "expected NfsInstance not to exist (be deleted), but it still exists")
		})
	})

})
