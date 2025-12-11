package cloudresources

import (
	"fmt"

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
)

var _ = Describe("Feature: SKR GcpRedisInstance", func() {

	It("Scenario: SKR GcpRedisInstance is created with specified IpRange", func() {

		skrIpRangeName := "custom-ip-range"
		skrIpRange := &cloudresourcesv1beta1.IpRange{}
		skrIpRangeId := "acb8e77d-f674-4691-91b2-6f0d5bc81fc6"

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

		gcpRedisInstanceName := "custom-redis-instance"
		gcpRedisInstance := &cloudresourcesv1beta1.GcpRedisInstance{}
		gcpRedisInstanceTier := cloudresourcesv1beta1.GcpRedisTierP2
		gcpRedisInstanceRedisVersion := "REDIS_7_0"
		gcpRedisInstanceAuthEnabled := true
		configKey := "maxmemory-policy"
		configValue := "allkeys-lru"
		gcpRedisInstanceRedisConfigs := map[string]string{
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

		gcpMaintanencePolicy := &cloudresourcesv1beta1.MaintenancePolicy{
			DayOfWeek: &cloudresourcesv1beta1.DayOfWeekPolicy{
				Day: "MONDAY",
				StartTime: cloudresourcesv1beta1.TimeOfDay{
					Hours:   15,
					Minutes: 35,
				},
			},
		}

		extraData := map[string]string{
			"foo":    "bar",
			"parsed": "{{.host}}:{{.port}}",
		}

		By("When GcpRedisInstance is created", func() {
			Eventually(CreateGcpRedisInstance).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), gcpRedisInstance,
					WithName(gcpRedisInstanceName),
					WithIpRange(skrIpRange.Name),
					WithGcpRedisInstanceRedisTier(gcpRedisInstanceTier),
					WithGcpRedisInstanceRedisVersion(gcpRedisInstanceRedisVersion),
					WithGcpRedisInstanceAuthEnabled(gcpRedisInstanceAuthEnabled),
					WithGcpRedisInstanceRedisConfigs(gcpRedisInstanceRedisConfigs),
					WithGcpRedisInstanceMaintenancePolicy(gcpMaintanencePolicy),
					WithGcpRedisInstanceAuthSecretName(authSecretName),
					WithGcpRedisInstanceAuthSecretLabels(authSecretLabels),
					WithGcpRedisInstanceAuthSecretAnnotations(authSecretAnnotations),
					WithGcpRedisInstanceAuthSecretExtraData(extraData),
				).
				Should(Succeed())
		})

		kcpRedisInstance := &cloudcontrolv1beta1.RedisInstance{}

		By("Then KCP RedisInstance is created", func() {
			// load SKR GcpRedisInstance to get ID
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					gcpRedisInstance,
					NewObjActions(),
					HavingGcpRedisInstanceStatusId(),
					HavingGcpRedisInstanceStatusState(cloudresourcesv1beta1.StateCreating),
				).
				Should(Succeed(), "expected SKR GcpRedisInstance to get status.id")

			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpRedisInstance,
					NewObjActions(
						WithName(gcpRedisInstance.Status.Id),
					),
				).
				Should(Succeed())

			By("And has annotaton cloud-manager.kyma-project.io/kymaName")
			Expect(kcpRedisInstance.Annotations[cloudcontrolv1beta1.LabelKymaName]).To(Equal(infra.SkrKymaRef().Name))

			By("And has annotaton cloud-manager.kyma-project.io/remoteName")
			Expect(kcpRedisInstance.Annotations[cloudcontrolv1beta1.LabelRemoteName]).To(Equal(gcpRedisInstance.Name))

			By("And has annotaton cloud-manager.kyma-project.io/remoteNamespace")
			Expect(kcpRedisInstance.Annotations[cloudcontrolv1beta1.LabelRemoteNamespace]).To(Equal(gcpRedisInstance.Namespace))

			By("And has spec.scope.name equal to SKR Cluster kyma name")
			Expect(kcpRedisInstance.Spec.Scope.Name).To(Equal(infra.SkrKymaRef().Name))

			By("And has spec.remoteRef matching to to SKR IpRange")
			Expect(kcpRedisInstance.Spec.RemoteRef.Namespace).To(Equal(gcpRedisInstance.Namespace))
			Expect(kcpRedisInstance.Spec.RemoteRef.Name).To(Equal(gcpRedisInstance.Name))

			By("And has spec.instance.gcp equal to SKR GcpRedisInstance.spec values")

			Expect(kcpRedisInstance.Spec.Instance.Gcp.Tier).To(Not(Equal("")))
			Expect(kcpRedisInstance.Spec.Instance.Gcp.MemorySizeGb).To(Not(Equal(int32(0))))
			Expect(kcpRedisInstance.Spec.Instance.Gcp.ReplicaCount).To(Equal(int32(1)))
			Expect(kcpRedisInstance.Spec.Instance.Gcp.RedisVersion).To(Equal(gcpRedisInstance.Spec.RedisVersion))
			Expect(kcpRedisInstance.Spec.Instance.Gcp.AuthEnabled).To(Equal(gcpRedisInstance.Spec.AuthEnabled))
			Expect(kcpRedisInstance.Spec.Instance.Gcp.RedisConfigs[configKey]).To(Equal(configValue))
			Expect((*kcpRedisInstance.Spec.Instance.Gcp.MaintenancePolicy).DayOfWeek.Day).To(Equal((*gcpRedisInstance.Spec.MaintenancePolicy).DayOfWeek.Day))
			Expect((*kcpRedisInstance.Spec.Instance.Gcp.MaintenancePolicy).DayOfWeek.StartTime.Hours).To(Equal((*gcpRedisInstance.Spec.MaintenancePolicy).DayOfWeek.StartTime.Hours))
			Expect((*kcpRedisInstance.Spec.Instance.Gcp.MaintenancePolicy).DayOfWeek.StartTime.Minutes).To(Equal((*gcpRedisInstance.Spec.MaintenancePolicy).DayOfWeek.StartTime.Minutes))

			By("And has spec.ipRange.name equal to SKR IpRange.status.id")
			Expect(kcpRedisInstance.Spec.IpRange.Name).To(Equal(skrIpRange.Status.Id))
		})

		kcpRedisInstancePrimaryEndpoint := "192.168.0.1:6576"
		kcpRedisInstanceReadEndpoint := "192.168.0.2:6576"
		kcpRedisInstanceAuthString := "a9461793-2449-48d2-8618-0881bbe61d05"

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

		By("Then SKR GcpRedisInstance has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					gcpRedisInstance,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
					HavingGcpRedisInstanceStatusState(cloudresourcesv1beta1.StateReady),
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
						WithNamespace(gcpRedisInstance.Namespace),
					),
				).
				Should(Succeed())

			By("And it has defined cloud-manager default labels")
			Expect(authSecret.Labels[util.WellKnownK8sLabelComponent]).ToNot(BeNil())
			Expect(authSecret.Labels[util.WellKnownK8sLabelPartOf]).ToNot(BeNil())
			Expect(authSecret.Labels[util.WellKnownK8sLabelManagedBy]).ToNot(BeNil())

			By("And it has defined ownmership label")
			Expect(authSecret.Labels[cloudresourcesv1beta1.LabelRedisInstanceStatusId]).To(Equal(gcpRedisInstance.Status.Id))

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
			WithArguments(infra.Ctx(), infra.SKR().Client(), gcpRedisInstance).
			Should(Succeed())
		Eventually(Delete).
			WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange).
			Should(Succeed())
	})

	It("Scenario: SKR GcpRedisInstance is deleted", func() {

		skrIpRangeName := "another-custom-ip-range"
		skrIpRange := &cloudresourcesv1beta1.IpRange{}
		skrIpRangeId := "84631231-903e-47af-82ba-4831c79f65b9"

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

		gcpRedisInstanceName := "another-gcp-redis-instance"
		gcpRedisInstance := &cloudresourcesv1beta1.GcpRedisInstance{}

		By("And Given GcpRedisInstance is created", func() {
			Eventually(CreateGcpRedisInstance).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), gcpRedisInstance,
					WithName(gcpRedisInstanceName),
					WithIpRange(skrIpRange.Name),
					WithGcpRedisInstanceDefaultSpec(),
				).
				Should(Succeed())
		})

		kcpRedisInstance := &cloudcontrolv1beta1.RedisInstance{}

		By("And Given KCP RedisInstance is created", func() {
			// load SKR GcpRedisInstance to get ID
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					gcpRedisInstance,
					NewObjActions(),
					HavingGcpRedisInstanceStatusId(),
					HavingGcpRedisInstanceStatusState(cloudresourcesv1beta1.StateCreating),
				).
				Should(Succeed(), "expected SKR GcpRedisInstance to get status.id")

			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpRedisInstance,
					NewObjActions(
						WithName(gcpRedisInstance.Status.Id),
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

		By("And Given SKR GcpRedisInstance has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					gcpRedisInstance,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
					HavingGcpRedisInstanceStatusState(cloudresourcesv1beta1.StateReady),
				).
				Should(Succeed(), "expected GcpRedisInstance to exist and have Ready condition")
		})

		authSecret := &corev1.Secret{}
		By("And Given SKR auth Secret is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					authSecret,
					NewObjActions(
						WithName(gcpRedisInstance.Name),
						WithNamespace(gcpRedisInstance.Namespace),
					),
				).
				Should(Succeed(), "failed creating auth Secret")
		})

		// DELETE START HERE

		By("When GcpRedisInstance is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), gcpRedisInstance).
				Should(Succeed(), "failed deleting GcpRedisInstance")
		})

		By("Then SKR GcpRedisInstance has Deleting state", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					gcpRedisInstance,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.StateDeleting),
					HavingGcpRedisInstanceStatusState(cloudresourcesv1beta1.StateDeleting),
				).
				Should(Succeed(), "expected GcpRedisInstance to have Deleting state")
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

		By("Then SKR GcpRedisInstance is deleted", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), gcpRedisInstance).
				Should(Succeed(), "expected GcpRedisInstance not to exist")
		})

		// CleanUp
		Eventually(Delete).
			WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange).
			Should(Succeed())
	})

	It("Scenario: SKR GcpRedisInstance is created with empty IpRange when default IpRange does not exist", func() {

		gcpRedisInstanceName := "64b571bd-dbab-40e4-9eeb-5a0eb3b3ed63"
		skrIpRangeId := "209a331b-185f-4413-8d84-e27eaf02ce1d"
		gcpRedisInstance := &cloudresourcesv1beta1.GcpRedisInstance{}
		kcpRedisInstance := &cloudcontrolv1beta1.RedisInstance{}
		skrIpRange := &cloudresourcesv1beta1.IpRange{}

		skriprange.Ignore.AddName("default")

		By("Given default SKR IpRange does not exist", func() {
			Consistently(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange,
					NewObjActions(WithName("default"), WithNamespace("kyma-system"))).
				ShouldNot(Succeed())
		})

		By("When GcpRedisInstance is created with empty IpRange", func() {
			Eventually(CreateGcpRedisInstance).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), gcpRedisInstance,
					WithName(gcpRedisInstanceName),
					WithGcpRedisInstanceDefaultSpec(),
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

		By("And Then GcpRedisInstance is not ready", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), gcpRedisInstance, NewObjActions()).
				Should(Succeed())
			Expect(meta.IsStatusConditionTrue(gcpRedisInstance.Status.Conditions, cloudresourcesv1beta1.ConditionTypeReady)).
				To(BeFalse(), "expected GcpRedisInstance not to have Ready condition, but it has")
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
			// load SKR GcpRedisInstance to get ID
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					gcpRedisInstance,
					NewObjActions(),
					HavingGcpRedisInstanceStatusId(),
					HavingGcpRedisInstanceStatusState(cloudresourcesv1beta1.StateCreating),
				).
				Should(Succeed(), "expected SKR GcpRedisInstance to get status.id and status creating")

			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpRedisInstance,
					NewObjActions(
						WithName(gcpRedisInstance.Status.Id),
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
					WithRedisInstanceAuthString("f85f28f9-0834-41f9-8079-5bfa32b6dadf"),
					WithConditions(KcpReadyCondition()),
				).
				Should(Succeed())
		})

		By("Then SKR GcpRedisInstance has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					gcpRedisInstance,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
					HavingGcpRedisInstanceStatusState(cloudresourcesv1beta1.StateReady),
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
						WithName(gcpRedisInstance.Name),
						WithNamespace(gcpRedisInstance.Namespace),
					),
				).
				Should(Succeed())
		})

		By("When GcpRedisInstance is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), gcpRedisInstance).
				Should(Succeed())
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), gcpRedisInstance).
				Should(Succeed(), "expected GcpRedisInstance not to exist, but it still does")
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

	It("Scenario: SKR GcpRedisInstance is created with empty IpRangeRef when default IpRange already exist", func() {

		gcpRedisInstanceName := "6fc84535-8702-4064-a1d4-92235d9d5dff"
		skrIpRangeId := "343ab759-ed5f-4d0d-93f0-7d4f518bb92e"
		gcpRedisInstance := &cloudresourcesv1beta1.GcpRedisInstance{}
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

		By("When GcpRedisInstance is created with empty IpRangeRef", func() {
			Eventually(CreateGcpRedisInstance).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), gcpRedisInstance,
					WithName(gcpRedisInstanceName),
					WithGcpRedisInstanceDefaultSpec(),
				).
				Should(Succeed())
		})

		By("Then KCP RedisInstance is created", func() {
			// load SKR GcpRedisInstance to get ID
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					gcpRedisInstance,
					NewObjActions(),
					HavingGcpRedisInstanceStatusId(),
					HavingGcpRedisInstanceStatusState(cloudresourcesv1beta1.StateCreating),
				).
				Should(Succeed(), "expected SKR GcpRedisInstance to get status.id and status creating")

			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpRedisInstance,
					NewObjActions(
						WithName(gcpRedisInstance.Status.Id),
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

		By("Then SKR GcpRedisInstance has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					gcpRedisInstance,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
					HavingGcpRedisInstanceStatusState(cloudresourcesv1beta1.StateReady),
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
						WithName(gcpRedisInstance.Name),
						WithNamespace(gcpRedisInstance.Namespace),
					),
				).
				Should(Succeed())
		})

		By("When GcpRedisInstance is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), gcpRedisInstance).
				Should(Succeed())
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), gcpRedisInstance).
				Should(Succeed(), "expected GcpRedisInstance not to exist, but it still does")
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

	It("Scenario: SKR GcpRedisInstance authSecret is modified", func() {

		gcpRedisInstanceName := "auth-secret-modified-redis"
		skrIpRangeId := "5c70629f-a13f-4b04-af47-1ab274c1c7ag"
		gcpRedisInstance := &cloudresourcesv1beta1.GcpRedisInstance{}
		redisVersion := "REDIS_7_0"
		tier := cloudresourcesv1beta1.GcpRedisTierP1
		skrIpRange := &cloudresourcesv1beta1.IpRange{}

		skriprange.Ignore.AddName("default")

		const (
			authSecretName = "gcp-auth-secret-test"
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

		By("And Given GcpRedisInstance is created with initial authSecret config", func() {
			Eventually(CreateGcpRedisInstance).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), gcpRedisInstance,
					WithName(gcpRedisInstanceName),
					WithGcpRedisInstanceRedisVersion(redisVersion),
					WithGcpRedisInstanceRedisTier(tier),
					WithGcpRedisInstanceAuthSecretName(authSecretName),
					WithGcpRedisInstanceAuthSecretLabels(authSecretLabels),
					WithGcpRedisInstanceAuthSecretAnnotations(authSecretAnnotations),
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
					gcpRedisInstance,
					NewObjActions(),
					HavingGcpRedisInstanceStatusId(),
					HavingGcpRedisInstanceStatusState(cloudresourcesv1beta1.StateCreating),
				).
				Should(Succeed())

			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpRedisInstance,
					NewObjActions(
						WithName(gcpRedisInstance.Status.Id),
					),
				).
				Should(Succeed())
		})

		kcpRedisInstancePrimaryEndpoint := "10.0.0.1:6379"
		kcpRedisInstanceAuthString := "initial-auth-string-12345"

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

		By("And Given SKR GcpRedisInstance has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					gcpRedisInstance,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
					HavingGcpRedisInstanceStatusState(cloudresourcesv1beta1.StateReady),
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
						WithNamespace(gcpRedisInstance.Namespace),
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

		By("When GcpRedisInstance authSecret config is modified with new labels, annotations, and extraData", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					gcpRedisInstance,
					NewObjActions(),
				).
				Should(Succeed())

			gcpRedisInstance.Spec.AuthSecret.Labels = newLabels
			gcpRedisInstance.Spec.AuthSecret.Annotations = newAnnotations
			gcpRedisInstance.Spec.AuthSecret.ExtraData = newExtraData

			Eventually(Update).
				WithArguments(infra.Ctx(), infra.SKR().Client(), gcpRedisInstance).
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
						WithNamespace(gcpRedisInstance.Namespace),
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
			Expect(authSecret.Labels).To(HaveKey(cloudresourcesv1beta1.LabelRedisInstanceStatusId))
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

		// CleanUp
		Eventually(Delete).
			WithArguments(infra.Ctx(), infra.SKR().Client(), gcpRedisInstance).
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
