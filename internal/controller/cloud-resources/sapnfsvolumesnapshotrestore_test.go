package cloudresources

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/gophercloud/gophercloud/v2/openstack/sharedfilesystems/v2/shares"
	"github.com/gophercloud/gophercloud/v2/openstack/sharedfilesystems/v2/snapshots"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	skrsapnfsvolume "github.com/kyma-project/cloud-manager/pkg/skr/sapnfsvolume"
	skrsapnfsvolumesnapshot "github.com/kyma-project/cloud-manager/pkg/skr/sapnfsvolumesnapshot"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	"github.com/kyma-project/cloud-manager/pkg/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Feature: SKR SapNfsVolumeSnapshotRestore", func() {

	It("Scenario: In-place revert restore completes successfully", func() {
		sapNfsVolumeName := "2f5d3139-e698-4a3e-b299-e73deb9fb887"
		sapNfsVolumeId := uuid.NewString()
		snapshotName := uuid.NewString()
		restoreName := uuid.NewString()
		sapNfsVolume := &cloudresourcesv1beta1.SapNfsVolume{}
		kcpNfsInstance := &cloudcontrolv1beta1.NfsInstance{}
		snapshot := &cloudresourcesv1beta1.SapNfsVolumeSnapshot{}
		restore := &cloudresourcesv1beta1.SapNfsVolumeSnapshotRestore{}
		scope := &cloudcontrolv1beta1.Scope{}

		sapMock := infra.SapMock().NewProject()

		kymaName := "3cf458d1-20a0-4946-82d6-d20cf0b24956"
		skrKymaRef := util.Must(infra.ScopeProvider().GetScope(infra.Ctx(), types.NamespacedName{Name: kymaName}))

		skrsapnfsvolume.Ignore.AddName(sapNfsVolumeName)
		defer skrsapnfsvolume.Ignore.RemoveName(sapNfsVolumeName)

		skrsapnfsvolumesnapshot.Ignore.AddName(snapshotName)
		defer skrsapnfsvolumesnapshot.Ignore.RemoveName(snapshotName)

		// Create a share in the mock so GetShare will find it
		var shareId string
		var openstackSnapshotId string

		By("Given KCP Scope exists", func() {
			Eventually(func() error {
				return GivenScopeOpenStackExists(infra.Ctx(), infra, scope, sapMock.ProviderParams(), WithName(skrKymaRef.Name))
			}).Should(Succeed())
		})

		By("And Given a Manila share exists in the mock", func() {
			share, err := sapMock.CreateShare(infra.Ctx(), shares.CreateOpts{
				ShareNetworkID: "test-net",
				Name:           "test-share",
				Size:           100,
			})
			Expect(err).NotTo(HaveOccurred())
			shareId = share.ID
			sapMock.SetShareStatus(shareId, "available")
		})

		By("And Given a Manila snapshot exists for that share", func() {
			snap, err := sapMock.CreateSnapshot(infra.Ctx(), snapshots.CreateOpts{
				ShareID: shareId,
				Name:    "test-snapshot",
			})
			Expect(err).NotTo(HaveOccurred())
			openstackSnapshotId = snap.ID
			sapMock.SetSnapshotStatus(openstackSnapshotId, "available")
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

		By("And Given SapNfsVolumeSnapshot exists in Ready state", func() {
			snapshot.Name = snapshotName
			snapshot.Namespace = DefaultSkrNamespace
			snapshot.Spec.SourceVolume = corev1.ObjectReference{
				Name: sapNfsVolumeName,
			}
			Eventually(func() error {
				return infra.SKR().Client().Create(infra.Ctx(), snapshot)
			}).Should(Succeed())

			// Set snapshot status directly to Ready with openstackId and shareId
			Eventually(func() error {
				err := infra.SKR().Client().Get(infra.Ctx(), client.ObjectKeyFromObject(snapshot), snapshot)
				if err != nil {
					return err
				}
				snapshot.Status.State = cloudresourcesv1beta1.StateReady
				snapshot.Status.OpenstackId = openstackSnapshotId
				snapshot.Status.ShareId = shareId
				snapshot.Status.Conditions = []metav1.Condition{
					{
						Type:               cloudresourcesv1beta1.ConditionTypeReady,
						Status:             metav1.ConditionTrue,
						Reason:             cloudresourcesv1beta1.ConditionReasonReady,
						Message:            "Ready",
						LastTransitionTime: metav1.Now(),
					},
				}
				return infra.SKR().Client().Status().Update(infra.Ctx(), snapshot)
			}).Should(Succeed())
		})

		By("When SapNfsVolumeSnapshotRestore is created for in-place revert", func() {
			restore.Name = restoreName
			restore.Namespace = DefaultSkrNamespace
			restore.Spec.SourceSnapshot = corev1.ObjectReference{
				Name: snapshotName,
			}
			restore.Spec.Destination = cloudresourcesv1beta1.SapNfsVolumeSnapshotRestoreDestination{
				ExistingVolume: &corev1.ObjectReference{
					Name: sapNfsVolumeName,
				},
			}
			Eventually(func() error {
				return infra.SKR().Client().Create(infra.Ctx(), restore)
			}).Should(Succeed())
		})

		By("Then the restore enters InProgress state", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), restore,
					NewObjActions(),
					HavingFieldValue(cloudresourcesv1beta1.JobStateInProgress, "status", "state"),
				).Should(Succeed(), "expected restore to be InProgress")
		})

		By("And Then revertInitiated is set to true", func() {
			Eventually(func() error {
				err := infra.SKR().Client().Get(infra.Ctx(), client.ObjectKeyFromObject(restore), restore)
				if err != nil {
					return err
				}
				if !restore.Status.RevertInitiated {
					return fmt.Errorf("revertInitiated not yet set")
				}
				return nil
			}).Should(Succeed())
		})

		By("When the share revert completes (share becomes available)", func() {
			sapMock.SetShareStatus(shareId, "available")
		})

		By("Then the restore reaches Done state", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), restore,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
					HavingFieldValue(cloudresourcesv1beta1.JobStateDone, "status", "state"),
				).Should(Succeed(), "expected restore to be Done")
		})

		// CleanUp
		Eventually(Delete).
			WithArguments(infra.Ctx(), infra.SKR().Client(), restore).
			Should(Succeed())
		Eventually(IsDeleted).
			WithArguments(infra.Ctx(), infra.SKR().Client(), restore).
			Should(Succeed(), "expected restore to be deleted")
	})

	It("Scenario: New-volume restore creates SapNfsVolume and completes", func() {
		sapNfsVolumeName := "6faf1777-61aa-44e6-a892-2382c09c986d"
		sapNfsVolumeId := uuid.NewString()
		snapshotName := uuid.NewString()
		restoreName := uuid.NewString()
		newVolumeName := "new-vol-" + uuid.NewString()[:8]
		sapNfsVolume := &cloudresourcesv1beta1.SapNfsVolume{}
		kcpNfsInstance := &cloudcontrolv1beta1.NfsInstance{}
		snapshot := &cloudresourcesv1beta1.SapNfsVolumeSnapshot{}
		restore := &cloudresourcesv1beta1.SapNfsVolumeSnapshotRestore{}
		scope := &cloudcontrolv1beta1.Scope{}

		sapMock := infra.SapMock().NewProject()

		kymaName := "25a08564-6294-4f0b-8e2d-cacaefbddfa8"
		skrKymaRef := util.Must(infra.ScopeProvider().GetScope(infra.Ctx(), types.NamespacedName{Name: kymaName}))

		skrsapnfsvolume.Ignore.AddName(sapNfsVolumeName)
		defer skrsapnfsvolume.Ignore.RemoveName(sapNfsVolumeName)

		skrsapnfsvolumesnapshot.Ignore.AddName(snapshotName)
		defer skrsapnfsvolumesnapshot.Ignore.RemoveName(snapshotName)

		// Also ignore the new volume so the SapNfsVolume reconciler doesn't process it
		skrsapnfsvolume.Ignore.AddName(newVolumeName)
		defer skrsapnfsvolume.Ignore.RemoveName(newVolumeName)

		var openstackSnapshotId string
		shareId := uuid.NewString()

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

		By("And Given SapNfsVolumeSnapshot exists in Ready state", func() {
			snapshot.Name = snapshotName
			snapshot.Namespace = DefaultSkrNamespace
			snapshot.Spec.SourceVolume = corev1.ObjectReference{
				Name: sapNfsVolumeName,
			}
			Eventually(func() error {
				return infra.SKR().Client().Create(infra.Ctx(), snapshot)
			}).Should(Succeed())

			openstackSnapshotId = uuid.NewString()
			// Set snapshot status directly to Ready with openstackId
			Eventually(func() error {
				err := infra.SKR().Client().Get(infra.Ctx(), client.ObjectKeyFromObject(snapshot), snapshot)
				if err != nil {
					return err
				}
				snapshot.Status.State = cloudresourcesv1beta1.StateReady
				snapshot.Status.OpenstackId = openstackSnapshotId
				snapshot.Status.ShareId = shareId
				snapshot.Status.Conditions = []metav1.Condition{
					{
						Type:               cloudresourcesv1beta1.ConditionTypeReady,
						Status:             metav1.ConditionTrue,
						Reason:             cloudresourcesv1beta1.ConditionReasonReady,
						Message:            "Ready",
						LastTransitionTime: metav1.Now(),
					},
				}
				return infra.SKR().Client().Status().Update(infra.Ctx(), snapshot)
			}).Should(Succeed())
		})

		By("When SapNfsVolumeSnapshotRestore is created for new-volume restore", func() {
			restore.Name = restoreName
			restore.Namespace = DefaultSkrNamespace
			restore.Spec.SourceSnapshot = corev1.ObjectReference{
				Name: snapshotName,
			}
			restore.Spec.Destination = cloudresourcesv1beta1.SapNfsVolumeSnapshotRestoreDestination{
				NewVolume: &cloudresourcesv1beta1.SapNfsVolumeSnapshotNewVolume{
					Metadata: cloudresourcesv1beta1.SapNfsVolumeSnapshotNewVolumeMetadata{
						Name: newVolumeName,
					},
					Spec: cloudresourcesv1beta1.SapNfsVolumeSpec{
						CapacityGb: 100,
					},
				},
			}
			Eventually(func() error {
				return infra.SKR().Client().Create(infra.Ctx(), restore)
			}).Should(Succeed())
		})

		By("Then the new SapNfsVolume is created with snapshot-id annotation", func() {
			newVol := &cloudresourcesv1beta1.SapNfsVolume{}
			Eventually(func() error {
				return infra.SKR().Client().Get(infra.Ctx(), types.NamespacedName{
					Name:      newVolumeName,
					Namespace: DefaultSkrNamespace,
				}, newVol)
			}).Should(Succeed(), "expected new SapNfsVolume to be created")

			Expect(newVol.Annotations).To(HaveKeyWithValue(
				cloudresourcesv1beta1.AnnotationSnapshotId,
				openstackSnapshotId,
			))
		})

		By("When the new SapNfsVolume reaches Ready state", func() {
			newVol := &cloudresourcesv1beta1.SapNfsVolume{}
			Eventually(func() error {
				return infra.SKR().Client().Get(infra.Ctx(), types.NamespacedName{
					Name:      newVolumeName,
					Namespace: DefaultSkrNamespace,
				}, newVol)
			}).Should(Succeed())

			newVol.Status.State = cloudresourcesv1beta1.StateReady
			newVol.Status.Conditions = []metav1.Condition{
				{
					Type:               cloudresourcesv1beta1.ConditionTypeReady,
					Status:             metav1.ConditionTrue,
					Reason:             cloudresourcesv1beta1.ConditionReasonReady,
					Message:            "Ready",
					LastTransitionTime: metav1.Now(),
				},
			}
			Eventually(func() error {
				if err := infra.SKR().Client().Get(infra.Ctx(), client.ObjectKeyFromObject(newVol), newVol); err != nil {
					return err
				}
				newVol.Status.State = cloudresourcesv1beta1.StateReady
				newVol.Status.Conditions = []metav1.Condition{
					{
						Type:               cloudresourcesv1beta1.ConditionTypeReady,
						Status:             metav1.ConditionTrue,
						Reason:             cloudresourcesv1beta1.ConditionReasonReady,
						Message:            "Ready",
						LastTransitionTime: metav1.Now(),
					},
				}
				return infra.SKR().Client().Status().Update(infra.Ctx(), newVol)
			}).Should(Succeed())
		})

		By("Then the restore reaches Done state with createdVolume reference", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), restore,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
					HavingFieldValue(cloudresourcesv1beta1.JobStateDone, "status", "state"),
				).Should(Succeed(), "expected restore to be Done")

			Expect(restore.Status.CreatedVolume).NotTo(BeNil())
			Expect(restore.Status.CreatedVolume.Name).To(Equal(newVolumeName))
		})

		// CleanUp
		Eventually(Delete).
			WithArguments(infra.Ctx(), infra.SKR().Client(), restore).
			Should(Succeed())
		Eventually(IsDeleted).
			WithArguments(infra.Ctx(), infra.SKR().Client(), restore).
			Should(Succeed(), "expected restore to be deleted")
	})

})
