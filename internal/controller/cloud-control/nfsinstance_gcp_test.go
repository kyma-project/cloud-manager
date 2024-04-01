package cloudcontrol

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	iprangePkg "github.com/kyma-project/cloud-manager/pkg/kcp/iprange"
	client2 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	scopePkg "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	"github.com/kyma-project/cloud-manager/pkg/util/debugged"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/api/googleapi"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

var _ = Describe("Feature: KCP NFSVolume for GCP", func() {
	const (
		kymaName = "5b30a61a-c4ae-49da-a8ad-903a71696d8b"

		interval = time.Millisecond * 250
	)
	var (
		timeout = time.Second * 20
	)

	if debugged.Debugged {
		timeout = time.Minute * 20
	}
	Context("Scenario: GCP NFSVolume Happy Path", Ordered, func() {
		gcpNfsInstance := &cloudcontrolv1beta1.NfsInstance{}
		It("Scenario: GCP NFSVolume Creation", func() {
			scope := &cloudcontrolv1beta1.Scope{}
			By("Given KCP Cluster", func() {

				// Tell Scope reconciler to ignore this kymaName
				scopePkg.Ignore.AddName(kymaName)
			})
			By("And Given KCP Scope exists", func() {

				// Given Scope exists
				Expect(
					infra.GivenScopeGcpExists(kymaName),
				).NotTo(HaveOccurred())

				// Load created scope
				Eventually(func() (exists bool, err error) {
					err = infra.KCP().Client().Get(infra.Ctx(), infra.KCP().ObjKey(kymaName), scope)
					exists = err == nil
					return exists, client.IgnoreNotFound(err)
				}, timeout, interval).
					Should(BeTrue(), "expected Scope to get created")
			})
			kcpIpRangeName := "gcp-nfs-iprange-1"
			kcpIpRange := &cloudcontrolv1beta1.IpRange{}

			// Tell IpRange reconciler to ignore this kymaName
			iprangePkg.Ignore.AddName(kcpIpRangeName)
			By("And Given KCP IPRange exists", func() {
				Eventually(CreateKcpIpRange).
					WithArguments(
						infra.Ctx(), infra.KCP().Client(), kcpIpRange,
						WithName(kcpIpRangeName),
						WithKcpIpRangeSpecScope(scope.Name),
					).
					Should(Succeed())
			})

			By("And Given KCP IpRange has Ready condition", func() {
				Eventually(UpdateStatus).
					WithArguments(
						infra.Ctx(), infra.KCP().Client(), kcpIpRange,
						WithKcpIpRangeStatusCidr(kcpIpRange.Spec.Cidr),
						WithConditions(KcpReadyCondition()),
					).WithTimeout(timeout).WithPolling(interval).
					Should(Succeed())
			})

			By("When KCP NfsVolume is created", func() {
				Eventually(CreateNfsInstance).
					WithArguments(
						infra.Ctx(), infra.KCP().Client(), gcpNfsInstance,
						WithName("gcp-nfs-instance-1"),
						WithRemoteRef("gcp-nfs-instance-1"),
						WithNfsInstanceScope(scope.Name),
						WithNfsInstanceIpRange(kcpIpRange.Name),
						WithNfsInstanceGcp(scope.Spec.Region),
					).
					Should(Succeed())
			})
			By("Then KCP NfsVolume will get Ready condition", func() {
				Eventually(func() (exists bool, err error) {
					err = infra.KCP().Client().Get(infra.Ctx(), client.ObjectKeyFromObject(gcpNfsInstance), gcpNfsInstance)
					if err != nil {
						return false, err
					}
					exists = meta.IsStatusConditionTrue(gcpNfsInstance.Status.Conditions, cloudcontrolv1beta1.ConditionTypeReady)
					return exists, nil
				}, timeout, interval).
					Should(BeTrue(), "expected NfsInstance for GCP with Ready condition")
			})
			By("And KCP NfsVolume has Ready state", func() {
				Expect(gcpNfsInstance.Status.State).To(Equal(cloudcontrolv1beta1.ReadyState))
			})
		})

		It("Scenario: GCP NFSVolume Updating", func() {
			By("Given NfsVolume is Ready", func() {
			})

			By("When KCP NfsVolume is updated", func() {
				Eventually(UpdateNfsInstance).
					WithArguments(
						infra.Ctx(), infra.KCP().Client(), gcpNfsInstance,
					).
					Should(Succeed())
			})
			By("Then at first KCP NfsVolume will get SyncFilestore state", func() {
				Eventually(func() (exists bool, err error) {
					err = infra.KCP().Client().Get(infra.Ctx(), client.ObjectKeyFromObject(gcpNfsInstance), gcpNfsInstance)
					if err != nil {
						return false, err
					}
					exists = gcpNfsInstance.Status.State == client2.SyncFilestore
					return exists, nil
				}, timeout, interval)
			})
			By("And NfsVolume keeps Ready state", func() {
				Expect(meta.IsStatusConditionTrue(gcpNfsInstance.Status.Conditions, cloudcontrolv1beta1.ConditionTypeReady)).To(BeTrue())
			})
			By("And eventually KCP NfsVolume will get Ready state", func() {
				Eventually(func() (exists bool, err error) {
					err = infra.KCP().Client().Get(infra.Ctx(), client.ObjectKeyFromObject(gcpNfsInstance), gcpNfsInstance)
					if err != nil {
						return false, err
					}
					exists = gcpNfsInstance.Status.State == cloudcontrolv1beta1.ReadyState
					return exists, nil
				}, timeout, interval)
			})
		})

		It("Scenario: GCP NFSVolume Deletion", func() {
			By("Given NfsVolume is Ready", func() {
			})
			By("When KCP NfsVolume is deleted", func() {
				Eventually(DeleteNfsInstance).
					WithArguments(
						infra.Ctx(), infra.KCP().Client(), gcpNfsInstance,
					).
					Should(Succeed())
			})
			By("Then at first KCP NfsVolume will get Deleted state", func() {
				Eventually(func() (exists bool, err error) {
					err = infra.KCP().Client().Get(infra.Ctx(), client.ObjectKeyFromObject(gcpNfsInstance), gcpNfsInstance)
					if err != nil {
						return false, err
					}
					exists = gcpNfsInstance.Status.State == client2.Deleted
					return exists, nil
				}, timeout, interval)
			})
			By("And NfsVolume keeps Ready state", func() {
				Expect(meta.IsStatusConditionTrue(gcpNfsInstance.Status.Conditions, cloudcontrolv1beta1.ConditionTypeReady)).To(BeTrue())
			})
			By("And eventually KCP NfsVolume will get permanently removed", func() {
				Eventually(func() (exists bool, err error) {
					err = infra.KCP().Client().Get(infra.Ctx(), client.ObjectKeyFromObject(gcpNfsInstance), gcpNfsInstance)
					exists = apierrors.IsNotFound(err)
					return exists, client.IgnoreNotFound(err)
				}, timeout, interval)
			})
		})
	})

	Context("Scenario: GCP NFSVolume Unhappy Path", Ordered, func() {
		gcpNfsInstance := &cloudcontrolv1beta1.NfsInstance{}
		scope := &cloudcontrolv1beta1.Scope{}
		kcpIpRangeName := "gcp-nfs-iprange-1"
		kcpIpRange := &cloudcontrolv1beta1.IpRange{}
		It("Scenario: GCP NFSVolume Creation Get Failure", func() {
			By("Given KCP Cluster", func() {

				// Tell Scope reconciler to ignore this kymaName
				scopePkg.Ignore.AddName(kymaName)
			})
			By("And Given KCP Scope exists", func() {

				// Given Scope exists
				Expect(
					infra.GivenScopeGcpExists(kymaName),
				).NotTo(HaveOccurred())

				// Load created scope
				Eventually(func() (exists bool, err error) {
					err = infra.KCP().Client().Get(infra.Ctx(), infra.KCP().ObjKey(kymaName), scope)
					exists = err == nil
					return exists, client.IgnoreNotFound(err)
				}, timeout, interval).
					Should(BeTrue(), "expected Scope to get created")
			})

			// Tell IpRange reconciler to ignore this kymaName
			iprangePkg.Ignore.AddName(kcpIpRangeName)
			By("And Given KCP IPRange exists", func() {
				Eventually(CreateKcpIpRange).
					WithArguments(
						infra.Ctx(), infra.KCP().Client(), kcpIpRange,
						WithName(kcpIpRangeName),
						WithKcpIpRangeSpecScope(scope.Name),
					).
					Should(Succeed())
			})

			By("And Given KCP IpRange has Ready condition", func() {
				Eventually(UpdateStatus).
					WithArguments(
						infra.Ctx(), infra.KCP().Client(), kcpIpRange,
						WithKcpIpRangeStatusCidr(kcpIpRange.Spec.Cidr),
						WithConditions(KcpReadyCondition()),
					).WithTimeout(timeout).WithPolling(interval).
					Should(Succeed())
			})

			By("When KCP NfsVolume is created but Filestore Get call fails", func() {
				infra.GcpMock().SetGetError(sample500Error())
				Eventually(CreateNfsInstance, timeout, interval).
					WithArguments(
						infra.Ctx(), infra.KCP().Client(), gcpNfsInstance,
						WithName("gcp-nfs-instance-2"),
						WithRemoteRef("gcp-nfs-instance-2"),
						WithNfsInstanceScope(scope.Name),
						WithNfsInstanceIpRange(kcpIpRange.Name),
						WithNfsInstanceGcp(scope.Spec.Region),
					).
					Should(Succeed())
			})
			By("Then KCP NfsVolume will get Error condition", func() {
				Eventually(func() (exists bool, err error) {
					err = infra.KCP().Client().Get(infra.Ctx(), client.ObjectKeyFromObject(gcpNfsInstance), gcpNfsInstance)
					if err != nil {
						return false, err
					}
					exists = meta.IsStatusConditionTrue(gcpNfsInstance.Status.Conditions, cloudcontrolv1beta1.ConditionTypeError)
					return exists, nil
				}, timeout, interval).
					Should(BeTrue(), "expected NfsInstance for GCP with Error condition")
			})
			By("And KCP NfsVolume has empty state", func() {
				Expect(len(gcpNfsInstance.Status.State)).To(Equal(0))
			})

		})
		It("Scenario: GCP NFSVolume Creation Call Failure", func() {
			By("Given KCP NfsVolume is created", func() {
				infra.GcpMock().SetGetError(nil)
			})
			By("When Filestore creation call fails", func() {
				infra.GcpMock().SetCreateError(sample500Error())
				// CreateNfsInstance is already called in previous test case
			})
			By("Then KCP NfsVolume has SyncFilestore state", func() {
				Eventually(func() (exists bool, err error) {
					err = infra.KCP().Client().Get(infra.Ctx(), client.ObjectKeyFromObject(gcpNfsInstance), gcpNfsInstance)
					if err != nil {
						return false, err
					}
					exists = gcpNfsInstance.Status.State == client2.SyncFilestore
					return exists, nil
				}, timeout, interval).
					Should(BeTrue(), "expected NfsInstance for GCP with SyncFilestore state")
				//Expect(gcpNfsInstance.Status.State).To(Equal(client2.SyncFilestore))
			})
			By("And KCP NfsVolume will get Error condition", func() {
				Expect(meta.IsStatusConditionTrue(gcpNfsInstance.Status.Conditions, cloudcontrolv1beta1.ConditionTypeError)).To(BeTrue())
			})
		})
		It("Scenario: GCP NFSVolume Creation Operation Failure", func() {
			By("Given KCP NfsVolume is created", func() {
			})
			By("And Filestore creation submitted", func() {
				infra.GcpMock().SetCreateError(nil)
			})
			By("When creation operation fails", func() {
				infra.GcpMock().SetOperationError(sample500Error())
			})
			By("Then KCP NfsVolume will get Error condition", func() {
				Eventually(func() (exists bool, err error) {
					err = infra.KCP().Client().Get(infra.Ctx(), client.ObjectKeyFromObject(gcpNfsInstance), gcpNfsInstance)
					if err != nil {
						return false, err
					}
					exists = meta.IsStatusConditionTrue(gcpNfsInstance.Status.Conditions, cloudcontrolv1beta1.ConditionTypeError)
					return exists, nil
				}, timeout, interval).
					Should(BeTrue(), "expected NfsInstance for GCP with Error condition")
			})
			By("And KCP NfsVolume has SyncFilestore state", func() {
				Expect(gcpNfsInstance.Status.State).To(Equal(client2.SyncFilestore))
			})

		})
		It("Scenario: GCP NFSVolume Update call Failure", func() {
			By("Given KCP NfsVolume has ready condition", func() {
				infra.GcpMock().SetOperationError(nil)
				Eventually(func() (exists bool, err error) {
					err = infra.KCP().Client().Get(infra.Ctx(), client.ObjectKeyFromObject(gcpNfsInstance), gcpNfsInstance)
					if err != nil {
						return false, err
					}
					exists = meta.IsStatusConditionTrue(gcpNfsInstance.Status.Conditions, cloudcontrolv1beta1.ConditionTypeReady)
					return exists, nil
				}, timeout, interval).
					Should(BeTrue(), "expected NfsInstance for GCP with Ready condition")
			})
			By("When KCP NfsVolume is updated", func() {
				infra.GcpMock().SetPatchError(sample500Error())
				Eventually(UpdateNfsInstance).
					WithArguments(
						infra.Ctx(), infra.KCP().Client(), gcpNfsInstance,
					).
					Should(Succeed())
			})
			By("And FileStore patch request fails", func() {
			})
			By("Then KCP NfsVolume will get Error condition", func() {
				Eventually(func() (exists bool, err error) {
					err = infra.KCP().Client().Get(infra.Ctx(), client.ObjectKeyFromObject(gcpNfsInstance), gcpNfsInstance)
					if err != nil {
						return false, err
					}
					exists = meta.IsStatusConditionTrue(gcpNfsInstance.Status.Conditions, cloudcontrolv1beta1.ConditionTypeError)
					return exists, nil
				}, timeout, interval).
					Should(BeTrue(), "expected NfsInstance for GCP with Error condition")
			})
			By("And KCP NfsVolume has SyncFilestore state", func() {
				Expect(gcpNfsInstance.Status.State).To(Equal(client2.SyncFilestore))
			})

		})
		It("Scenario: GCP NFSVolume Update Operation Failure", func() {
			By("Given NfsVolume is updated", func() {
				// Already updated. No need to call again
			})
			By("And Filestore patch request submitted", func() {
				infra.GcpMock().SetPatchError(nil)
			})
			By("When Filestore patch operation fails", func() {
				infra.GcpMock().SetOperationError(sample500Error())
			})
			By("Then KCP NfsVolume will get Error condition", func() {
				Eventually(func() (exists bool, err error) {
					err = infra.KCP().Client().Get(infra.Ctx(), client.ObjectKeyFromObject(gcpNfsInstance), gcpNfsInstance)
					if err != nil {
						return false, err
					}
					exists = meta.IsStatusConditionTrue(gcpNfsInstance.Status.Conditions, cloudcontrolv1beta1.ConditionTypeError)
					return exists, nil
				}, timeout, interval).
					Should(BeTrue(), "expected NfsInstance for GCP with Error condition")
			})
			By("And KCP NfsVolume has SyncFilestore state", func() {
				Expect(gcpNfsInstance.Status.State).To(Equal(client2.SyncFilestore))
			})

		})
		It("Scenario: GCP NFSVolume Delete call Failure", func() {
			By("Given KCP NfsVolume has ready condition", func() {
				infra.GcpMock().SetOperationError(nil)
				Eventually(func() (exists bool, err error) {
					err = infra.KCP().Client().Get(infra.Ctx(), client.ObjectKeyFromObject(gcpNfsInstance), gcpNfsInstance)
					if err != nil {
						return false, err
					}
					exists = meta.IsStatusConditionTrue(gcpNfsInstance.Status.Conditions, cloudcontrolv1beta1.ConditionTypeReady)
					return exists, nil
				}, timeout, interval).
					Should(BeTrue(), "expected NfsInstance for GCP with Ready condition")
			})
			By("When KCP NfsVolume is deleted", func() {
				infra.GcpMock().SetDeleteError(sample500Error())
				Eventually(DeleteNfsInstance).
					WithArguments(
						infra.Ctx(), infra.KCP().Client(), gcpNfsInstance,
					).
					Should(Succeed())
			})
			By("And Delete patch request fails", func() {
				// We set the DeleteError on mock client to simulate the failure
			})
			By("Then KCP NfsVolume will get Error condition", func() {
				Eventually(func() (exists bool, err error) {
					err = infra.KCP().Client().Get(infra.Ctx(), client.ObjectKeyFromObject(gcpNfsInstance), gcpNfsInstance)
					if err != nil {
						return false, err
					}
					exists = meta.IsStatusConditionTrue(gcpNfsInstance.Status.Conditions, cloudcontrolv1beta1.ConditionTypeError)
					return exists, nil
				}, timeout, interval).
					Should(BeTrue(), "expected NfsInstance for GCP with Error condition")
			})
			By("And KCP NfsVolume has SyncFilestore state", func() {
				Expect(gcpNfsInstance.Status.State).To(Equal(client2.SyncFilestore))
			})

		})
		It("Scenario: GCP NFSVolume Delete Operation Failure", func() {
			By("Given NfsVolume is deleted", func() {
				// Already updated. No need to call again
			})
			By("And Filestore delete request submitted", func() {
				infra.GcpMock().SetPatchError(nil)
			})
			By("When Filestore delete operation fails", func() {
				infra.GcpMock().SetOperationError(sample500Error())
			})
			By("Then KCP NfsVolume will get Error condition", func() {
				Eventually(func() (exists bool, err error) {
					err = infra.KCP().Client().Get(infra.Ctx(), client.ObjectKeyFromObject(gcpNfsInstance), gcpNfsInstance)
					if err != nil {
						return false, err
					}
					exists = meta.IsStatusConditionTrue(gcpNfsInstance.Status.Conditions, cloudcontrolv1beta1.ConditionTypeError)
					return exists, nil
				}, timeout, interval).
					Should(BeTrue(), "expected NfsInstance for GCP with Error condition")
			})
			By("And KCP NfsVolume has SyncFilestore state", func() {
				Expect(gcpNfsInstance.Status.State).To(Equal(client2.SyncFilestore))
			})
			// Let it be deleted
			infra.GcpMock().SetOperationError(nil)
		})

	})
})

func sample500Error() *googleapi.Error {
	return &googleapi.Error{Code: 500, Message: "Internal Server Error"}
}
