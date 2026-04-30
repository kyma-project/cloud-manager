package cloudresources

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kyma-project/cloud-manager/api"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	skrsapnfsvolume "github.com/kyma-project/cloud-manager/pkg/skr/sapnfsvolume"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	"github.com/kyma-project/cloud-manager/pkg/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var _ = Describe("Feature: SKR SapNfsVolumeSnapshot", func() {

	It("Scenario: SKR SapNfsVolumeSnapshot is created and deleted", func() {
		sapNfsVolumeName := "d5720673-2774-4f85-8b3d-6c2913567721"
		sapNfsVolumeId := uuid.NewString()
		snapshotName := uuid.NewString()
		shareId := uuid.NewString()
		sapNfsVolume := &cloudresourcesv1beta1.SapNfsVolume{}
		kcpNfsInstance := &cloudcontrolv1beta1.NfsInstance{}
		snapshot := &cloudresourcesv1beta1.SapNfsVolumeSnapshot{}
		scope := &cloudcontrolv1beta1.Scope{}

		sapMock := infra.SapMock().NewProject()

		kymaName := "f1be1f6f-10fb-4254-a1b2-34f4f71f182e"
		skrKymaRef := util.Must(infra.ScopeProvider().GetScope(infra.Ctx(), types.NamespacedName{Name: kymaName}))

		skrsapnfsvolume.Ignore.AddName(sapNfsVolumeName)
		defer skrsapnfsvolume.Ignore.RemoveName(sapNfsVolumeName)

		By("Given KCP Scope exists", func() {
			Eventually(func() error {
				return GivenScopeOpenStackExists(infra.Ctx(), infra, scope, sapMock.ProviderParams(), WithName(skrKymaRef.Name))
			}).Should(Succeed())
		})

		By("And Given SapNfsVolume exists in Ready state", func() {
			Eventually(CreateSapNfsVolume).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), sapNfsVolume,
					WithName(sapNfsVolumeName),
					WithSapNfsVolumeCapacity(100),
				).Should(Succeed())

			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), sapNfsVolume,
					WithConditions(SkrReadyCondition()),
					WithSapNfsVolumeStatusId(sapNfsVolumeId),
				).Should(Succeed())
		})

		By("And Given KCP NfsInstance exists with shareId", func() {
			kcpNfsInstance.Name = sapNfsVolumeId
			kcpNfsInstance.Namespace = infra.KCP().Namespace()
			kcpNfsInstance.Spec = cloudcontrolv1beta1.NfsInstanceSpec{
				RemoteRef: cloudcontrolv1beta1.RemoteRef{
					Namespace: DefaultSkrNamespace,
					Name:      sapNfsVolumeName,
				},
				IpRange: cloudcontrolv1beta1.IpRangeRef{Name: "default"},
				Scope:   cloudcontrolv1beta1.ScopeRef{Name: skrKymaRef.Name},
				Instance: cloudcontrolv1beta1.NfsInstanceInfo{
					OpenStack: &cloudcontrolv1beta1.NfsInstanceOpenStack{SizeGb: 100},
				},
			}
			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpNfsInstance).
				Should(Succeed())

			kcpNfsInstance.SetStateData("shareId", shareId)
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.KCP().Client(), kcpNfsInstance,
					WithConditions(KcpReadyCondition()),
				).Should(Succeed())
		})

		By("When SapNfsVolumeSnapshot is created", func() {
			snapshot.Name = snapshotName
			snapshot.Namespace = DefaultSkrNamespace
			snapshot.Spec.SourceVolume = corev1.ObjectReference{
				Name: sapNfsVolumeName,
			}
			Eventually(func() error {
				return infra.SKR().Client().Create(infra.Ctx(), snapshot)
			}).Should(Succeed())
		})

		By("Then SapNfsVolumeSnapshot has status.id set", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), snapshot,
					NewObjActions(),
					HavingFieldSet("status", "id"),
				).Should(Succeed(), "expected snapshot to get status.id")
		})

		By("And Then SapNfsVolumeSnapshot has status.openstackId set", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), snapshot,
					NewObjActions(),
					HavingFieldSet("status", "openstackId"),
				).Should(Succeed(), "expected snapshot to get status.openstackId")
		})

		By("When Manila snapshot becomes available", func() {
			Eventually(func() error {
				err := infra.SKR().Client().Get(infra.Ctx(), client.ObjectKeyFromObject(snapshot), snapshot)
				if err != nil {
					return err
				}
				if snapshot.Status.OpenstackId == "" {
					return fmt.Errorf("openstackId not yet set")
				}
				sapMock.SetSnapshotStatus(snapshot.Status.OpenstackId, "available")
				return nil
			}).Should(Succeed())
		})

		By("Then SapNfsVolumeSnapshot has Ready state", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), snapshot,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
					HavingFieldValue(cloudresourcesv1beta1.StateReady, "status", "state"),
				).Should(Succeed(), "expected snapshot to be Ready")
		})

		By("And Then SapNfsVolumeSnapshot has finalizer", func() {
			Expect(controllerutil.ContainsFinalizer(snapshot, api.CommonFinalizerDeletionHook)).To(BeTrue())
		})

		By("And Then SapNfsVolumeSnapshot has shareId in status", func() {
			Expect(snapshot.Status.ShareId).To(Equal(shareId))
		})

		// DELETE

		By("When SapNfsVolumeSnapshot is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), snapshot).
				Should(Succeed())
		})

		By("Then SapNfsVolumeSnapshot is removed", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), snapshot).
				Should(Succeed(), "expected SapNfsVolumeSnapshot to be deleted")
		})
	})

	It("Scenario: SKR SapNfsVolumeSnapshot TTL expiry triggers deletion", func() {
		sapNfsVolumeName := "48debda9-2c59-4f36-a12c-26604d18b9b9"
		sapNfsVolumeId := uuid.NewString()
		snapshotName := uuid.NewString()
		shareId := uuid.NewString()
		sapNfsVolume := &cloudresourcesv1beta1.SapNfsVolume{}
		kcpNfsInstance := &cloudcontrolv1beta1.NfsInstance{}
		snapshot := &cloudresourcesv1beta1.SapNfsVolumeSnapshot{}
		scope := &cloudcontrolv1beta1.Scope{}

		sapMock := infra.SapMock().NewProject()

		kymaName := "1486b581-3b15-4070-ad55-6504b2cdeeee"
		skrKymaRef := util.Must(infra.ScopeProvider().GetScope(infra.Ctx(), types.NamespacedName{Name: kymaName}))

		skrsapnfsvolume.Ignore.AddName(sapNfsVolumeName)
		defer skrsapnfsvolume.Ignore.RemoveName(sapNfsVolumeName)

		By("Given KCP Scope exists", func() {
			Eventually(func() error {
				return GivenScopeOpenStackExists(infra.Ctx(), infra, scope, sapMock.ProviderParams(), WithName(skrKymaRef.Name))
			}).Should(Succeed())
		})

		By("And Given SapNfsVolume exists in Ready state", func() {
			Eventually(CreateSapNfsVolume).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), sapNfsVolume,
					WithName(sapNfsVolumeName),
					WithSapNfsVolumeCapacity(100),
				).Should(Succeed())

			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), sapNfsVolume,
					WithConditions(SkrReadyCondition()),
					WithSapNfsVolumeStatusId(sapNfsVolumeId),
				).Should(Succeed())
		})

		By("And Given KCP NfsInstance exists with shareId", func() {
			kcpNfsInstance.Name = sapNfsVolumeId
			kcpNfsInstance.Namespace = infra.KCP().Namespace()
			kcpNfsInstance.Spec = cloudcontrolv1beta1.NfsInstanceSpec{
				RemoteRef: cloudcontrolv1beta1.RemoteRef{
					Namespace: DefaultSkrNamespace,
					Name:      sapNfsVolumeName,
				},
				IpRange: cloudcontrolv1beta1.IpRangeRef{Name: "default"},
				Scope:   cloudcontrolv1beta1.ScopeRef{Name: skrKymaRef.Name},
				Instance: cloudcontrolv1beta1.NfsInstanceInfo{
					OpenStack: &cloudcontrolv1beta1.NfsInstanceOpenStack{SizeGb: 100},
				},
			}
			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpNfsInstance).
				Should(Succeed())

			kcpNfsInstance.SetStateData("shareId", shareId)
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.KCP().Client(), kcpNfsInstance,
					WithConditions(KcpReadyCondition()),
				).Should(Succeed())
		})

		By("When SapNfsVolumeSnapshot is created with deleteAfterDays=1", func() {
			snapshot.Name = snapshotName
			snapshot.Namespace = DefaultSkrNamespace
			snapshot.Spec.SourceVolume = corev1.ObjectReference{
				Name: sapNfsVolumeName,
			}
			snapshot.Spec.DeleteAfterDays = 1

			Eventually(func() error {
				return infra.SKR().Client().Create(infra.Ctx(), snapshot)
			}).Should(Succeed())
		})

		By("And When Manila snapshot becomes available", func() {
			Eventually(func() error {
				err := infra.SKR().Client().Get(infra.Ctx(), client.ObjectKeyFromObject(snapshot), snapshot)
				if err != nil {
					return err
				}
				if snapshot.Status.OpenstackId == "" {
					return fmt.Errorf("openstackId not yet set")
				}
				sapMock.SetSnapshotStatus(snapshot.Status.OpenstackId, "available")
				return nil
			}).Should(Succeed())
		})

		By("And When SapNfsVolumeSnapshot reaches Ready state", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), snapshot,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
					HavingFieldValue(cloudresourcesv1beta1.StateReady, "status", "state"),
				).Should(Succeed())
		})

		By("And When the fake clock is advanced past the TTL", func() {
			testFakeClock.Step(48 * time.Hour) // advance 2 days past the 1-day TTL
		})

		By("Then SapNfsVolumeSnapshot is deleted by TTL expiry", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), snapshot).
				Should(Succeed(), "expected snapshot to be deleted by TTL expiry")
		})
	})

	It("Scenario: SKR SapNfsVolumeSnapshot reports error when source volume is not found", func() {
		snapshotName := uuid.NewString()
		nonExistentVolumeName := uuid.NewString()
		snapshot := &cloudresourcesv1beta1.SapNfsVolumeSnapshot{}
		scope := &cloudcontrolv1beta1.Scope{}

		sapMock := infra.SapMock().NewProject()

		kymaName := "109ba5c3-31d5-4734-949e-da09177f90c6"
		skrKymaRef := util.Must(infra.ScopeProvider().GetScope(infra.Ctx(), types.NamespacedName{Name: kymaName}))

		By("Given KCP Scope exists", func() {
			Eventually(func() error {
				return GivenScopeOpenStackExists(infra.Ctx(), infra, scope, sapMock.ProviderParams(), WithName(skrKymaRef.Name))
			}).Should(Succeed())
		})

		By("When SapNfsVolumeSnapshot is created referencing a non-existent volume", func() {
			snapshot.Name = snapshotName
			snapshot.Namespace = DefaultSkrNamespace
			snapshot.Spec.SourceVolume = corev1.ObjectReference{
				Name: nonExistentVolumeName,
			}
			Eventually(func() error {
				return infra.SKR().Client().Create(infra.Ctx(), snapshot)
			}).Should(Succeed())
		})

		By("Then SapNfsVolumeSnapshot has Error state", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), snapshot,
					NewObjActions(),
					HavingFieldValue(cloudresourcesv1beta1.StateError, "status", "state"),
				).Should(Succeed(), "expected snapshot to have Error state")
		})

		By("And Then SapNfsVolumeSnapshot has error condition about missing volume", func() {
			cond := findCondition(snapshot.Status.Conditions, cloudresourcesv1beta1.ConditionTypeError)
			Expect(cond).NotTo(BeNil())
			Expect(cond.Status).To(Equal(metav1.ConditionTrue))
			Expect(cond.Reason).To(Equal(cloudresourcesv1beta1.ConditionReasonMissingNfsVolume))
		})

		// Clean up
		By("// Cleanup", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), snapshot).
				Should(Succeed())
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), snapshot).
				Should(Succeed())
		})
	})
})
