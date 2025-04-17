package cloudresources

import (
	"fmt"

	"github.com/kyma-project/cloud-manager/api"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	skrgcpsubnet "github.com/kyma-project/cloud-manager/pkg/skr/gcpsubnet"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	"github.com/kyma-project/cloud-manager/pkg/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Feature: SKR GcpRedisCluster", func() {

	It("Scenario: SKR GcpRedisCluster is created with specified Subnet", func() {

		skrGcpSubnetName := "skr-e2125d55-9711-4a75-acba-acdb6a913a5b"
		skrGcpSubnet := &cloudresourcesv1beta1.GcpSubnet{}
		skrGcpSubnetId := "79adc53c-8fec-4e42-a9d1-a25ce93a9259"

		By("And Given SKR GcpSubnet exists", func() {
			// tell skrgcpsubnet reconciler to ignore this SKR GcpSubnet
			skrgcpsubnet.Ignore.AddName(skrGcpSubnetName)

			Eventually(CreateSkrGcpSubnet).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrGcpSubnet,
					WithName(skrGcpSubnetName),
				).
				Should(Succeed())
		})
		By("And Given SKR GcpSubnet has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrGcpSubnet,
					WithSkrGcpSubnetStatusCidr(skrGcpSubnet.Spec.Cidr),
					WithSkrGcpSubnetStatusId(skrGcpSubnetId),
					WithConditions(SkrReadyCondition()),
				).
				Should(Succeed())
		})

		gcpRedisClusterName := "custom-redis-cluster-123"
		gcpRedisCluster := &cloudresourcesv1beta1.GcpRedisCluster{}
		gpRedisClusterTier := cloudresourcesv1beta1.GcpRedisClusterTierC3
		shardCount := 3
		replicasPerShard := 2
		configKey := "maxmemory-policy"
		configValue := "allkeys-lru"
		gcpRedisClusterRedisConfigs := map[string]string{
			configKey: configValue,
		}

		const (
			authSecretName = "custom-auth-secretname"
		)
		authSecretLabels := map[string]string{
			"foo": "1",
		}
		authSecretAnnotations := map[string]string{
			"bar": "2",
		}

		extraData := map[string]string{
			"foo":    "bar",
			"parsed": "{{.host}}:{{.port}}",
		}

		By("When GcpRedisCluster is created", func() {
			Eventually(CreateSkrGcpRedisCluster).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), gcpRedisCluster,
					WithName(gcpRedisClusterName),
					WithGcpSubnet(skrGcpSubnet.Name),
					WithSkrGcpRedisClusterRedisTier(gpRedisClusterTier),
					WithSkrGcpRedisClusterShardCount(int32(shardCount)),
					WithSkrGcpRedisClusterReplicasPerShard(int32(replicasPerShard)),
					WithSkrGcpRedisClusterRedisConfigs(gcpRedisClusterRedisConfigs),
					WithSkrGcpRedisClusterAuthSecretName(authSecretName),
					WithSkrGcpRedisClusterAuthSecretLabels(authSecretLabels),
					WithSkrGcpRedisClusterAuthSecretAnnotations(authSecretAnnotations),
					WithSkrGcpRedisClusterAuthSecretExtraData(extraData),
				).
				Should(Succeed())
		})

		kcpRedisCluster := &cloudcontrolv1beta1.GcpRedisCluster{}

		By("Then KCP RedisCluster is created", func() {
			// load SKR GcpRedisCluster to get ID
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					gcpRedisCluster,
					NewObjActions(),
					HavingSkrGcpRedisClusterStatusId(),
					HavingSkrGcpRedisClusterStatusState(cloudresourcesv1beta1.StateCreating),
				).
				Should(Succeed(), "expected SKR GcpRedisCluster to get status.id")

			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpRedisCluster,
					NewObjActions(
						WithName(gcpRedisCluster.Status.Id),
					),
				).
				Should(Succeed())

			By("And has annotaton cloud-manager.kyma-project.io/kymaName")
			Expect(kcpRedisCluster.Annotations[cloudcontrolv1beta1.LabelKymaName]).To(Equal(infra.SkrKymaRef().Name))

			By("And has annotaton cloud-manager.kyma-project.io/remoteName")
			Expect(kcpRedisCluster.Annotations[cloudcontrolv1beta1.LabelRemoteName]).To(Equal(gcpRedisCluster.Name))

			By("And has annotaton cloud-manager.kyma-project.io/remoteNamespace")
			Expect(kcpRedisCluster.Annotations[cloudcontrolv1beta1.LabelRemoteNamespace]).To(Equal(gcpRedisCluster.Namespace))

			By("And has spec.scope.name equal to SKR Cluster kyma name")
			Expect(kcpRedisCluster.Spec.Scope.Name).To(Equal(infra.SkrKymaRef().Name))

			By("And has spec.remoteRef matching to to SKR GcpSubnet")
			Expect(kcpRedisCluster.Spec.RemoteRef.Namespace).To(Equal(gcpRedisCluster.Namespace))
			Expect(kcpRedisCluster.Spec.RemoteRef.Name).To(Equal(gcpRedisCluster.Name))

			By("And has spec equal to SKR GcpRedisCluster.spec values")

			Expect(kcpRedisCluster.Spec.NodeType).To(Not(Equal("")))
			Expect(kcpRedisCluster.Spec.ReplicasPerShard).To(Equal(int32(2)))
			Expect(kcpRedisCluster.Spec.ShardCount).To(Equal(int32(3)))
			Expect(kcpRedisCluster.Spec.RedisConfigs[configKey]).To(Equal(configValue))

			By("And has spec.subnet.name equal to SKR GcpSubnet.status.id")
			Expect(kcpRedisCluster.Spec.Subnet.Name).To(Equal(skrGcpSubnet.Status.Id))
		})

		kcpRedisClusterDiscoveryEndpoint := "192.168.0.1:6576"

		By("When KCP RedisCluster has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpRedisCluster,
					WithKcpGcpRedisClusterDiscoveryEndpoint(kcpRedisClusterDiscoveryEndpoint),

					WithConditions(KcpReadyCondition()),
				).
				Should(Succeed())
		})

		By("Then SKR GcpRedisCluster has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					gcpRedisCluster,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
					HavingSkrGcpRedisClusterStatusState(cloudresourcesv1beta1.StateReady),
				).
				Should(Succeed())
		})

		authSecret := &corev1.Secret{}
		By("And Then SKR auth Secret is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					authSecret,
					NewObjActions(
						WithName(authSecretName),
						WithNamespace(gcpRedisCluster.Namespace),
					),
				).
				Should(Succeed())

			By("And it has defined cloud-manager default labels")
			Expect(authSecret.Labels[util.WellKnownK8sLabelComponent]).ToNot(BeNil())
			Expect(authSecret.Labels[util.WellKnownK8sLabelPartOf]).ToNot(BeNil())
			Expect(authSecret.Labels[util.WellKnownK8sLabelManagedBy]).ToNot(BeNil())

			By("And it has defined ownmership label")
			Expect(authSecret.Labels[cloudresourcesv1beta1.LabelRedisClusterStatusId]).To(Equal(gcpRedisCluster.Status.Id))

			By("And it has user defined custom labels")
			for k, v := range authSecretLabels {
				Expect(authSecret.Labels).To(HaveKeyWithValue(k, v), fmt.Sprintf("expected auth Secret to have label %s=%s", k, v))
			}
			By("And it has user defined custom annotations")
			for k, v := range authSecretAnnotations {
				Expect(authSecret.Annotations).To(HaveKeyWithValue(k, v), fmt.Sprintf("expected auth Secret to have annotation %s=%s", k, v))
			}

			By("And it has user defined custom extraData")
			Expect(authSecret.Data).To(HaveKeyWithValue("foo", []byte("bar")), "expected auth secret data to have foo=bar")
			Expect(authSecret.Data).To(HaveKeyWithValue("parsed", []byte(kcpRedisClusterDiscoveryEndpoint)), "expected auth secret data to have parsed=host:port")

			By("And it has defined cloud-manager finalizer")
			Expect(authSecret.Finalizers).To(ContainElement(api.CommonFinalizerDeletionHook))
		})

		// CleanUp
		Eventually(Delete).
			WithArguments(infra.Ctx(), infra.SKR().Client(), gcpRedisCluster).
			Should(Succeed())
		Eventually(Delete).
			WithArguments(infra.Ctx(), infra.SKR().Client(), skrGcpSubnet).
			Should(Succeed())
	})

	It("Scenario: SKR GcpRedisCluster is deleted", func() {

		skrGcpSubnetName := "skr-a75d5e87-3b25-48d2-a273-3b0d7b0139f2"
		skrGcpSubnet := &cloudresourcesv1beta1.GcpSubnet{}
		skrGcpSubnetId := "e0eb6d25-7198-43f8-9a39-221fc2277c45"

		By("And Given SKR GcpSubnet exists", func() {
			// tell skrgcpsubnet reconciler to ignore this SKR GcpSubnet
			skrgcpsubnet.Ignore.AddName(skrGcpSubnetName)

			Eventually(CreateSkrGcpSubnet).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrGcpSubnet,
					WithName(skrGcpSubnetName),
				).
				Should(Succeed())
		})
		By("And Given SKR GcpSubnet has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrGcpSubnet,
					WithSkrGcpSubnetStatusCidr(skrGcpSubnet.Spec.Cidr),
					WithSkrGcpSubnetStatusId(skrGcpSubnetId),
					WithConditions(SkrReadyCondition()),
				).
				Should(Succeed())
		})

		gcpRedisClusterName := "another-gcp-redis-instance"
		gcpRedisCluster := &cloudresourcesv1beta1.GcpRedisCluster{}

		By("And Given GcpRedisCluster is created", func() {
			Eventually(CreateSkrGcpRedisCluster).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), gcpRedisCluster,
					WithName(gcpRedisClusterName),
					WithGcpSubnet(skrGcpSubnet.Name),
					WithSkrGcpRedisClusterDefaultSpec(),
				).
				Should(Succeed())
		})

		kcpRedisCluster := &cloudcontrolv1beta1.GcpRedisCluster{}

		By("And Given KCP RedisCluster is created", func() {
			// load SKR GcpRedisCluster to get ID
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					gcpRedisCluster,
					NewObjActions(),
					HavingSkrGcpRedisClusterStatusId(),
					HavingSkrGcpRedisClusterStatusState(cloudresourcesv1beta1.StateCreating),
				).
				Should(Succeed(), "expected SKR GcpRedisCluster to get status.id")

			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpRedisCluster,
					NewObjActions(
						WithName(gcpRedisCluster.Status.Id),
					),
				).
				Should(Succeed(), "expected KCP RedisCluster to be created, but it was not")

			Eventually(Update).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpRedisCluster, AddFinalizer(api.CommonFinalizerDeletionHook)).
				Should(Succeed(), "failed adding finalizer on KCP RedisCluster")
		})

		By("And Given KCP RedisCluster has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpRedisCluster,
					WithConditions(KcpReadyCondition()),
				).
				Should(Succeed(), "failed setting KCP RedisCluster Ready condition")
		})

		By("And Given SKR GcpRedisCluster has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					gcpRedisCluster,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
					HavingSkrGcpRedisClusterStatusState(cloudresourcesv1beta1.StateReady),
				).
				Should(Succeed(), "expected GcpRedisCluster to exist and have Ready condition")
		})

		authSecret := &corev1.Secret{}
		By("And Given SKR auth Secret is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					authSecret,
					NewObjActions(
						WithName(gcpRedisCluster.Name),
						WithNamespace(gcpRedisCluster.Namespace),
					),
				).
				Should(Succeed(), "failed creating auth Secret")
		})

		// DELETE START HERE

		By("When GcpRedisCluster is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), gcpRedisCluster).
				Should(Succeed(), "failed deleting GcpRedisCluster")
		})

		By("Then SKR GcpRedisCluster has Deleting state", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					gcpRedisCluster,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.StateDeleting),
					HavingSkrGcpRedisClusterStatusState(cloudresourcesv1beta1.StateDeleting),
				).
				Should(Succeed(), "expected GcpRedisCluster to have Deleting state")
		})

		By("And Then SKR auth Secret is deleted", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), authSecret).
				Should(Succeed(), "expected authSecret not to exist")
		})

		By("And Then KCP RedisCluster is marked for deletion", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpRedisCluster, NewObjActions(), HavingDeletionTimestamp()).
				Should(Succeed(), "expected KCP RedisCluster to be marked for deletion")
		})

		By("When KCP RedisCluster finalizer is removed and it is deleted", func() {
			Eventually(Update).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpRedisCluster, RemoveFinalizer(api.CommonFinalizerDeletionHook)).
				Should(Succeed(), "failed removing finalizer on KCP RedisCluster")
		})

		By("Then SKR GcpRedisCluster is deleted", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), gcpRedisCluster).
				Should(Succeed(), "expected GcpRedisCluster not to exist")
		})

		// CleanUp
		Eventually(Delete).
			WithArguments(infra.Ctx(), infra.SKR().Client(), skrGcpSubnet).
			Should(Succeed())
	})

})
