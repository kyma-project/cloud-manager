package cloudcontrol

import (
	"fmt"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"
	"cloud.google.com/go/longrunning/autogen/longrunningpb"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	kcpscope "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"
)

var _ = Describe("Feature: KCP GcpRedisCluster", func() {

	It("Scenario: KCP GCP GcpRedisCluster is created and deleted", func() {

		name := "10a1ff0e-cb76-4eb2-ae70-2951bb6bc439"
		scope := &cloudcontrolv1beta1.Scope{}

		gcpMock := infra.GcpMock2().NewSubscription("redis-cluster")
		defer gcpMock.Delete()

		By("Given Scope exists", func() {
			// Tell Scope reconciler to ignore this kymaName
			kcpscope.Ignore.AddName(name)

			Eventually(CreateScopeGcp2).
				WithArguments(infra.Ctx(), infra, scope, gcpMock.ProjectId(), WithName(name)).
				Should(Succeed())
		})

		By("And Given GCP VPC network exists", func() {
			op, err := gcpMock.InsertNetwork(infra.Ctx(), &computepb.InsertNetworkRequest{
				Project: gcpMock.ProjectId(),
				NetworkResource: &computepb.Network{
					Name: ptr.To(scope.Spec.Scope.Gcp.VpcNetwork),
				},
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(op.Wait(infra.Ctx())).To(Succeed())
		})

		var kcpNetworkKyma *cloudcontrolv1beta1.Network

		By("And Given KCP Kyma Network exists in Ready state", func() {
			kcpNetworkKyma = cloudcontrolv1beta1.NewNetworkBuilder().
				WithScope(name).
				WithName(common.KcpNetworkKymaCommonName(name)).
				WithGcpRef(scope.Spec.Scope.Gcp.Project, scope.Spec.Scope.Gcp.VpcNetwork).
				WithType(cloudcontrolv1beta1.NetworkTypeKyma).
				Build()

			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpNetworkKyma).
				Should(Succeed())

			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpNetworkKyma, NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady)).
				Should(Succeed())
		})

		kcpGcpSubnetName := "2599ac8d-b435-491e-8e17-30ede7b0b571"
		kcpGcpSubnet := &cloudcontrolv1beta1.GcpSubnet{}

		By("And Given KCP GcpSubnet is created and Ready", func() {
			Eventually(CreateKcpGcpSubnet).
				WithArguments(
					infra.Ctx(), infra.KCP().Client(), kcpGcpSubnet,
					WithName(kcpGcpSubnetName),
					WithScope(scope.Name),
					WithRemoteRef("foo-subnet"),
					WithKcpGcpSubnetSpecCidr("10.250.0.0/24"),
					WithKcpGcpSubnetPurposePrivate(),
				).
				Should(Succeed())

			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpGcpSubnet,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
				).
				Should(Succeed(), "Expected KCP GcpSubnet to become ready")
		})

		redisCluster := &cloudcontrolv1beta1.GcpRedisCluster{}
		clusterNodeType := "REDIS_SHARED_CORE_NANO"
		clusterShardCount := 3
		clusterReplicasCount := 1

		By("When GcpRedisCluster is created", func() {
			Eventually(CreateKcpGcpRedisCluster).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster,
					WithName(name),
					WithRemoteRef("skr-rediscluster-example"),
					WithGcpSubnet(kcpGcpSubnetName),
					WithScope(scope.Name),
					WithKcpGcpRedisClusterNodeType(clusterNodeType),
					WithKcpGcpRedisClusterShardCount(int32(clusterShardCount)),
					WithKcpGcpRedisClusterReplicasPerShard(int32(clusterReplicasCount)),
					WithKcpGcpRedisClusterConfigs(map[string]string{
						"maxmemory-policy": "allkeys-lru",
					}),
				).
				Should(Succeed(), "failed creating GcpRedisCluster")
		})

		By("Then GCP Redis is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster,
					NewObjActions(),
					HavingFieldSet("status", "id")).
				Should(Succeed(), "expected GcpRedisCluster to get status.id")
		})

		var createOpName string
		By("When GCP Redis create operation is resolved", func() {
			Eventually(func() error {
				it := gcpMock.ListRedisClusterOperations(infra.Ctx(), &longrunningpb.ListOperationsRequest{})
				for op, err := it.Next(); err == nil; op, err = it.Next() {
					if !op.Done && op.Name != "" {
						createOpName = op.Name
						return gcpMock.ResolveRedisClusterOperation(infra.Ctx(), createOpName)
					}
				}
				return fmt.Errorf("no pending create operation found yet")
			}).Should(Succeed(), "expected to find and resolve create operation")
		})

		By("Then GcpRedisCluster has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
					HavingState("Ready"),
				).
				Should(Succeed(), "expected GcpRedisCluster to has Ready state, but it didn't")
		})

		By("And Then GcpRedisCluster has .status.discoveryEndpoint set", func() {
			Expect(len(redisCluster.Status.DiscoveryEndpoint) > 0).To(Equal(true))
		})

		By("And Then GcpRedisCluster has .status.nodeType set", func() {
			Expect(redisCluster.Status.NodeType).To(Equal(clusterNodeType))
		})

		By("And Then GcpRedisCluster has .status.shardCount set", func() {
			Expect(redisCluster.Status.ShardCount).To(Equal(int32(clusterShardCount)))
		})

		By("And Then GcpRedisCluster has .status.replicasPerShard set", func() {
			Expect(redisCluster.Status.ReplicasPerShard).To(Equal(int32(clusterReplicasCount)))
		})

		// DELETE

		By("When GcpRedisCluster is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster).
				Should(Succeed(), "failed deleting GcpRedisCluster")
		})

		By("And When GCP Redis delete operation is resolved", func() {
			Eventually(func() error {
				it := gcpMock.ListRedisClusterOperations(infra.Ctx(), &longrunningpb.ListOperationsRequest{})
				for op, err := it.Next(); err == nil; op, err = it.Next() {
					if !op.Done && op.Name != "" && op.Name != createOpName {
						return gcpMock.ResolveRedisClusterOperation(infra.Ctx(), op.Name)
					}
				}
				return fmt.Errorf("no pending delete operation found yet")
			}).Should(Succeed(), "expected to find and resolve delete operation")
		})

		By("Then GcpRedisCluster does not exist", func() {
			Eventually(IsDeleted, 5*time.Second).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster).
				Should(Succeed(), "expected GcpRedisCluster not to exist (be deleted), but it still exists")
		})
	})

})
