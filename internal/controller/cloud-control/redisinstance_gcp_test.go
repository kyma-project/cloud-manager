package cloudcontrol

import (
	"fmt"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"
	"cloud.google.com/go/longrunning/autogen/longrunningpb"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	kcpiprange "github.com/kyma-project/cloud-manager/pkg/kcp/iprange"
	kcpscope "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"
)

var _ = Describe("Feature: KCP RedisInstance", func() {

	It("Scenario: KCP GCP RedisInstance is created and deleted", func() {

		name := "924a92cf-9e72-408d-a1e8-017a2fd8d42d"
		scope := &cloudcontrolv1beta1.Scope{}

		gcpMock := infra.GcpMock2().NewSubscription("redis-instance")
		defer gcpMock.Delete()

		By("Given Scope exists", func() {
			// Tell Scope reconciler to ignore this kymaName
			kcpscope.Ignore.AddName(name)

			Eventually(CreateScopeGcp).
				WithArguments(infra.Ctx(), infra, scope, WithName(name)).
				Should(Succeed())

			// Update Scope's GCP project to match mock subscription before any reconciler uses it
			scope.Spec.Scope.Gcp.Project = gcpMock.ProjectId()
			Expect(infra.KCP().Client().Update(infra.Ctx(), scope)).To(Succeed())
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

		kcpIpRangeName := "ffc7ebcc-114e-4d68-948c-241405fd01b5"
		kcpIpRange := &cloudcontrolv1beta1.IpRange{}

		// Tell IpRange reconciler to ignore this kymaName
		kcpiprange.Ignore.AddName(kcpIpRangeName)
		By("And Given KCP IPRange exists", func() {
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

		redisInstance := &cloudcontrolv1beta1.RedisInstance{}
		memorySizeGb := 5
		replicaCount := 1

		By("When RedisInstance is created", func() {
			Eventually(CreateRedisInstance).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance,
					WithName(name),
					WithRemoteRef("skr-redis-example"),
					WithIpRange(kcpIpRangeName),
					WithScope(name),
					WithRedisInstanceGcp(),
					WithKcpGcpRedisInstanceTier("STANDARD_HA"),
					WithKcpGcpRedisInstanceMemorySizeGb(int32(memorySizeGb)),
					WithKcpGcpRedisInstanceRedisVersion("REDIS_7_0"),
					WithKcpGcpRedisInstanceAuthEnabled(true),
					WithKcpGcpRedisInstanceReplicaCount(int32(replicaCount)),
					WithKcpGcpRedisInstanceConfigs(map[string]string{
						"maxmemory-policy": "allkeys-lru",
					}),
					WithKcpGcpRedisInstanceMaintenancePolicy(&cloudcontrolv1beta1.MaintenancePolicyGcp{
						DayOfWeek: &cloudcontrolv1beta1.DayOfWeekPolicyGcp{
							Day: "MONDAY",
							StartTime: cloudcontrolv1beta1.TimeOfDayGcp{
								Hours:   14,
								Minutes: 45,
							},
						},
					}),
				).
				Should(Succeed(), "failed creating RedisInstance")
		})

		By("Then GCP Redis is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance,
					NewObjActions(),
					HavingRedisInstanceStatusId()).
				Should(Succeed(), "expected RedisInstance to get status.id")
		})

		var createOpName string
		By("When GCP Redis create operation is resolved", func() {
			it := gcpMock.ListRedisInstanceOperations(infra.Ctx(), &longrunningpb.ListOperationsRequest{})
			for op, err := it.Next(); err == nil; op, err = it.Next() {
				if !op.Done && op.Name != "" {
					createOpName = op.Name
					break
				}
			}
			Expect(createOpName).ToNot(BeEmpty(), "expected to find a pending create operation")
			Expect(gcpMock.ResolveRedisInstanceOperation(infra.Ctx(), createOpName)).To(Succeed())
		})

		By("Then RedisInstance has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
					HavingState("Ready"),
				).
				Should(Succeed(), "expected RedisInstance to has Ready state, but it didn't")
		})

		By("And Then RedisInstance has .status.primaryEndpoint set", func() {
			Expect(len(redisInstance.Status.PrimaryEndpoint) > 0).To(Equal(true))
		})
		By("And Then RedisInstance has .status.readEndpoint set", func() {
			Expect(len(redisInstance.Status.ReadEndpoint) > 0).To(Equal(true))
		})
		By("And Then RedisInstance has .status.authString set", func() {
			Expect(len(redisInstance.Status.AuthString) > 0).To(Equal(true))
		})

		By("And Then RedisInstance has .status.memorySizeGb set", func() {
			Expect(redisInstance.Status.MemorySizeGb).To(Equal(int32(memorySizeGb)))
		})

		By("And Then RedisInstance has .status.replicaCount set", func() {
			Expect(redisInstance.Status.ReplicaCount).To(Equal(int32(replicaCount)))
		})

		// DELETE

		By("When RedisInstance is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance).
				Should(Succeed(), "failed deleting RedisInstance")
		})

		By("And When GCP Redis delete operation is resolved", func() {
			Eventually(func() error {
				it := gcpMock.ListRedisInstanceOperations(infra.Ctx(), &longrunningpb.ListOperationsRequest{})
				for op, err := it.Next(); err == nil; op, err = it.Next() {
					if !op.Done && op.Name != "" && op.Name != createOpName {
						return gcpMock.ResolveRedisInstanceOperation(infra.Ctx(), op.Name)
					}
				}
				return fmt.Errorf("no pending delete operation found yet")
			}).Should(Succeed(), "expected to find and resolve delete operation")
		})

		By("Then RedisInstance does not exist", func() {
			Eventually(IsDeleted, 5*time.Second).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance).
				Should(Succeed(), "expected RedisInstance not to exist (be deleted), but it still exists")
		})
	})

})
