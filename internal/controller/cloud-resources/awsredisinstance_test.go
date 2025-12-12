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

var _ = Describe("Feature: SKR AwsRedisInstance", func() {

	It("Scenario: SKR AwsRedisInstance is created with specified IpRange", func() {

		skrIpRangeName := "8886105f-ce02-4384-959e-afc7bb0dc700"
		skrIpRange := &cloudresourcesv1beta1.IpRange{}
		skrIpRangeId := "2b978a2a-df7c-4811-819f-97396175cd28"

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

		awsRedisInstanceName := "897253b7-5ed1-4bbd-9782-99a2e07aea94"
		awsRedisInstance := &cloudresourcesv1beta1.AwsRedisInstance{}

		const (
			authSecretName = "26bc6c7b-190a-489a-83d2-afe272cbdffb"
		)
		authSecretLabels := map[string]string{
			"foo": "1",
		}
		authSecretAnnotations := map[string]string{
			"bar": "2",
		}

		redisTier := cloudresourcesv1beta1.AwsRedisTierP1
		engineVersion := "6.x"
		autoMinorVersionUpgrade := true
		authEnabled := true

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

		By("When AwsRedisInstance is created", func() {
			Eventually(CreateAwsRedisInstance).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), awsRedisInstance,
					WithName(awsRedisInstanceName),
					WithIpRange(skrIpRange.Name),
					WithAwsRedisInstanceAuthSecretName(authSecretName),
					WithAwsRedisInstanceAuthSecretLabels(authSecretLabels),
					WithAwsRedisInstanceAuthSecretAnnotations(authSecretAnnotations),
					WithAwsRedisInstanceAuthSecretExtraData(extraData),
					WithAwsRedisInstanceRedisTier(redisTier),
					WithAwsRedisInstanceEngineVersion(engineVersion),
					WithAwsRedisInstanceAutoMinorVersionUpgrade(autoMinorVersionUpgrade),
					WithAwsRedisInstanceAuthEnabled(authEnabled),
					WithAwsRedisInstancePreferredMaintenanceWindow(preferredMaintenanceWindow),
					WithAwsRedisInstanceParameters(parameters),
				).
				Should(Succeed())
		})

		kcpRedisInstance := &cloudcontrolv1beta1.RedisInstance{}

		By("Then KCP RedisInstance is created", func() {
			// load SKR AwsRedisInstance to get ID
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					awsRedisInstance,
					NewObjActions(),
					HavingAwsRedisInstanceStatusId(),
					HavingAwsRedisInstanceStatusState(cloudresourcesv1beta1.StateCreating),
				).
				Should(Succeed(), "expected SKR AwsRedisInstance to get status.id")

			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpRedisInstance,
					NewObjActions(
						WithName(awsRedisInstance.Status.Id),
					),
				).
				Should(Succeed())

			By("And has annotaton cloud-manager.kyma-project.io/kymaName")
			Expect(kcpRedisInstance.Annotations[cloudcontrolv1beta1.LabelKymaName]).To(Equal(infra.SkrKymaRef().Name))

			By("And has annotaton cloud-manager.kyma-project.io/remoteName")
			Expect(kcpRedisInstance.Annotations[cloudcontrolv1beta1.LabelRemoteName]).To(Equal(awsRedisInstance.Name))

			By("And has annotaton cloud-manager.kyma-project.io/remoteNamespace")
			Expect(kcpRedisInstance.Annotations[cloudcontrolv1beta1.LabelRemoteNamespace]).To(Equal(awsRedisInstance.Namespace))

			By("And has spec.scope.name equal to SKR Cluster kyma name")
			Expect(kcpRedisInstance.Spec.Scope.Name).To(Equal(infra.SkrKymaRef().Name))

			By("And has spec.remoteRef matching to to SKR IpRange")
			Expect(kcpRedisInstance.Spec.RemoteRef.Namespace).To(Equal(awsRedisInstance.Namespace))
			Expect(kcpRedisInstance.Spec.RemoteRef.Name).To(Equal(awsRedisInstance.Name))

			By("And has spec.instance.aws equal to SKR AwsRedisInstance.spec values")
			Expect(kcpRedisInstance.Spec.Instance.Aws.CacheNodeType).To(Not(Equal("")))
			Expect(kcpRedisInstance.Spec.Instance.Aws.ReadReplicas).To(Equal(int32(1)))
			Expect(kcpRedisInstance.Spec.Instance.Aws.EngineVersion).To(Equal(awsRedisInstance.Spec.EngineVersion))
			Expect(kcpRedisInstance.Spec.Instance.Aws.AutoMinorVersionUpgrade).To(Equal(awsRedisInstance.Spec.AutoMinorVersionUpgrade))
			Expect(kcpRedisInstance.Spec.Instance.Aws.PreferredMaintenanceWindow).To(Equal(awsRedisInstance.Spec.PreferredMaintenanceWindow))
			Expect(kcpRedisInstance.Spec.Instance.Aws.Parameters[parameterKey]).To(Equal(parameterValue))

			By("And has spec.ipRange.name equal to SKR IpRange.status.id")
			Expect(kcpRedisInstance.Spec.IpRange.Name).To(Equal(skrIpRange.Status.Id))
		})

		kcpRedisInstancePrimaryEndpoint := "192.168.0.1:6576"
		kcpRedisInstanceReadEndpoint := "192.168.0.2:6576"
		kcpRedisInstanceAuthString := "cdaa7502-3433-441e-802d-310d931848bf"

		By("When KCP RedisInstance has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpRedisInstance,
					WithRedisInstancePrimaryEndpoint(kcpRedisInstancePrimaryEndpoint),
					WithRedisInstanceReadEndpoint(kcpRedisInstanceReadEndpoint),
					WithRedisInstanceAuthString(kcpRedisInstanceAuthString),

					WithConditions(KcpReadyCondition()),
				).
				Should(Succeed())
		})

		By("Then SKR AwsRedisInstance has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					awsRedisInstance,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
					HavingAwsRedisInstanceStatusState(cloudresourcesv1beta1.StateReady),
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
						WithNamespace(awsRedisInstance.Namespace),
					),
				).
				Should(Succeed())

			By("And it has defined cloud-manager default labels")
			Expect(authSecret.Labels[util.WellKnownK8sLabelComponent]).ToNot(BeNil())
			Expect(authSecret.Labels[util.WellKnownK8sLabelPartOf]).ToNot(BeNil())
			Expect(authSecret.Labels[util.WellKnownK8sLabelManagedBy]).ToNot(BeNil())

			By("And it has defined ownmership label")
			Expect(authSecret.Labels[cloudresourcesv1beta1.LabelRedisInstanceStatusId]).To(Equal(awsRedisInstance.Status.Id))

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
			Expect(authSecret.Data).To(HaveKeyWithValue("parsed", []byte(kcpRedisInstancePrimaryEndpoint)), "expected auth secret data to have parsed=host:port")

			By("And it has defined cloud-manager finalizer")
			Expect(authSecret.Finalizers).To(ContainElement(api.CommonFinalizerDeletionHook))
		})

		// CleanUp
		Eventually(Delete).
			WithArguments(infra.Ctx(), infra.SKR().Client(), awsRedisInstance).
			Should(Succeed())
		Eventually(Delete).
			WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange).
			Should(Succeed())
	})

	It("Scenario: SKR AwsRedisInstance is deleted", func() {

		skrIpRangeName := "09fcc653-500c-478c-84da-6cea9948e8af"
		skrIpRange := &cloudresourcesv1beta1.IpRange{}
		skrIpRangeId := "33bc0194-d9de-4ac4-a582-10a6ac26f850"

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

		awsRedisInstanceName := "58137995-df7c-4612-80ef-fde1bac32755"
		awsRedisInstance := &cloudresourcesv1beta1.AwsRedisInstance{}

		By("And Given AwsRedisInstance is created", func() {
			Eventually(CreateAwsRedisInstance).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), awsRedisInstance,
					WithName(awsRedisInstanceName),
					WithIpRange(skrIpRange.Name),
					WithAwsRedisInstanceDefautSpecs(),
				).
				Should(Succeed())
		})

		kcpRedisInstance := &cloudcontrolv1beta1.RedisInstance{}

		By("And Given KCP RedisInstance is created", func() {
			// load SKR AwsRedisInstance to get ID
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					awsRedisInstance,
					NewObjActions(),
					HavingAwsRedisInstanceStatusId(),
					HavingAwsRedisInstanceStatusState(cloudresourcesv1beta1.StateCreating),
				).
				Should(Succeed(), "expected SKR AwsRedisInstance to get status.id")

			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpRedisInstance,
					NewObjActions(
						WithName(awsRedisInstance.Status.Id),
					),
				).
				Should(Succeed(), "expected KCP RedisInstance to be created, but it was not")

			Eventually(Update).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpRedisInstance, AddFinalizer(api.CommonFinalizerDeletionHook)).
				Should(Succeed(), "failed adding finalizer on KCP RedisInstance")
		})

		By("And Given KCP RedisInstance has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpRedisInstance,
					WithConditions(KcpReadyCondition()),
				).
				Should(Succeed(), "failed setting KCP RedisInstance Ready condition")
		})

		By("And Given SKR AwsRedisInstance has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					awsRedisInstance,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
					HavingAwsRedisInstanceStatusState(cloudresourcesv1beta1.StateReady),
				).
				Should(Succeed(), "expected AwsRedisInstance to exist and have Ready condition")
		})

		authSecret := &corev1.Secret{}
		By("And Given SKR auth Secret is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					authSecret,
					NewObjActions(
						WithName(awsRedisInstance.Name),
						WithNamespace(awsRedisInstance.Namespace),
					),
				).
				Should(Succeed(), "failed creating auth Secret")
		})

		// DELETE START HERE

		By("When AwsRedisInstance is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsRedisInstance).
				Should(Succeed(), "failed deleting AwsRedisInstance")
		})

		By("Then SKR AwsRedisInstance has Deleting state", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					awsRedisInstance,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.StateDeleting),
					HavingAwsRedisInstanceStatusState(cloudresourcesv1beta1.StateDeleting),
				).
				Should(Succeed(), "expected AwsRedisInstance to have Deleting state")
		})

		By("And Then SKR auth Secret is deleted", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), authSecret).
				Should(Succeed(), "expected authSecret not to exist")
		})

		By("And Then KCP RedisInstance is marked for deletion", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpRedisInstance, NewObjActions(), HavingDeletionTimestamp()).
				Should(Succeed(), "expected KCP RedisInstance to be marked for deletion")
		})

		By("When KCP RedisInstance finalizer is removed and it is deleted", func() {
			Eventually(Update).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpRedisInstance, RemoveFinalizer(api.CommonFinalizerDeletionHook)).
				Should(Succeed(), "failed removing finalizer on KCP RedisInstance")
		})

		By("Then SKR AwsRedisInstance is deleted", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsRedisInstance).
				Should(Succeed(), "expected AwsRedisInstance not to exist")
		})

		// CleanUp
		Eventually(Delete).
			WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange).
			Should(Succeed())
	})

	It("Scenario: SKR AwsRedisInstance is created with empty IpRange when default IpRange does not exist", func() {

		awsRedisInstanceName := "b5351da0-5f49-4612-b9cd-e9a8357c0ea2"
		skrIpRangeId := "5c70629f-a13f-4b04-af47-1ab274c1c7cd"
		awsRedisInstance := &cloudresourcesv1beta1.AwsRedisInstance{}
		kcpRedisInstance := &cloudcontrolv1beta1.RedisInstance{}
		skrIpRange := &cloudresourcesv1beta1.IpRange{}

		skriprange.Ignore.AddName("default")

		By("Given default SKR IpRange does not exist", func() {
			Consistently(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange,
					NewObjActions(WithName("default"), WithNamespace("kyma-system"))).
				ShouldNot(Succeed())
		})

		By("When AwsRedisInstance is created with empty IpRange", func() {
			Eventually(CreateAwsRedisInstance).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), awsRedisInstance,
					WithName(awsRedisInstanceName),
					WithAwsRedisInstanceDefautSpecs(),
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

		By("And Then AwsRedisInstance is not ready", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsRedisInstance, NewObjActions()).
				Should(Succeed())
			Expect(meta.IsStatusConditionTrue(awsRedisInstance.Status.Conditions, cloudresourcesv1beta1.ConditionTypeReady)).
				To(BeFalse(), "expected AwsRedisInstance not to have Ready condition, but it has")
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

		By("Then KCP RedisInstance is created", func() {
			// load SKR AwsRedisInstance to get ID
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					awsRedisInstance,
					NewObjActions(),
					HavingAwsRedisInstanceStatusId(),
					HavingAwsRedisInstanceStatusState(cloudresourcesv1beta1.StateCreating),
				).
				Should(Succeed(), "expected SKR AwsRedisInstance to get status.id and status creating")

			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpRedisInstance,
					NewObjActions(
						WithName(awsRedisInstance.Status.Id),
					),
				).
				Should(Succeed())
		})

		By("When KCP RedisInstance has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpRedisInstance,
					WithRedisInstancePrimaryEndpoint("192.168.0.1"),
					WithRedisInstanceReadEndpoint("192.168.2.2"),
					WithRedisInstanceAuthString("9d9c7159-39be-4992-90a2-95e81cf6298a"),
					WithConditions(KcpReadyCondition()),
				).
				Should(Succeed())
		})

		By("Then SKR AwsRedisInstance has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					awsRedisInstance,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
					HavingAwsRedisInstanceStatusState(cloudresourcesv1beta1.StateReady),
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
						WithName(awsRedisInstance.Name),
						WithNamespace(awsRedisInstance.Namespace),
					),
				).
				Should(Succeed())
		})

		By("When AwsRedisInstance is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsRedisInstance).
				Should(Succeed())
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsRedisInstance).
				Should(Succeed(), "expected AwsRedisInstance not to exist, but it still does")
		})

		By("Then auth Secret does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), authSecret).
				Should(Succeed(), "expected auth Secret not to exist, but it still does")
		})

		By("And Then KCP RedisInstance does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpRedisInstance).
				Should(Succeed(), "expected KCP RedisInstance not to exist, but it still does")
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

	It("Scenario: SKR AwsRedisInstance is created with empty IpRangeRef when default IpRange already exist", func() {

		awsRedisInstanceName := "7f86e5fc-8b2b-44c5-8275-967e6e2403a4"
		skrIpRangeId := "7f09262c-41fe-43be-91eb-10aa3e273d7b"
		awsRedisInstance := &cloudresourcesv1beta1.AwsRedisInstance{}
		kcpRedisInstance := &cloudcontrolv1beta1.RedisInstance{}
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

		By("When AwsRedisInstance is created with empty IpRangeRef", func() {
			Eventually(CreateAwsRedisInstance).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), awsRedisInstance,
					WithName(awsRedisInstanceName),
					WithAwsRedisInstanceDefautSpecs(),
				).
				Should(Succeed())
		})

		By("Then KCP RedisInstance is created", func() {
			// load SKR AwsRedisInstance to get ID
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					awsRedisInstance,
					NewObjActions(),
					HavingAwsRedisInstanceStatusId(),
					HavingAwsRedisInstanceStatusState(cloudresourcesv1beta1.StateCreating),
				).
				Should(Succeed(), "expected SKR AwsRedisInstance to get status.id and status creating")

			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpRedisInstance,
					NewObjActions(
						WithName(awsRedisInstance.Status.Id),
					),
				).
				Should(Succeed())
		})

		By("When KCP RedisInstance has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpRedisInstance,
					WithRedisInstancePrimaryEndpoint("192.168.0.1"),
					WithConditions(KcpReadyCondition()),
				).
				Should(Succeed())
		})

		By("Then SKR AwsRedisInstance has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					awsRedisInstance,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
					HavingAwsRedisInstanceStatusState(cloudresourcesv1beta1.StateReady),
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
						WithName(awsRedisInstance.Name),
						WithNamespace(awsRedisInstance.Namespace),
					),
				).
				Should(Succeed())
		})

		By("When AwsRedisInstance is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsRedisInstance).
				Should(Succeed())
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsRedisInstance).
				Should(Succeed(), "expected AwsRedisInstance not to exist, but it still does")
		})

		By("Then auth Secret does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), authSecret).
				Should(Succeed(), "expected auth Secret not to exist, but it still does")
		})

		By("And Then KCP RedisInstance does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpRedisInstance).
				Should(Succeed(), "expected KCP RedisInstance not to exist, but it still does")
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

	It("Scenario: SKR AwsRedisInstance authSecret is modified", func() {
		awsRedisInstanceName := "auth-secret-modified-instance"
		skrIpRangeId := "6a07d35a-0e27-45e9-b7da-b99527cb0f09"
		awsRedisInstance := &cloudresourcesv1beta1.AwsRedisInstance{}
		tier := cloudresourcesv1beta1.AwsRedisTierP1
		skrIpRange := &cloudresourcesv1beta1.IpRange{}

		skriprange.Ignore.AddName("default")

		const (
			authSecretName = "aws-instance-auth-secret-test"
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

		By("And Given AwsRedisInstance is created with initial authSecret config", func() {
			Eventually(CreateAwsRedisInstance).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), awsRedisInstance,
					WithName(awsRedisInstanceName),
					WithAwsRedisInstanceRedisTier(tier),
					WithAwsRedisInstanceEngineVersion("7.1"),
					WithAwsRedisInstanceAuthSecretName(authSecretName),
					WithAwsRedisInstanceAuthSecretLabels(authSecretLabels),
					WithAwsRedisInstanceAuthSecretAnnotations(authSecretAnnotations),
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

		kcpRedisInstance := &cloudcontrolv1beta1.RedisInstance{}

		By("And Given KCP RedisInstance is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					awsRedisInstance,
					NewObjActions(),
					HavingAwsRedisInstanceStatusId(),
					HavingAwsRedisInstanceStatusState(cloudresourcesv1beta1.StateCreating),
				).
				Should(Succeed())

			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpRedisInstance,
					NewObjActions(
						WithName(awsRedisInstance.Status.Id),
					),
				).
				Should(Succeed())
		})

		kcpRedisInstancePrimaryEndpoint := "10.0.0.3:6379"
		kcpRedisInstanceAuthString := "aws-instance-auth-string-11111"

		By("And Given KCP RedisInstance has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpRedisInstance,
					WithRedisInstancePrimaryEndpoint(kcpRedisInstancePrimaryEndpoint),
					WithRedisInstanceAuthString(kcpRedisInstanceAuthString),
					WithConditions(KcpReadyCondition()),
				).
				Should(Succeed())
		})

		By("And Given SKR AwsRedisInstance has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					awsRedisInstance,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
					HavingAwsRedisInstanceStatusState(cloudresourcesv1beta1.StateReady),
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
						WithNamespace(awsRedisInstance.Namespace),
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

		By("When AwsRedisInstance authSecret config is modified with new labels, annotations, and extraData", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					awsRedisInstance,
					NewObjActions(),
				).
				Should(Succeed())

			awsRedisInstance.Spec.AuthSecret.Labels = newLabels
			awsRedisInstance.Spec.AuthSecret.Annotations = newAnnotations
			awsRedisInstance.Spec.AuthSecret.ExtraData = newExtraData

			Eventually(func() error {
				return infra.SKR().Client().Update(infra.Ctx(), awsRedisInstance)
			}).Should(Succeed())
		})

		By("Then SKR auth Secret is updated with new labels, annotations, and extraData", func() {
			// Wait for controller to reconcile the changes
			Eventually(func() map[string]string {
				_ = infra.SKR().Client().Get(infra.Ctx(), types.NamespacedName{
					Name:      authSecretName,
					Namespace: awsRedisInstance.Namespace,
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
				HaveKeyWithValue("endpoint", []byte(kcpRedisInstancePrimaryEndpoint)),
				HaveKey("host"),
				HaveKey("port"),
				HaveKey("authString"),
			))
		})

		// Cleanup
		Eventually(Delete).
			WithArguments(infra.Ctx(), infra.SKR().Client(), awsRedisInstance).
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
