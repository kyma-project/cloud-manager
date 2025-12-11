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
	"k8s.io/apimachinery/pkg/api/meta"
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

	It("Scenario: SKR GcpRedisCluster is created with empty GcpSubnet when default GcpSubnet does not exist", func() {

		gcpGcpRedisClusterName := "7b6e392f-bfe8-4f7f-8fbc-318aad1d3cba"
		skrGcpSubnetId := "5170ec2e-e829-4c94-8c33-f791c35aa984"
		gcpGcpRedisCluster := &cloudresourcesv1beta1.GcpRedisCluster{}
		kcpGcpRedisCluster := &cloudcontrolv1beta1.GcpRedisCluster{}
		skrGcpSubnet := &cloudresourcesv1beta1.GcpSubnet{}

		skrgcpsubnet.Ignore.AddName("default")

		By("Given default SKR GcpSubnet does not exist", func() {
			Consistently(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrGcpSubnet,
					NewObjActions(WithName("default"))).
				ShouldNot(Succeed())
		})

		By("When GcpRedisCluster is created with empty GcpSubnet", func() {
			Eventually(CreateSkrGcpRedisCluster).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), gcpGcpRedisCluster,
					WithName(gcpGcpRedisClusterName),
					WithSkrGcpRedisClusterDefaultSpec(),
				).
				Should(Succeed())
		})

		By("Then default SKR GcpSubnet is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrGcpSubnet,
					NewObjActions(WithName("default"), WithNamespace("kyma-system"))).
				Should(Succeed())
		})

		By("And Then default SKR GcpSubnet has label app.kubernetes.io/managed-by: cloud-manager", func() {
			Expect(skrGcpSubnet.Labels[util.WellKnownK8sLabelManagedBy]).To(Equal("cloud-manager"))
		})

		By("And Then default SKR GcpSubnet has label app.kubernetes.io/part-of: kyma", func() {
			Expect(skrGcpSubnet.Labels[util.WellKnownK8sLabelPartOf]).To(Equal("kyma"))
		})

		By("And Then GcpRedisCluster is not ready", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), gcpGcpRedisCluster, NewObjActions()).
				Should(Succeed())
			Expect(meta.IsStatusConditionTrue(gcpGcpRedisCluster.Status.Conditions, cloudresourcesv1beta1.ConditionTypeReady)).
				To(BeFalse(), "expected GcpRedisCluster not to have Ready condition, but it has")
		})

		By("When default SKR GcpSubnet has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrGcpSubnet,
					WithSkrGcpSubnetStatusId(skrGcpSubnetId),
					WithConditions(SkrReadyCondition()),
				).
				Should(Succeed())
		})

		By("Then KCP GcpRedisCluster is created", func() {
			// load SKR GcpRedisCluster to get ID
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					gcpGcpRedisCluster,
					NewObjActions(),
					HavingSkrGcpRedisClusterStatusId(),
					HavingSkrGcpRedisClusterStatusState(cloudresourcesv1beta1.StateCreating),
				).
				Should(Succeed(), "expected SKR GcpRedisCluster to get status.id and status creating")

			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpGcpRedisCluster,
					NewObjActions(
						WithName(gcpGcpRedisCluster.Status.Id),
					),
				).
				Should(Succeed())
		})

		By("When KCP GcpRedisCluster has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpGcpRedisCluster,
					WithKcpGcpRedisClusterDiscoveryEndpoint("192.168.0.1:6576"),
					WithConditions(KcpReadyCondition()),
				).
				Should(Succeed())
		})

		By("Then SKR GcpRedisCluster has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					gcpGcpRedisCluster,
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
						WithName(gcpGcpRedisCluster.Name),
						WithNamespace(gcpGcpRedisCluster.Namespace),
					),
				).
				Should(Succeed())
		})

		By("When GcpRedisCluster is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), gcpGcpRedisCluster).
				Should(Succeed())
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), gcpGcpRedisCluster).
				Should(Succeed(), "expected GcpRedisCluster not to exist, but it still does")
		})

		By("Then auth Secret does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), authSecret).
				Should(Succeed(), "expected auth Secret not to exist, but it still does")
		})

		By("And Then KCP GcpRedisCluster does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpGcpRedisCluster).
				Should(Succeed(), "expected KCP GcpRedisCluster not to exist, but it still does")
		})

		By("And Then SKR default GcpSubnet exists", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrGcpSubnet, NewObjActions()).
				Should(Succeed())
		})

		By("// cleanup: delete default SKR GcpSubnet", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrGcpSubnet).
				Should(Succeed())
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrGcpSubnet).
				Should(Succeed())
		})
	})

	It("Scenario: SKR GcpRedisCluster is created with empty GcpSubnetRef when default GcpSubnet already exist", func() {

		gcpGcpRedisClusterName := "d41efe57-665c-47fe-9f3c-af306fb8161f"
		skrGcpSubnetId := "154b6ecf-cc8c-4dac-b3d0-a32b38c98764"
		gcpGcpRedisCluster := &cloudresourcesv1beta1.GcpRedisCluster{}
		kcpGcpRedisCluster := &cloudcontrolv1beta1.GcpRedisCluster{}
		skrGcpSubnet := &cloudresourcesv1beta1.GcpSubnet{}

		skrgcpsubnet.Ignore.AddName("default")

		By("Given default SKR GcpSubnet exists", func() {
			Eventually(CreateSkrGcpSubnet).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrGcpSubnet, WithName("default")).
				Should(Succeed())
		})

		By("And Given default SKR GcpSubnet has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrGcpSubnet,
					WithSkrGcpSubnetStatusId(skrGcpSubnetId),
					WithConditions(SkrReadyCondition()),
				).
				Should(Succeed())
		})

		By("When GcpRedisCluster is created with empty GcpSubnetRef", func() {
			Eventually(CreateSkrGcpRedisCluster).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), gcpGcpRedisCluster,
					WithName(gcpGcpRedisClusterName),
					WithSkrGcpRedisClusterDefaultSpec(),
				).
				Should(Succeed())
		})

		By("Then KCP GcpRedisCluster is created", func() {
			// load SKR GcpRedisCluster to get ID
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					gcpGcpRedisCluster,
					NewObjActions(),
					HavingSkrGcpRedisClusterStatusId(),
					HavingSkrGcpRedisClusterStatusState(cloudresourcesv1beta1.StateCreating),
				).
				Should(Succeed(), "expected SKR GcpRedisCluster to get status.id and status creating")

			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpGcpRedisCluster,
					NewObjActions(
						WithName(gcpGcpRedisCluster.Status.Id),
					),
				).
				Should(Succeed())
		})

		By("When KCP GcpRedisCluster has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpGcpRedisCluster,
					WithKcpGcpRedisClusterDiscoveryEndpoint("192.168.0.1:6576"),
					WithConditions(KcpReadyCondition()),
				).
				Should(Succeed())
		})

		By("Then SKR GcpRedisCluster has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					gcpGcpRedisCluster,
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
						WithName(gcpGcpRedisCluster.Name),
						WithNamespace(gcpGcpRedisCluster.Namespace),
					),
				).
				Should(Succeed())
		})

		By("When GcpRedisCluster is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), gcpGcpRedisCluster).
				Should(Succeed())
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), gcpGcpRedisCluster).
				Should(Succeed(), "expected GcpRedisCluster not to exist, but it still does")
		})

		By("Then auth Secret does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), authSecret).
				Should(Succeed(), "expected auth Secret not to exist, but it still does")
		})

		By("And Then KCP GcpRedisCluster does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpGcpRedisCluster).
				Should(Succeed(), "expected KCP GcpRedisCluster not to exist, but it still does")
		})

		By("And Then SKR default GcpSubnet exists", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrGcpSubnet, NewObjActions()).
				Should(Succeed())
		})

		By("// cleanup: delete default SKR GcpSubnet", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrGcpSubnet).
				Should(Succeed())
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrGcpSubnet).
				Should(Succeed())
		})
	})

	It("Scenario: SKR GcpRedisCluster authSecret is modified", func() {
		gcpRedisClusterName := "auth-secret-modified-cluster"
		skrGcpSubnetId := "89ef2464-afa9-457a-a09e-8aac592cb7ff"
		gcpRedisCluster := &cloudresourcesv1beta1.GcpRedisCluster{}
		tier := cloudresourcesv1beta1.GcpRedisClusterTierC1
		skrGcpSubnet := &cloudresourcesv1beta1.GcpSubnet{}

		skrgcpsubnet.Ignore.AddName("default")

		const (
			authSecretName = "gcp-cluster-auth-secret-test"
		)
		authSecretLabels := map[string]string{
			"env": "test",
		}
		authSecretAnnotations := map[string]string{
			"purpose": "testing",
		}

		By("Given default SKR GcpSubnet does not exist", func() {
			Consistently(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrGcpSubnet,
					NewObjActions(WithName("default"), WithNamespace("kyma-system"))).
				ShouldNot(Succeed())
		})

		By("And Given GcpRedisCluster is created with initial authSecret config", func() {
			Eventually(CreateSkrGcpRedisCluster).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), gcpRedisCluster,
					WithName(gcpRedisClusterName),
					WithSkrGcpRedisClusterRedisTier(tier),
					WithSkrGcpRedisClusterShardCount(2),
					WithSkrGcpRedisClusterReplicasPerShard(1),
					WithSkrGcpRedisClusterAuthSecretName(authSecretName),
					WithSkrGcpRedisClusterAuthSecretLabels(authSecretLabels),
					WithSkrGcpRedisClusterAuthSecretAnnotations(authSecretAnnotations),
				).
				Should(Succeed())
		})

		By("And Given default SKR GcpSubnet is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrGcpSubnet,
					NewObjActions(WithName("default"), WithNamespace("kyma-system"))).
				Should(Succeed())
		})

		By("And Given default SKR GcpSubnet has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrGcpSubnet,
					WithSkrGcpSubnetStatusId(skrGcpSubnetId),
					WithConditions(SkrReadyCondition()),
				).
				Should(Succeed())
		})

		kcpRedisCluster := &cloudcontrolv1beta1.GcpRedisCluster{}

		By("And Given KCP RedisCluster is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					gcpRedisCluster,
					NewObjActions(),
					HavingSkrGcpRedisClusterStatusId(),
					HavingSkrGcpRedisClusterStatusState(cloudresourcesv1beta1.StateCreating),
				).
				Should(Succeed())

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
		})

		kcpRedisClusterPrimaryEndpoint := "10.0.0.2:6379"
		kcpRedisClusterAuthString := "cluster-auth-string-67890"

		By("And Given KCP GcpRedisCluster has Ready condition", func() {
			// Manually set authString since there's no DSL helper
			kcpRedisCluster.Status.AuthString = kcpRedisClusterAuthString

			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpRedisCluster,
					WithKcpGcpRedisClusterDiscoveryEndpoint(kcpRedisClusterPrimaryEndpoint),
					WithConditions(KcpReadyCondition()),
				).
				Should(Succeed())
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
				Should(Succeed())
		})

		authSecret := &corev1.Secret{}
		By("And Given SKR auth Secret is created with initial values", func() {
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

			Expect(authSecret.Labels).To(HaveKeyWithValue("env", "test"))
			Expect(authSecret.Annotations).To(HaveKeyWithValue("purpose", "testing"))
		})

		newLabels := map[string]string{
			"env":  "production",
			"team": "platform",
		}
		newAnnotations := map[string]string{
			"purpose":     "production-testing",
			"cost-center": "12345",
		}
		newExtraData := map[string]string{
			"custom-key":        "custom-value",
			"connection-string": "{{.primaryEndpoint}}",
		}

		By("When GcpRedisCluster authSecret config is modified with new labels, annotations, and extraData", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					gcpRedisCluster,
					NewObjActions(),
				).
				Should(Succeed())

			gcpRedisCluster.Spec.AuthSecret.Labels = newLabels
			gcpRedisCluster.Spec.AuthSecret.Annotations = newAnnotations
			gcpRedisCluster.Spec.AuthSecret.ExtraData = newExtraData

			Eventually(Update).
				WithArguments(infra.Ctx(), infra.SKR().Client(), gcpRedisCluster).
				Should(Succeed())
		})

		By("Then SKR auth Secret is updated with new labels, annotations, and extraData", func() {
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

			// Verify user labels (filter out system labels)
			userLabels := map[string]string{}
			for k, v := range authSecret.Labels {
				if k == "env" || k == "team" {
					userLabels[k] = v
				}
			}
			Expect(userLabels).To(And(
				HaveKeyWithValue("env", "production"),
				HaveKeyWithValue("team", "platform"),
				HaveLen(2),
			))
			Expect(authSecret.Labels).To(HaveKey(cloudresourcesv1beta1.LabelCloudManaged))

			// Verify annotations
			Expect(authSecret.Annotations).To(And(
				HaveKeyWithValue("purpose", "production-testing"),
				HaveKeyWithValue("cost-center", "12345"),
				HaveLen(2),
			))

			// Verify extraData
			Expect(authSecret.Data).To(And(
				HaveKeyWithValue("custom-key", []byte("custom-value")),
				HaveKeyWithValue("connection-string", []byte(kcpRedisClusterPrimaryEndpoint)),
				HaveKey("primaryEndpoint"),
				HaveKey("authString"),
			))
		})

		oldAuthSecret := authSecret.DeepCopy()
		newAuthSecretName := "gcp-cluster-auth-secret-renamed"

		By("When GcpRedisCluster authSecret name is changed", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					gcpRedisCluster,
					NewObjActions(),
				).
				Should(Succeed())

			gcpRedisCluster.Spec.AuthSecret.Name = newAuthSecretName

			Eventually(Update).
				WithArguments(infra.Ctx(), infra.SKR().Client(), gcpRedisCluster).
				Should(Succeed())
		})

		newAuthSecret := &corev1.Secret{}
		By("Then new SKR auth Secret is created with the new name", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					newAuthSecret,
					NewObjActions(
						WithName(newAuthSecretName),
						WithNamespace(gcpRedisCluster.Namespace),
					),
				).
				Should(Succeed())

			Expect(newAuthSecret.Data).To(And(
				HaveKey("primaryEndpoint"),
				HaveKey("authString"),
			))
		})

		By("And Then old SKR auth Secret is deleted", func() {
			Eventually(IsDeleted).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					oldAuthSecret,
				).
				Should(Succeed())
		})

		Eventually(Delete).
			WithArguments(infra.Ctx(), infra.SKR().Client(), gcpRedisCluster).
			Should(Succeed())

		By("// cleanup: delete default SKR GcpSubnet", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrGcpSubnet).
				Should(Succeed())
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrGcpSubnet).
				Should(Succeed())
		})
	})

})
