package cloudresources

import (
	"fmt"
	"time"

	"github.com/kyma-project/cloud-manager/api"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	azureutil "github.com/kyma-project/cloud-manager/pkg/skr/azurerediscluster"
	skriprange "github.com/kyma-project/cloud-manager/pkg/skr/iprange"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	"github.com/kyma-project/cloud-manager/pkg/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Feature: SKR AzureRedisCluster", func() {

	It("Scenario: SKR AzureRedisCluster is created", func() {

		azureRedisClusterName := "custom-redis-cluster"
		skrIpRangeId := "5c70629f-a13f-4b04-af47-1ab274c1c7rt"
		azureRedisCluster := &cloudresourcesv1beta1.AzureRedisCluster{}
		redisVersion := "6.0"
		tier := cloudresourcesv1beta1.AzureRedisTierC4
		var shardCount int32 = 2
		var replicaCount int32 = 4
		azureRedisClusterRedisConfigs := cloudresourcesv1beta1.RedisClusterAzureConfigs{}
		azureRedisClusterRedisConfigs.MaxClients = "5"
		skrIpRange := &cloudresourcesv1beta1.IpRange{}

		skriprange.Ignore.AddName("default")

		const (
			authSecretName = "azure-custom-auth-secretname"
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

		By("Given default SKR IpRange does not exist", func() {
			Consistently(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange,
					NewObjActions(WithName("default"), WithNamespace("kyma-system"))).
				ShouldNot(Succeed())
		})

		By("When AzureRedisCluster is created", func() {
			Eventually(CreateAzureRedisCluster).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), azureRedisCluster,
					WithName(azureRedisClusterName),
					WithAzureRedisClusterRedisVersion(redisVersion),
					WithAzureRedisClusterRedisTier(tier),
					WithAzureRedisClusterShardCount(shardCount),
					WithAzureRedisClusterReplicaCount(replicaCount),
					WithAzureRedisClusterRedisConfigs(azureRedisClusterRedisConfigs),
					WithAzureRedisClusterAuthSecretName(authSecretName),
					WithAzureRedisClusterAuthSecretLabels(authSecretLabels),
					WithAzureRedisClusterAuthSecretAnnotations(authSecretAnnotations),
					WithAzureRedisClusterAuthSecretExtraData(extraData),
				).
				Should(Succeed())
		})

		By("Then default SKR IpRange is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange,
					NewObjActions(WithName("default"), WithNamespace("kyma-system"))).
				Should(Succeed())
		})

		By("When default SKR IpRange has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrIpRange,
					WithSkrIpRangeStatusId(skrIpRangeId),
					WithConditions(SkrReadyCondition()),
				).
				Should(Succeed())
		})

		kcpRedisCluster := &cloudcontrolv1beta1.RedisCluster{}

		By("Then KCP RedisCluster is created", func() {
			// load SKR AzureRedisCluster to get ID
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					azureRedisCluster,
					NewObjActions(),
					HavingAzureRedisClusterStatusId(),
					HavingAzureRedisClusterStatusState(cloudresourcesv1beta1.StateCreating),
				).
				Should(Succeed())

			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpRedisCluster,
					NewObjActions(
						WithName(azureRedisCluster.Status.Id),
					),
				).
				Should(Succeed())

			By("And has annotaton cloud-manager.kyma-project.io/kymaName")
			Expect(kcpRedisCluster.Annotations[cloudcontrolv1beta1.LabelKymaName]).To(Equal(infra.SkrKymaRef().Name))

			By("And has annotaton cloud-manager.kyma-project.io/remoteName")
			Expect(kcpRedisCluster.Annotations[cloudcontrolv1beta1.LabelRemoteName]).To(Equal(azureRedisCluster.Name))

			By("And has annotaton cloud-manager.kyma-project.io/remoteNamespace")
			Expect(kcpRedisCluster.Annotations[cloudcontrolv1beta1.LabelRemoteNamespace]).To(Equal(azureRedisCluster.Namespace))

			By("And has spec.scope.name equal to SKR Cluster kyma name")
			Expect(kcpRedisCluster.Spec.Scope.Name).To(Equal(infra.SkrKymaRef().Name))

			By("And has spec.remoteRef matching to to SKR IpRange")
			Expect(kcpRedisCluster.Spec.RemoteRef.Namespace).To(Equal(azureRedisCluster.Namespace))
			Expect(kcpRedisCluster.Spec.RemoteRef.Name).To(Equal(azureRedisCluster.Name))

			By("And has spec.cluster.azure equal to SKR AzureRedisCluster.spec values")
			expectedSKUCapacity := 2
			redisSKUCapacity, _ := azureutil.RedisTierToSKUCapacityConverter(azureRedisCluster.Spec.RedisTier)
			Expect(expectedSKUCapacity).To(Equal(redisSKUCapacity))
			Expect(kcpRedisCluster.Spec.Instance.Azure.SKU.Capacity).To(Equal(redisSKUCapacity))
			Expect(kcpRedisCluster.Spec.Instance.Azure.RedisVersion).To(Equal(azureRedisCluster.Spec.RedisVersion))
			Expect(kcpRedisCluster.Spec.Instance.Azure.RedisConfiguration.MaxClients).To(Equal(azureRedisCluster.Spec.RedisConfiguration.MaxClients))
		})

		kcpRedisClusterPrimaryEndpoint := "192.168.0.1:6576"
		kcpRedisClusterAuthString := "a9461793-2449-48d2-8618-0881bbe61d07"

		By("When KCP RedisCluster has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpRedisCluster,
					WithRedisInstanceDiscoveryEndpoint(kcpRedisClusterPrimaryEndpoint),
					WithRedisInstanceAuthString(kcpRedisClusterAuthString),

					WithConditions(KcpReadyCondition()),
				).
				Should(Succeed())
		})

		By("Then SKR AzureRedisCluster has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					azureRedisCluster,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
					HavingAzureRedisClusterStatusState(cloudresourcesv1beta1.StateReady),
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
						WithNamespace(azureRedisCluster.Namespace),
					),
				).
				Should(Succeed())

			By("And it has defined cloud-manager default labels")
			Expect(authSecret.Labels[util.WellKnownK8sLabelComponent]).ToNot(BeNil())
			Expect(authSecret.Labels[util.WellKnownK8sLabelPartOf]).ToNot(BeNil())
			Expect(authSecret.Labels[util.WellKnownK8sLabelManagedBy]).ToNot(BeNil())

			By("And it has defined ownmership label")
			Expect(authSecret.Labels[cloudresourcesv1beta1.LabelRedisClusterStatusId]).To(Equal(azureRedisCluster.Status.Id))

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
			Expect(authSecret.Data).To(HaveKeyWithValue("parsed", []byte(kcpRedisClusterPrimaryEndpoint)), "expected auth secret data to have parsed=host:port")

			By("And it has defined cloud-manager finalizer")
			Expect(authSecret.Finalizers).To(ContainElement(api.CommonFinalizerDeletionHook))
		})

		// CleanUp
		Eventually(Delete).
			WithArguments(infra.Ctx(), infra.SKR().Client(), azureRedisCluster).
			Should(Succeed())

		By("// cleanup: delete default SKR IpRange", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange).
				Should(Succeed())
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange).
				Should(Succeed())
		})
	})

	It("Scenario: SKR AzureRedisCluster is modified", func() {

		azureRedisClusterName := "modified-redis-cluster"
		skrIpRangeId := "5c70629f-a13f-4b04-af47-1ab274c1c7rt"
		azureRedisCluster := &cloudresourcesv1beta1.AzureRedisCluster{}
		redisVersion := "6.0"
		tier := cloudresourcesv1beta1.AzureRedisTierC4
		azureRedisClusterRedisConfigs := cloudresourcesv1beta1.RedisClusterAzureConfigs{}
		azureRedisClusterRedisConfigs.MaxClients = "5"
		skrIpRange := &cloudresourcesv1beta1.IpRange{}

		skriprange.Ignore.AddName("default")

		const (
			authSecretName = "azure-custom-auth-secretname"
		)
		authSecretLabels := map[string]string{
			"foo": "1",
		}
		authSecretAnnotations := map[string]string{
			"bar": "2",
		}

		By("Given default SKR IpRange does not exist", func() {
			Consistently(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange,
					NewObjActions(WithName("default"), WithNamespace("kyma-system"))).
				ShouldNot(Succeed())
		})

		By("Given AzureRedisCluster exists", func() {
			Eventually(CreateAzureRedisCluster).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), azureRedisCluster,
					WithName(azureRedisClusterName),
					WithAzureRedisClusterRedisVersion(redisVersion),
					WithAzureRedisClusterRedisTier(tier),
					WithAzureRedisClusterRedisConfigs(azureRedisClusterRedisConfigs),
					WithAzureRedisClusterAuthSecretName(authSecretName),
					WithAzureRedisClusterAuthSecretLabels(authSecretLabels),
					WithAzureRedisClusterAuthSecretAnnotations(authSecretAnnotations),
				).
				Should(Succeed())
		})

		By("Then default SKR IpRange is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange,
					NewObjActions(WithName("default"), WithNamespace("kyma-system"))).
				Should(Succeed())
		})

		By("When default SKR IpRange has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrIpRange,
					WithSkrIpRangeStatusId(skrIpRangeId),
					WithConditions(SkrReadyCondition()),
				).
				Should(Succeed())
		})

		kcpRedisCluster := &cloudcontrolv1beta1.RedisCluster{}

		By("And RedisCluster exists", func() {
			// load SKR AzureRedisCluster to get ID
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					azureRedisCluster,
					NewObjActions(),
					HavingAzureRedisClusterStatusId(),
					HavingAzureRedisClusterStatusState(cloudresourcesv1beta1.StateCreating),
				).
				Should(Succeed())

			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpRedisCluster,
					NewObjActions(
						WithName(azureRedisCluster.Status.Id),
					),
				).
				Should(Succeed())

			By("And has annotaton cloud-manager.kyma-project.io/kymaName")
			Expect(kcpRedisCluster.Annotations[cloudcontrolv1beta1.LabelKymaName]).To(Equal(infra.SkrKymaRef().Name))

			By("And has annotaton cloud-manager.kyma-project.io/remoteName")
			Expect(kcpRedisCluster.Annotations[cloudcontrolv1beta1.LabelRemoteName]).To(Equal(azureRedisCluster.Name))

			By("And has annotaton cloud-manager.kyma-project.io/remoteNamespace")
			Expect(kcpRedisCluster.Annotations[cloudcontrolv1beta1.LabelRemoteNamespace]).To(Equal(azureRedisCluster.Namespace))

			By("And has spec.scope.name equal to SKR Cluster kyma name")
			Expect(kcpRedisCluster.Spec.Scope.Name).To(Equal(infra.SkrKymaRef().Name))

			By("And has spec.remoteRef matching to to SKR IpRange")
			Expect(kcpRedisCluster.Spec.RemoteRef.Namespace).To(Equal(azureRedisCluster.Namespace))
			Expect(kcpRedisCluster.Spec.RemoteRef.Name).To(Equal(azureRedisCluster.Name))

			By("And has spec.cluster.azure equal to SKR AzureRedisCluster.spec values")
			redisSKUCapacity, _ := azureutil.RedisTierToSKUCapacityConverter(azureRedisCluster.Spec.RedisTier)
			Expect(kcpRedisCluster.Spec.Instance.Azure.SKU.Capacity).To(Equal(redisSKUCapacity))
			Expect(kcpRedisCluster.Spec.Instance.Azure.RedisVersion).To(Equal(azureRedisCluster.Spec.RedisVersion))
			Expect(kcpRedisCluster.Spec.Instance.Azure.RedisConfiguration.MaxClients).To(Equal(azureRedisCluster.Spec.RedisConfiguration.MaxClients))
		})

		tier = cloudresourcesv1beta1.AzureRedisTierC4

		By("When AzureRedisCluster is modified", func() {
			Eventually(UpdateAzureRedisCluster).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), azureRedisCluster,
					WithAzureRedisClusterRedisTier(tier),
				).
				Should(Succeed())
		})

		By("And AzureRedsiCluster SKU.Capacity has modified value")
		Expect(azureRedisCluster.Spec.RedisTier).To(Equal(tier))

		By("Then KCP RedisCluster SKU.Capacity is modified", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpRedisCluster,
					NewObjActions(
						WithName(azureRedisCluster.Status.Id),
					),
				).
				Should(Succeed())

			By("And KCP RedisCluster SKU.Capacity has modified value")
			Expect(azureRedisCluster.Spec.RedisTier).To(Equal(tier))
		})

		// CleanUp
		Eventually(Delete).
			WithArguments(infra.Ctx(), infra.SKR().Client(), azureRedisCluster).
			Should(Succeed())

		By("// cleanup: delete default SKR IpRange", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange).
				Should(Succeed())
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange).
				Should(Succeed())
		})
	})

	It("Scenario: SKR AzureRedisCluster is deleted", func() {

		azureRedisClusterName := "another-azure-redis-cluster"
		skrIpRangeId := "5c70629f-a13f-4b04-af47-1ab274c1c7rcr"
		azureRedisCluster := &cloudresourcesv1beta1.AzureRedisCluster{}
		tier := cloudresourcesv1beta1.AzureRedisTierC4
		skrIpRange := &cloudresourcesv1beta1.IpRange{}

		skriprange.Ignore.AddName("default")

		By("Given default SKR IpRange does not exist", func() {
			Consistently(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange,
					NewObjActions(WithName("default"), WithNamespace("kyma-system"))).
				ShouldNot(Succeed())
		})

		By("Given AzureRedisCluster is created", func() {
			Eventually(CreateAzureRedisCluster).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), azureRedisCluster,
					WithName(azureRedisClusterName),
					WithAzureRedisClusterRedisTier(tier),
					WithAzureRedisClusterRedisVersion("6.0"),
				).
				Should(Succeed())
		})

		By("Then default SKR IpRange is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange,
					NewObjActions(WithName("default"), WithNamespace("kyma-system"))).
				Should(Succeed())
		})

		By("When default SKR IpRange has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrIpRange,
					WithSkrIpRangeStatusId(skrIpRangeId),
					WithConditions(SkrReadyCondition()),
				).
				Should(Succeed())
		})

		kcpRedisCluster := &cloudcontrolv1beta1.RedisCluster{}

		By("And Given KCP RedisCluster is created", func() {
			// load SKR AzureRedisCluster to get ID
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					azureRedisCluster,
					NewObjActions(),
					HavingAzureRedisClusterStatusId(),
					HavingAzureRedisClusterStatusState(cloudresourcesv1beta1.StateCreating),
				).
				Should(Succeed(), "expected SKR AzureRedisCluster to get status.id")

			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpRedisCluster,
					NewObjActions(
						WithName(azureRedisCluster.Status.Id),
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

		By("And Given SKR AzureRedisCluster has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					azureRedisCluster,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
					HavingAzureRedisClusterStatusState(cloudresourcesv1beta1.StateReady),
				).
				Should(Succeed(), "expected AzureRedisCluster to exist and have Ready condition")
		})

		authSecret := &corev1.Secret{}
		By("And Given SKR auth Secret is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					authSecret,
					NewObjActions(
						WithName(azureRedisCluster.Name),
						WithNamespace(azureRedisCluster.Namespace),
					),
				).
				Should(Succeed(), "failed creating auth Secret")
		})

		// DELETE START HERE

		By("When AzureRedisCluster is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), azureRedisCluster).
				Should(Succeed(), "failed deleting AzureRedisCluster")
		})

		By("Then SKR AzureRedisCluster has Deleting state", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					azureRedisCluster,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.StateDeleting),
					HavingAzureRedisClusterStatusState(cloudresourcesv1beta1.StateDeleting),
				).
				Should(Succeed(), "expected AzureRedisCluster to have Deleting state")
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

		By("Then SKR AzureRedisCluster is deleted", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), azureRedisCluster).
				Should(Succeed(), "expected AzureRedisCluster not to exist")
		})

		By("// cleanup: delete default SKR IpRange", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange).
				Should(Succeed())
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange).
				Should(Succeed())
		})
	})

	It("Scenario: SKR AzureRedisCluster authSecret is modified", func() {
		azureRedisClusterName := "auth-secret-modified-redis"
		skrIpRangeId := "5c70629f-a13f-4b04-af47-1ab274c1c7ac"
		azureRedisCluster := &cloudresourcesv1beta1.AzureRedisCluster{}
		redisVersion := "6.0"
		tier := cloudresourcesv1beta1.AzureRedisTierC3
		var shardCount int32 = 1
		var replicaCount int32 = 2
		skrIpRange := &cloudresourcesv1beta1.IpRange{}

		skriprange.Ignore.AddName("default")

		const (
			authSecretName = "azure-cluster-auth-secret-test"
		)
		authSecretLabels := map[string]string{
			"env": "test",
		}
		authSecretAnnotations := map[string]string{
			"purpose": "testing",
		}

		By("Given default SKR IpRange does not exist", func() {
			Consistently(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange,
					NewObjActions(WithName("default"), WithNamespace("kyma-system"))).
				ShouldNot(Succeed())
		})

		By("And Given AzureRedisCluster is created with initial authSecret config", func() {
			Eventually(CreateAzureRedisCluster).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), azureRedisCluster,
					WithName(azureRedisClusterName),
					WithAzureRedisClusterRedisVersion(redisVersion),
					WithAzureRedisClusterRedisTier(tier),
					WithAzureRedisClusterShardCount(shardCount),
					WithAzureRedisClusterReplicaCount(replicaCount),
					WithAzureRedisClusterAuthSecretName(authSecretName),
					WithAzureRedisClusterAuthSecretLabels(authSecretLabels),
					WithAzureRedisClusterAuthSecretAnnotations(authSecretAnnotations),
				).
				Should(Succeed())
		})

		By("And Given default SKR IpRange is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange,
					NewObjActions(WithName("default"), WithNamespace("kyma-system"))).
				Should(Succeed())
		})

		By("And Given default SKR IpRange has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrIpRange,
					WithSkrIpRangeStatusId(skrIpRangeId),
					WithConditions(SkrReadyCondition()),
				).
				Should(Succeed())
		})

		kcpRedisCluster := &cloudcontrolv1beta1.RedisCluster{}

		By("And Given KCP RedisCluster is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					azureRedisCluster,
					NewObjActions(),
					HavingAzureRedisClusterStatusId(),
					HavingAzureRedisClusterStatusState(cloudresourcesv1beta1.StateCreating),
				).
				Should(Succeed())

			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpRedisCluster,
					NewObjActions(
						WithName(azureRedisCluster.Status.Id),
					),
				).
				Should(Succeed())
		})

		kcpRedisClusterPrimaryEndpoint := "10.0.0.2:6379"
		kcpRedisClusterAuthString := "initial-cluster-auth-string-67890"

		By("And Given KCP RedisCluster has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpRedisCluster,
					WithRedisInstanceDiscoveryEndpoint(kcpRedisClusterPrimaryEndpoint),
					WithRedisInstanceAuthString(kcpRedisClusterAuthString),
					WithConditions(KcpReadyCondition()),
				).
				Should(Succeed())
		})

		By("And Given SKR AzureRedisCluster has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					azureRedisCluster,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
					HavingAzureRedisClusterStatusState(cloudresourcesv1beta1.StateReady),
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
						WithNamespace(azureRedisCluster.Namespace),
					),
				).
				Should(Succeed())

			By("And it has initial labels")
			Expect(authSecret.Labels).To(HaveKeyWithValue("env", "test"))

			By("And it has initial annotations")
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
			"custom-key": "custom-value",
			"endpoint":   "{{.host}}:{{.port}}",
		}

		By("When AzureRedisCluster authSecret config is modified with new labels, annotations, and extraData", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					azureRedisCluster,
					NewObjActions(),
				).
				Should(Succeed())

			azureRedisCluster.Spec.AuthSecret.Labels = newLabels
			azureRedisCluster.Spec.AuthSecret.Annotations = newAnnotations

			Eventually(UpdateAzureRedisCluster).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), azureRedisCluster,
					WithAzureRedisClusterAuthSecretExtraData(newExtraData),
				).
				Should(Succeed())
		})

		By("Then SKR auth Secret is updated with new labels", func() {
			Eventually(func() map[string]string {
				secretKey := types.NamespacedName{Name: authSecretName, Namespace: azureRedisCluster.Namespace}
				err := infra.SKR().Client().Get(infra.Ctx(), secretKey, authSecret)
				if err != nil {
					return nil
				}
				userLabels := map[string]string{}
				for k, v := range authSecret.Labels {
					if k == "env" || k == "team" {
						userLabels[k] = v
					}
				}
				return userLabels
			}).WithTimeout(20 * time.Second).WithPolling(200 * time.Millisecond).Should(And(
				HaveKeyWithValue("env", "production"),
				HaveKeyWithValue("team", "platform"),
				HaveLen(2),
			))

			Expect(authSecret.Labels).To(HaveKey(cloudresourcesv1beta1.LabelRedisClusterStatusId))
			Expect(authSecret.Labels).To(HaveKey(cloudresourcesv1beta1.LabelCloudManaged))
		})

		By("And Then SKR auth Secret has new annotations", func() {
			Eventually(func() map[string]string {
				secretKey := types.NamespacedName{Name: authSecretName, Namespace: azureRedisCluster.Namespace}
				err := infra.SKR().Client().Get(infra.Ctx(), secretKey, authSecret)
				if err != nil {
					return nil
				}
				return authSecret.Annotations
			}).WithTimeout(20 * time.Second).WithPolling(200 * time.Millisecond).Should(And(
				HaveKeyWithValue("purpose", "production-testing"),
				HaveKeyWithValue("cost-center", "12345"),
				HaveLen(2),
			))
		})

		By("And Then auth Secret data includes extraData fields", func() {
			Eventually(func() map[string][]byte {
				secretKey := types.NamespacedName{Name: authSecretName, Namespace: azureRedisCluster.Namespace}
				err := infra.SKR().Client().Get(infra.Ctx(), secretKey, authSecret)
				if err != nil {
					return nil
				}
				return authSecret.Data
			}).WithTimeout(20 * time.Second).WithPolling(200 * time.Millisecond).Should(And(
				HaveKeyWithValue("custom-key", []byte("custom-value")),
				HaveKeyWithValue("endpoint", []byte(kcpRedisClusterPrimaryEndpoint)),
				HaveKey("host"),
				HaveKey("port"),
				HaveKey("authString"),
			))
		})

		Eventually(Delete).
			WithArguments(infra.Ctx(), infra.SKR().Client(), azureRedisCluster).
			Should(Succeed())

		By("// cleanup: delete default SKR IpRange", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange).
				Should(Succeed())
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange).
				Should(Succeed())
		})
	})

})
