package cloudresources

import (
	"fmt"
	"time"

	"github.com/kyma-project/cloud-manager/api"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	skriprange "github.com/kyma-project/cloud-manager/pkg/skr/iprange"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	"github.com/kyma-project/cloud-manager/pkg/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
)

var _ = Describe("Feature: SKR AwsRedisCluster", func() {

	It("Scenario: SKR AwsRedisCluster is created with specified IpRange", func() {

		skrIpRangeName := "b9dd93f9-4acd-4f05-96d7-5ba5371b6b3b"
		skrIpRange := &cloudresourcesv1beta1.IpRange{}
		skrIpRangeId := "28e426ed-fb37-4388-a682-7d378662377f"

		By("And Given SKR IpRange exists", func() {
			// tell skriprange reconciler to ignore this SKR IpRange
			skriprange.Ignore.AddName(skrIpRangeName)

			Eventually(CreateSkrIpRange).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrIpRange,
					WithName(skrIpRangeName),
				).
				Should(Succeed())
		})
		By("And Given SKR IpRange has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrIpRange,
					WithSkrIpRangeStatusCidr(skrIpRange.Spec.Cidr),
					WithSkrIpRangeStatusId(skrIpRangeId),
					WithConditions(SkrReadyCondition()),
				).
				Should(Succeed())
		})

		awsRedisClusterName := "07d85fea-6005-4747-9800-831173d7c11b"
		awsRedisCluster := &cloudresourcesv1beta1.AwsRedisCluster{}

		const (
			authSecretName = "b35a05d0-c000-428d-b6b3-1f62b79631b1"
		)
		authSecretLabels := map[string]string{
			"foo": "1",
		}
		authSecretAnnotations := map[string]string{
			"bar": "2",
		}

		redisTier := cloudresourcesv1beta1.AwsRedisTierC5
		engineVersion := "6.x"
		autoMinorVersionUpgrade := true
		authEnabled := true
		shardCount := int32(10)
		replicasPerShard := int32(2)

		preferredMaintenanceWindow := ptr.To("sun:23:00-mon:01:30")

		parameterKey := "active-defrag-cycle-max"
		parameterValue := "85"
		parameters := map[string]string{
			parameterKey: parameterValue,
		}

		extraData := map[string]string{
			"foo":    "bar",
			"parsed": "{{.host}}:{{.port}}",
		}

		By("When AwsRedisCluster is created", func() {
			Eventually(CreateAwsRedisCluster).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), awsRedisCluster,
					WithName(awsRedisClusterName),
					WithIpRange(skrIpRange.Name),
					WithAwsRedisClusterAuthSecretName(authSecretName),
					WithAwsRedisClusterAuthSecretLabels(authSecretLabels),
					WithAwsRedisClusterAuthSecretAnnotations(authSecretAnnotations),
					WithAwsRedisClusterAuthSecretExtraData(extraData),
					WithAwsRedisClusterRedisTier(redisTier),
					WithAwsRedisClusterShardCount(shardCount),
					WithAwsRedisClusterReplicasPerShard(replicasPerShard),
					WithAwsRedisClusterEngineVersion(engineVersion),
					WithAwsRedisClusterAutoMinorVersionUpgrade(autoMinorVersionUpgrade),
					WithAwsRedisClusterAuthEnabled(authEnabled),
					WithAwsRedisClusterPreferredMaintenanceWindow(preferredMaintenanceWindow),
					WithAwsRedisClusterParameters(parameters),
				).
				Should(Succeed())
		})

		kcpRedisCluster := &cloudcontrolv1beta1.RedisCluster{}

		By("Then KCP RedisCluster is created", func() {
			// load SKR AwsRedisCluster to get ID
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					awsRedisCluster,
					NewObjActions(),
					HavingAwsRedisClusterStatusId(),
					HavingAwsRedisClusterStatusState(cloudresourcesv1beta1.StateCreating),
				).
				Should(Succeed(), "expected SKR AwsRedisCluster to get status.id")

			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpRedisCluster,
					NewObjActions(
						WithName(awsRedisCluster.Status.Id),
					),
				).
				Should(Succeed())

			By("And has annotaton cloud-manager.kyma-project.io/kymaName")
			Expect(kcpRedisCluster.Annotations[cloudcontrolv1beta1.LabelKymaName]).To(Equal(infra.SkrKymaRef().Name))

			By("And has annotaton cloud-manager.kyma-project.io/remoteName")
			Expect(kcpRedisCluster.Annotations[cloudcontrolv1beta1.LabelRemoteName]).To(Equal(awsRedisCluster.Name))

			By("And has annotaton cloud-manager.kyma-project.io/remoteNamespace")
			Expect(kcpRedisCluster.Annotations[cloudcontrolv1beta1.LabelRemoteNamespace]).To(Equal(awsRedisCluster.Namespace))

			By("And has spec.scope.name equal to SKR Cluster kyma name")
			Expect(kcpRedisCluster.Spec.Scope.Name).To(Equal(infra.SkrKymaRef().Name))

			By("And has spec.remoteRef matching to to SKR IpRange")
			Expect(kcpRedisCluster.Spec.RemoteRef.Namespace).To(Equal(awsRedisCluster.Namespace))
			Expect(kcpRedisCluster.Spec.RemoteRef.Name).To(Equal(awsRedisCluster.Name))

			By("And has spec.instance.aws equal to SKR AwsRedisCluster.spec values")
			Expect(kcpRedisCluster.Spec.Instance.Aws.CacheNodeType).To(Not(Equal("")))
			Expect(kcpRedisCluster.Spec.Instance.Aws.ReplicasPerShard).To(Equal(int32(2)))
			Expect(kcpRedisCluster.Spec.Instance.Aws.ShardCount).To(Equal(int32(10)))
			Expect(kcpRedisCluster.Spec.Instance.Aws.EngineVersion).To(Equal(awsRedisCluster.Spec.EngineVersion))
			Expect(kcpRedisCluster.Spec.Instance.Aws.AutoMinorVersionUpgrade).To(Equal(awsRedisCluster.Spec.AutoMinorVersionUpgrade))
			Expect(kcpRedisCluster.Spec.Instance.Aws.PreferredMaintenanceWindow).To(Equal(awsRedisCluster.Spec.PreferredMaintenanceWindow))
			Expect(kcpRedisCluster.Spec.Instance.Aws.Parameters[parameterKey]).To(Equal(parameterValue))

			By("And has spec.ipRange.name equal to SKR IpRange.status.id")
			Expect(kcpRedisCluster.Spec.IpRange.Name).To(Equal(skrIpRange.Status.Id))
		})

		kcpRedisClusterDiscoveryEndpoint := "192.168.0.1:6576"
		kcpRedisClusterAuthString := "38d6bb99-edf3-43cc-aec6-9ee5d826b0bd"

		By("When KCP RedisCluster has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpRedisCluster,
					WithRedisInstanceDiscoveryEndpoint(kcpRedisClusterDiscoveryEndpoint),
					WithRedisInstanceAuthString(kcpRedisClusterAuthString),

					WithConditions(KcpReadyCondition()),
				).
				Should(Succeed())
		})

		By("Then SKR AwsRedisCluster has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					awsRedisCluster,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
					HavingAwsRedisClusterStatusState(cloudresourcesv1beta1.StateReady),
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
						WithNamespace(awsRedisCluster.Namespace),
					),
				).
				Should(Succeed())

			By("And it has defined cloud-manager default labels")
			Expect(authSecret.Labels[util.WellKnownK8sLabelComponent]).ToNot(BeNil())
			Expect(authSecret.Labels[util.WellKnownK8sLabelPartOf]).ToNot(BeNil())
			Expect(authSecret.Labels[util.WellKnownK8sLabelManagedBy]).ToNot(BeNil())

			By("And it has defined ownmership label")
			Expect(authSecret.Labels[cloudresourcesv1beta1.LabelRedisClusterStatusId]).To(Equal(awsRedisCluster.Status.Id))

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
			WithArguments(infra.Ctx(), infra.SKR().Client(), awsRedisCluster).
			Should(Succeed())
		Eventually(Delete).
			WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange).
			Should(Succeed())
	})

	It("Scenario: SKR AwsRedisCluster is deleted", func() {

		skrIpRangeName := "ee1cbe3e-6f0c-496a-8261-b47c86dacdcf"
		skrIpRange := &cloudresourcesv1beta1.IpRange{}
		skrIpRangeId := "a85fc0dc-cb03-4034-a059-8f1f03edc7d2"

		By("And Given SKR IpRange exists", func() {
			// tell skriprange reconciler to ignore this SKR IpRange
			skriprange.Ignore.AddName(skrIpRangeName)

			Eventually(CreateSkrIpRange).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrIpRange,
					WithName(skrIpRangeName),
				).
				Should(Succeed())
		})
		By("And Given SKR IpRange has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrIpRange,
					WithSkrIpRangeStatusCidr(skrIpRange.Spec.Cidr),
					WithSkrIpRangeStatusId(skrIpRangeId),
					WithConditions(SkrReadyCondition()),
				).
				Should(Succeed())
		})

		awsRedisClusterName := "304418d4-2d0c-4952-b78e-6126c1a3d153"
		awsRedisCluster := &cloudresourcesv1beta1.AwsRedisCluster{}

		By("And Given AwsRedisCluster is created", func() {
			Eventually(CreateAwsRedisCluster).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), awsRedisCluster,
					WithName(awsRedisClusterName),
					WithIpRange(skrIpRange.Name),
					WithAwsRedisClusterDefautSpecs(),
				).
				Should(Succeed())
		})

		kcpRedisCluster := &cloudcontrolv1beta1.RedisCluster{}

		By("And Given KCP RedisCluster is created", func() {
			// load SKR AwsRedisCluster to get ID
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					awsRedisCluster,
					NewObjActions(),
					HavingAwsRedisClusterStatusId(),
					HavingAwsRedisClusterStatusState(cloudresourcesv1beta1.StateCreating),
				).
				Should(Succeed(), "expected SKR AwsRedisCluster to get status.id")

			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpRedisCluster,
					NewObjActions(
						WithName(awsRedisCluster.Status.Id),
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

		By("And Given SKR AwsRedisCluster has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					awsRedisCluster,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
					HavingAwsRedisClusterStatusState(cloudresourcesv1beta1.StateReady),
				).
				Should(Succeed(), "expected AwsRedisCluster to exist and have Ready condition")
		})

		authSecret := &corev1.Secret{}
		By("And Given SKR auth Secret is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					authSecret,
					NewObjActions(
						WithName(awsRedisCluster.Name),
						WithNamespace(awsRedisCluster.Namespace),
					),
				).
				Should(Succeed(), "failed creating auth Secret")
		})

		// DELETE START HERE

		By("When AwsRedisCluster is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsRedisCluster).
				Should(Succeed(), "failed deleting AwsRedisCluster")
		})

		By("Then SKR AwsRedisCluster has Deleting state", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					awsRedisCluster,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.StateDeleting),
					HavingAwsRedisClusterStatusState(cloudresourcesv1beta1.StateDeleting),
				).
				Should(Succeed(), "expected AwsRedisCluster to have Deleting state")
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

		By("Then SKR AwsRedisCluster is deleted", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsRedisCluster).
				Should(Succeed(), "expected AwsRedisCluster not to exist")
		})

		// CleanUp
		Eventually(Delete).
			WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange).
			Should(Succeed())
	})

	It("Scenario: SKR AwsRedisCluster is created with empty IpRange when default IpRange does not exist", func() {

		awsRedisClusterName := "311d3ee3-289c-4d81-afac-b852fc261db7"
		skrIpRangeId := "5565877b-df40-4953-8d08-32938f973430"
		awsRedisCluster := &cloudresourcesv1beta1.AwsRedisCluster{}
		kcpRedisCluster := &cloudcontrolv1beta1.RedisCluster{}
		skrIpRange := &cloudresourcesv1beta1.IpRange{}

		skriprange.Ignore.AddName("default")

		By("Given default SKR IpRange does not exist", func() {
			Consistently(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange,
					NewObjActions(WithName("default"), WithNamespace("kyma-system"))).
				ShouldNot(Succeed())
		})

		By("When AwsRedisCluster is created with empty IpRange", func() {
			Eventually(CreateAwsRedisCluster).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), awsRedisCluster,
					WithName(awsRedisClusterName),
					WithAwsRedisClusterDefautSpecs(),
				).
				Should(Succeed())
		})

		By("Then default SKR IpRange is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange,
					NewObjActions(WithName("default"), WithNamespace("kyma-system"))).
				Should(Succeed())
		})

		By("And Then default SKR IpRange has label app.kubernetes.io/managed-by: cloud-manager", func() {
			Expect(skrIpRange.Labels[util.WellKnownK8sLabelManagedBy]).To(Equal("cloud-manager"))
		})

		By("And Then default SKR IpRange has label app.kubernetes.io/part-of: kyma", func() {
			Expect(skrIpRange.Labels[util.WellKnownK8sLabelPartOf]).To(Equal("kyma"))
		})

		By("And Then AwsRedisCluster is not ready", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsRedisCluster, NewObjActions()).
				Should(Succeed())
			Expect(meta.IsStatusConditionTrue(awsRedisCluster.Status.Conditions, cloudresourcesv1beta1.ConditionTypeReady)).
				To(BeFalse(), "expected AwsRedisCluster not to have Ready condition, but it has")
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

		By("Then KCP RedisCluster is created", func() {
			// load SKR AwsRedisCluster to get ID
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					awsRedisCluster,
					NewObjActions(),
					HavingAwsRedisClusterStatusId(),
					HavingAwsRedisClusterStatusState(cloudresourcesv1beta1.StateCreating),
				).
				Should(Succeed(), "expected SKR AwsRedisCluster to get status.id and status creating")

			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpRedisCluster,
					NewObjActions(
						WithName(awsRedisCluster.Status.Id),
					),
				).
				Should(Succeed())
		})

		By("When KCP RedisCluster has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpRedisCluster,
					WithRedisInstanceDiscoveryEndpoint("192.168.0.1"),
					WithConditions(KcpReadyCondition()),
				).
				Should(Succeed())
		})

		By("Then SKR AwsRedisCluster has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					awsRedisCluster,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
					HavingAwsRedisClusterStatusState(cloudresourcesv1beta1.StateReady),
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
						WithName(awsRedisCluster.Name),
						WithNamespace(awsRedisCluster.Namespace),
					),
				).
				Should(Succeed())
		})

		By("When AwsRedisCluster is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsRedisCluster).
				Should(Succeed())
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsRedisCluster).
				Should(Succeed(), "expected AwsRedisCluster not to exist, but it still does")
		})

		By("Then auth Secret does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), authSecret).
				Should(Succeed(), "expected auth Secret not to exist, but it still does")
		})

		By("And Then KCP RedisCluster does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpRedisCluster).
				Should(Succeed(), "expected KCP RedisCluster not to exist, but it still does")
		})

		By("And Then SKR default IpRange exists", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange, NewObjActions()).
				Should(Succeed())
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

	It("Scenario: SKR AwsRedisCluster is created with empty IpRangeRef when default IpRange already exist", func() {

		awsRedisClusterName := "6ab84eb7-d9ef-44b9-b3af-133b075928e8"
		skrIpRangeId := "b7c4c688-dfd1-4116-b94d-068f4df8c581"
		awsRedisCluster := &cloudresourcesv1beta1.AwsRedisCluster{}
		kcpRedisCluster := &cloudcontrolv1beta1.RedisCluster{}
		skrIpRange := &cloudresourcesv1beta1.IpRange{}

		skriprange.Ignore.AddName("default")

		By("Given default SKR IpRange exists", func() {
			Eventually(CreateSkrIpRange).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange, WithName("default"), WithNamespace("kyma-system")).
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

		By("When AwsRedisCluster is created with empty IpRangeRef", func() {
			Eventually(CreateAwsRedisCluster).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), awsRedisCluster,
					WithName(awsRedisClusterName),
					WithAwsRedisClusterDefautSpecs(),
				).
				Should(Succeed())
		})

		By("Then KCP RedisCluster is created", func() {
			// load SKR AwsRedisCluster to get ID
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					awsRedisCluster,
					NewObjActions(),
					HavingAwsRedisClusterStatusId(),
					HavingAwsRedisClusterStatusState(cloudresourcesv1beta1.StateCreating),
				).
				Should(Succeed(), "expected SKR AwsRedisCluster to get status.id and status creating")

			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpRedisCluster,
					NewObjActions(
						WithName(awsRedisCluster.Status.Id),
					),
				).
				Should(Succeed())
		})

		By("When KCP RedisCluster has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpRedisCluster,
					WithRedisInstanceDiscoveryEndpoint("192.168.0.1"),
					WithConditions(KcpReadyCondition()),
				).
				Should(Succeed())
		})

		By("Then SKR AwsRedisCluster has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					awsRedisCluster,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
					HavingAwsRedisClusterStatusState(cloudresourcesv1beta1.StateReady),
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
						WithName(awsRedisCluster.Name),
						WithNamespace(awsRedisCluster.Namespace),
					),
				).
				Should(Succeed())
		})

		By("When AwsRedisCluster is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsRedisCluster).
				Should(Succeed())
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsRedisCluster).
				Should(Succeed(), "expected AwsRedisCluster not to exist, but it still does")
		})

		By("Then auth Secret does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), authSecret).
				Should(Succeed(), "expected auth Secret not to exist, but it still does")
		})

		By("And Then KCP RedisCluster does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpRedisCluster).
				Should(Succeed(), "expected KCP RedisCluster not to exist, but it still does")
		})

		By("And Then SKR default IpRange exists", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange, NewObjActions()).
				Should(Succeed())
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

	It("Scenario: SKR AwsRedisCluster authSecret is modified", func() {
		awsRedisClusterName := "auth-secret-modified-cluster"
		skrIpRangeId := "7b18f46c-1f38-4fe0-b8ea-c10638d81f20"
		awsRedisCluster := &cloudresourcesv1beta1.AwsRedisCluster{}
		tier := cloudresourcesv1beta1.AwsRedisTierC1
		skrIpRange := &cloudresourcesv1beta1.IpRange{}

		skriprange.Ignore.AddName("default")

		const (
			authSecretName = "aws-cluster-auth-secret-test"
		)
		authSecretLabels := map[string]string{
			"env": "test",
		}
		authSecretAnnotations := map[string]string{
			"purpose": "testing",
		}

		By("Given AwsRedisCluster is created with initial authSecret config", func() {
			Eventually(CreateAwsRedisCluster).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), awsRedisCluster,
					WithName(awsRedisClusterName),
					WithAwsRedisClusterRedisTier(tier),
					WithAwsRedisClusterShardCount(2),
					WithAwsRedisClusterReplicasPerShard(1),
					WithAwsRedisClusterEngineVersion("7.0"),
					WithAwsRedisClusterAuthSecretName(authSecretName),
					WithAwsRedisClusterAuthSecretLabels(authSecretLabels),
					WithAwsRedisClusterAuthSecretAnnotations(authSecretAnnotations),
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
					awsRedisCluster,
					NewObjActions(),
					HavingAwsRedisClusterStatusId(),
					HavingAwsRedisClusterStatusState(cloudresourcesv1beta1.StateCreating),
				).
				Should(Succeed())

			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpRedisCluster,
					NewObjActions(
						WithName(awsRedisCluster.Status.Id),
					),
				).
				Should(Succeed())
		})

		kcpRedisClusterPrimaryEndpoint := "10.0.0.4:6379"
		kcpRedisClusterAuthString := "aws-cluster-auth-string-22222"

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

		By("And Given SKR AwsRedisCluster has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					awsRedisCluster,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
					HavingAwsRedisClusterStatusState(cloudresourcesv1beta1.StateReady),
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
						WithNamespace(awsRedisCluster.Namespace),
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
			"custom-key": "custom-value",
			"endpoint":   "{{.host}}:{{.port}}",
		}

		By("When AwsRedisCluster authSecret config is modified with new labels, annotations, and extraData", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					awsRedisCluster,
					NewObjActions(),
				).
				Should(Succeed())

			awsRedisCluster.Spec.AuthSecret.Labels = newLabels
			awsRedisCluster.Spec.AuthSecret.Annotations = newAnnotations
			awsRedisCluster.Spec.AuthSecret.ExtraData = newExtraData

			Eventually(func() error {
				return infra.SKR().Client().Update(infra.Ctx(), awsRedisCluster)
			}).Should(Succeed())
		})

		By("Then SKR auth Secret is updated with new labels, annotations, and extraData", func() {
			// Wait for controller to reconcile the changes
			Eventually(func() map[string]string {
				_ = infra.SKR().Client().Get(infra.Ctx(), types.NamespacedName{
					Name:      authSecretName,
					Namespace: awsRedisCluster.Namespace,
				}, authSecret)
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
				HaveKeyWithValue("endpoint", []byte(kcpRedisClusterPrimaryEndpoint)),
				HaveKey("host"),
				HaveKey("port"),
				HaveKey("authString"),
			))
		})

		// Cleanup
		Eventually(Delete).
			WithArguments(infra.Ctx(), infra.SKR().Client(), awsRedisCluster).
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
