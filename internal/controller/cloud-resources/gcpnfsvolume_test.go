package cloudresources

import (
	"fmt"
	"time"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	kcpiprange "github.com/kyma-project/cloud-manager/pkg/kcp/iprange"
	skriprange "github.com/kyma-project/cloud-manager/pkg/skr/iprange"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("SKR GcpNfsVolume workflows", func() {

	const (
		interval = time.Millisecond * 250
	)
	var (
		timeout = time.Second * 10
	)

	Context("Given SKR Cluster with SKR and KCP IPRanges in Ready state", Ordered, func() {

		skrIpRangeName := "gcp-iprange-1"
		skrIpRange := &cloudresourcesv1beta1.IpRange{}
		kcpIpRangeName := "513f20b4-7b73-4246-9397-f8dd55344479"
		kcpIpRange := &cloudcontrolv1beta1.IpRange{}

		BeforeEach(func() {
			//Create namespace if it doesn't exist.
			Eventually(CreateNamespace).
				WithArguments(infra.Ctx(), infra.SKR().Client(), &corev1.Namespace{}).
				Should(Succeed())

			// tell skriprange reconciler to ignore this SKR IpRange
			skriprange.Ignore.AddName(skrIpRangeName)

			//Create SKR IPRange if it doesn't exist.
			Eventually(CreateSkrIpRange).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrIpRange,
					WithName(skrIpRangeName),
				).
				Should(Succeed())

			// tell kcpiprange reconciler to ignore this KCP IpRange
			kcpiprange.Ignore.AddName(kcpIpRangeName)

			//Create KCP IPRange if it doesn't exist.
			Eventually(CreateKcpIpRange).
				WithArguments(
					infra.Ctx(), infra.KCP().Client(), kcpIpRange,
					WithName(kcpIpRangeName),
					WithNamespace(DefaultKcpNamespace),
					WithLabels(map[string]string{
						cloudcontrolv1beta1.LabelKymaName:        infra.SkrKymaRef().Name,
						cloudcontrolv1beta1.LabelRemoteName:      skrIpRangeName,
						cloudcontrolv1beta1.LabelRemoteNamespace: DefaultSkrNamespace,
					}),
				).
				Should(Succeed())

			//Update SKR IpRange status to Ready
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.KCP().Client(), kcpIpRange,
					WithKcpIpRangeStatusCidr(kcpIpRange.Spec.Cidr),
					WithConditions(KcpReadyCondition()),
				).
				Should(Succeed())

			//Update KCP IpRange status to Ready
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrIpRange,
					WithSkrIpRangeStatusCidr(skrIpRange.Spec.Cidr),
					WithSkrIpRangeStatusId(kcpIpRangeName),
					WithConditions(SkrReadyCondition()),
				).
				Should(Succeed())
		})

		Describe("When GcpNfsVolume Create is called", Ordered, func() {
			//Define variables.
			gcpNfsVolume := &cloudresourcesv1beta1.GcpNfsVolume{}
			kcpNfsInstance := &cloudcontrolv1beta1.NfsInstance{}
			pv := &corev1.PersistentVolume{}

			gcpNfsVolumeName := "gcp-nfs-volume-1"
			nfsIpAddress := "10.11.12.14"
			pvName := fmt.Sprintf("%s--%s", DefaultSkrNamespace, gcpNfsVolumeName)

			BeforeEach(func() {
				Eventually(CreateGcpNfsVolume).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), gcpNfsVolume,
						WithName(gcpNfsVolumeName),
						WithGcpNfsVolumeIpRange(skrIpRange.Name),
					).
					Should(Succeed())
			})
			It("Then GcpNfsVolume is created in SKR", func() {
				// load GcpNfsVolume to get ID
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(),
						infra.SKR().Client(),
						gcpNfsVolume,
						NewObjActions(),
						AssertGcpNfsVolumeHasId(),
					).
					Should(Succeed())
			})

			It("Then NfsInstance is created in KCP", func() {

				// check KCP NfsInstance is created with name=gcpNfsVolume.ID
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(),
						infra.KCP().Client(),
						kcpNfsInstance,
						NewObjActions(WithName(gcpNfsVolume.Status.Id)),
					).
					Should(Succeed())

				By("And has label cloud-manager.kyma-project.io/kymaName")
				Expect(kcpNfsInstance.Labels[cloudcontrolv1beta1.LabelKymaName]).To(Equal(infra.SkrKymaRef().Name))

				By("And has label cloud-manager.kyma-project.io/remoteName")
				Expect(kcpNfsInstance.Labels[cloudcontrolv1beta1.LabelRemoteName]).To(Equal(gcpNfsVolume.Name))

				By("And has label cloud-manager.kyma-project.io/remoteNamespace")
				Expect(kcpNfsInstance.Labels[cloudcontrolv1beta1.LabelRemoteNamespace]).To(Equal(gcpNfsVolume.Namespace))

				By("And has spec.scope.name equal to SKR Cluster kyma name")
				Expect(kcpNfsInstance.Spec.Scope.Name).To(Equal(infra.SkrKymaRef().Name))

				By("And has spec.remoteRef matching to to SKR IpRange")
				Expect(kcpNfsInstance.Spec.RemoteRef.Namespace).To(Equal(gcpNfsVolume.Namespace))
				Expect(kcpNfsInstance.Spec.RemoteRef.Name).To(Equal(gcpNfsVolume.Name))
			})

			Describe("When KCP NfsInstance is switched to Ready condition", Ordered, func() {
				BeforeEach(func() {
					Eventually(LoadAndCheck).
						WithArguments(
							infra.Ctx(),
							infra.KCP().Client(),
							kcpNfsInstance,
							NewObjActions(WithName(gcpNfsVolume.Status.Id)),
						).
						Should(Succeed())
					Eventually(UpdateStatus).
						WithArguments(
							infra.Ctx(), infra.KCP().Client(), kcpNfsInstance,
							WithConditions(KcpReadyCondition()),
							WithKcpNfsStatusState(cloudcontrolv1beta1.ReadyState),
							WithKcpNfsStatusHost(nfsIpAddress),
							WithKcpNfsStatusCapacity(gcpNfsVolume.Spec.CapacityGb),
						).
						Should(Succeed())
				})

				It("Then SKR GcpNfsVolume will get to Ready condition", func() {

					Eventually(LoadAndCheck).
						WithArguments(
							infra.Ctx(),
							infra.SKR().Client(),
							gcpNfsVolume,
							NewObjActions(),
							AssertHasConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
						).
						Should(Succeed())

					By("And has the File store Host")
					Expect(gcpNfsVolume.Status.Hosts).To(HaveLen(1))

					By("And has the capacity equal to provisioned value.")
					Expect(gcpNfsVolume.Status.CapacityGb).To(Equal(gcpNfsVolume.Spec.CapacityGb))
				})

				It("Then PersistentVolume is created in SKR", func() {
					Eventually(LoadAndCheck).
						WithArguments(
							infra.Ctx(),
							infra.SKR().Client(),
							pv,
							NewObjActions(
								WithName(pvName),
							),
						).
						Should(Succeed())

					By("And has the Server Name matching the provisioned File server host.")
					Expect(pv.Spec.PersistentVolumeSource.NFS.Server).To(Equal(gcpNfsVolume.Status.Hosts[0]))

					By("And has the Volume Name matching the requested FileShare name.")
					path := fmt.Sprintf("/%s", gcpNfsVolume.Spec.FileShareName)
					Expect(pv.Spec.PersistentVolumeSource.NFS.Path).To(Equal(path))

					By("And has the Capacity equal to requested value in GB.")
					expectedCapacity := int64(gcpNfsVolume.Status.CapacityGb) * 1024 * 1024 * 1024
					quantity, _ := pv.Spec.Capacity["storage"]
					pvQuantity, _ := quantity.AsInt64()
					Expect(pvQuantity).To(Equal(expectedCapacity))
				})
			})
		})

		Describe("When SKR GcpNfsVolume Update is called with increased CapacityGb", Ordered, func() {
			//Define variable
			gcpNfsVolume := &cloudresourcesv1beta1.GcpNfsVolume{}
			kcpNfsInstance := &cloudcontrolv1beta1.NfsInstance{}
			pv := &corev1.PersistentVolume{}

			gcpNfsVolumeName := "gcp-nfs-volume-2"
			nfsIpAddress := "10.11.12.16"
			pvName := fmt.Sprintf("%s--%s", DefaultSkrNamespace, gcpNfsVolumeName)
			updatedCapacityGb := 2048

			BeforeEach(func() {
				//Create GcpNfsVolume
				Eventually(CreateGcpNfsVolume).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), gcpNfsVolume,
						WithName(gcpNfsVolumeName),
						WithGcpNfsVolumeIpRange(skrIpRange.Name),
					).
					Should(Succeed())

				// Load GcpNfsVolume and check the Status.ID
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(),
						infra.SKR().Client(),
						gcpNfsVolume,
						NewObjActions(),
						AssertGcpNfsVolumeHasId(),
					).
					Should(Succeed())

				// Load KCP NfsInstance is created with name=gcpNfsVolume.ID
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(),
						infra.KCP().Client(),
						kcpNfsInstance,
						NewObjActions(WithName(gcpNfsVolume.Status.Id)),
					).
					Should(Succeed())

				//Update KCP NfsInstance to Ready state
				Eventually(UpdateStatus).
					WithArguments(
						infra.Ctx(), infra.KCP().Client(), kcpNfsInstance,
						WithConditions(KcpReadyCondition()),
						WithKcpNfsStatusState(cloudcontrolv1beta1.ReadyState),
						WithKcpNfsStatusHost(nfsIpAddress),
						WithKcpNfsStatusCapacity(gcpNfsVolume.Spec.CapacityGb),
					).
					Should(Succeed())

				Eventually(Update).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), gcpNfsVolume,
						WithGcpNfsVolumeCapacity(updatedCapacityGb),
					).
					Should(Succeed())
			})

			It("Then SKR GcpNfsVolume is updated ", func() {
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), gcpNfsVolume,
						NewObjActions(),
					).
					Should(Succeed())

				By("And has the updated CapacityGb value.")
				Expect(gcpNfsVolume.Spec.CapacityGb).To(Equal(updatedCapacityGb))
			})

			It("And Then KCP NfsInstance is updated", func() {
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(),
						infra.KCP().Client(),
						kcpNfsInstance,
						NewObjActions(),
					).
					Should(Succeed())

				By("And has the CapacityGb matching that of the SKR GcpNfsVolume.")
				Expect(kcpNfsInstance.Spec.Instance.Gcp.CapacityGb).To(Equal(updatedCapacityGb))
			})

			Describe("When KCP NfsInstance Status is updated with new Capacity.", Ordered, func() {
				BeforeEach(func() {
					Eventually(LoadAndCheck).
						WithArguments(
							infra.Ctx(),
							infra.KCP().Client(),
							kcpNfsInstance,
							NewObjActions(WithName(gcpNfsVolume.Status.Id)),
						).
						Should(Succeed())

					Eventually(UpdateStatus).
						WithArguments(
							infra.Ctx(), infra.KCP().Client(), kcpNfsInstance,
							WithConditions(KcpReadyCondition()),
							WithKcpNfsStatusState(cloudcontrolv1beta1.ReadyState),
							WithKcpNfsStatusHost(nfsIpAddress),
							WithKcpNfsStatusCapacity(updatedCapacityGb),
						).
						Should(Succeed())
				})
				It("Then SKR GcpNfsVolume status is updated", func() {
					Eventually(LoadAndCheck).
						WithArguments(
							infra.Ctx(),
							infra.SKR().Client(),
							gcpNfsVolume,
							NewObjActions(),
						).
						Should(Succeed())

					By("And has status CapacityGb matching that of the SKR GcpNfsVolume Spec.")
					Expect(gcpNfsVolume.Status.CapacityGb).To(Equal(updatedCapacityGb))
				})

				It("And Then SKR PersistentVolume is updated", func() {
					Eventually(LoadAndCheck).
						WithArguments(
							infra.Ctx(),
							infra.SKR().Client(),
							pv,
							NewObjActions(
								WithName(pvName),
							),
						).
						Should(Succeed())

					By("And has the Capacity matching that of the SKR GcpNfsVolume.")
					expectedCapacity := int64(updatedCapacityGb) * 1024 * 1024 * 1024
					quantity, _ := pv.Spec.Capacity["storage"]
					pvQuantity, _ := quantity.AsInt64()
					Expect(pvQuantity).To(Equal(expectedCapacity))
				})
			})
		})

		Describe("When SKR GcpNfsVolume Delete is called", Ordered, func() {
			gcpNfsVolume := &cloudresourcesv1beta1.GcpNfsVolume{}
			kcpNfsInstance := &cloudcontrolv1beta1.NfsInstance{}
			pv := &corev1.PersistentVolume{}

			gcpNfsVolumeName := "gcp-nfs-volume-3"
			nfsIpAddress := "10.11.12.16"
			pvName := fmt.Sprintf("%s--%s", DefaultSkrNamespace, gcpNfsVolumeName)

			BeforeEach(func() {
				//Create GcpNfsVolume
				Eventually(CreateGcpNfsVolume).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), gcpNfsVolume,
						WithName(gcpNfsVolumeName),
						WithGcpNfsVolumeIpRange(skrIpRange.Name),
					).
					Should(Succeed())

				// Load GcpNfsVolume and check the Status.ID
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(),
						infra.SKR().Client(),
						gcpNfsVolume,
						NewObjActions(),
						AssertGcpNfsVolumeHasId(),
					).
					Should(Succeed())

				// Load KCP NfsInstance is created with name=gcpNfsVolume.ID
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(),
						infra.KCP().Client(),
						kcpNfsInstance,
						NewObjActions(WithName(gcpNfsVolume.Status.Id)),
					).
					Should(Succeed())

				//Update KCP NfsInstance to Ready state
				Eventually(UpdateStatus).
					WithArguments(
						infra.Ctx(), infra.KCP().Client(), kcpNfsInstance,
						WithConditions(KcpReadyCondition()),
						WithKcpNfsStatusState(cloudcontrolv1beta1.ReadyState),
						WithKcpNfsStatusHost(nfsIpAddress),
						WithKcpNfsStatusCapacity(gcpNfsVolume.Spec.CapacityGb),
					).
					Should(Succeed())

				//Load PV
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), pv,
						NewObjActions(
							WithName(pvName),
						),
					).
					Should(Succeed())

				//Delete SKR GcpNfsVolume
				Eventually(Delete).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), gcpNfsVolume,
					).Should(Succeed())
			})

			It("Then SKR GcpNfsVolume is marked for Deletion ", func() {
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), gcpNfsVolume,
						NewObjActions(),
					).
					Should(Succeed())

				By("And has DeletionTimestamp set.")
				Expect(gcpNfsVolume.DeletionTimestamp.IsZero()).NotTo(BeTrue())
			})

			Describe("When PV phase changes to Available, and all finalizers removed", Ordered, func() {
				BeforeEach(func() {
					//Update PV Phase to Available.
					Eventually(UpdateStatus).
						WithArguments(
							infra.Ctx(),
							infra.SKR().Client(),
							pv,
							WithPvStatusPhase(corev1.VolumeAvailable),
						).
						Should(Succeed())

					//Remove pv-protection finalizer
					Eventually(Update).
						WithArguments(
							infra.Ctx(), infra.SKR().Client(), pv,
							RemoveFinalizer("kubernetes.io/pv-protection"),
						).
						Should(Succeed())
				})
				It("Then PV, NfsInstance, and GcpNfsVolume are all deleted.", func() {
					Eventually(IsDeleted, timeout, interval).
						WithArguments(
							infra.Ctx(), infra.SKR().Client(), pv,
						).
						Should(BeTrue())
					By("And the PersistentVolume in SKR is deleted.")

					Eventually(IsDeleted, timeout, interval).
						WithArguments(
							infra.Ctx(), infra.KCP().Client(), kcpNfsInstance,
						).
						Should(BeTrue())
					By("And the NfsInstance in KCP is deleted.")

					Eventually(IsDeleted, timeout, interval).
						WithArguments(
							infra.Ctx(), infra.SKR().Client(), gcpNfsVolume,
						).
						Should(BeTrue())
					By("And the GcpNfsVolume in SKR is deleted.")
				})
			})
		})
	})
})
