package cloudcontrol

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/api"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/resource"
	"time"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	kcpiprange "github.com/kyma-project/cloud-manager/pkg/kcp/iprange"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	kcpscope "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	"github.com/kyma-project/cloud-manager/pkg/util/debugged"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/api/googleapi"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
				kcpscope.Ignore.AddName(kymaName)
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
			kcpiprange.Ignore.AddName(kcpIpRangeName)
			By("And Given KCP IPRange exists", func() {
				Eventually(CreateKcpIpRange).
					WithArguments(
						infra.Ctx(), infra.KCP().Client(), kcpIpRange,
						WithName(kcpIpRangeName),
						WithScope(scope.Name),
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
						WithScope(scope.Name),
						WithIpRange(kcpIpRange.Name),
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
				Expect(gcpNfsInstance.Status.State).To(Equal(cloudcontrolv1beta1.StateReady))
			})
			By("And KCP NfsVolume has correct status.Capacity", func() {
				Expect(gcpNfsInstance.Status.CapacityGb).To(Equal(DefaultGcpNfsInstanceCapacityGb))
				Expect(gcpNfsInstance.Status.Capacity).To(BeComparableTo(resource.MustParse(fmt.Sprintf("%dGi", DefaultGcpNfsInstanceCapacityGb))))
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
			By("And NfsVolume keeps Ready state", func() {
				Expect(meta.IsStatusConditionTrue(gcpNfsInstance.Status.Conditions, cloudcontrolv1beta1.ConditionTypeReady)).To(BeTrue())
			})
			By("And eventually KCP NfsVolume will get Ready state", func() {
				Eventually(func() (exists bool, err error) {
					err = infra.KCP().Client().Get(infra.Ctx(), client.ObjectKeyFromObject(gcpNfsInstance), gcpNfsInstance)
					if err != nil {
						return false, err
					}
					exists = gcpNfsInstance.Status.State == cloudcontrolv1beta1.StateReady
					return exists, nil
				}, timeout, interval).Should(BeTrue(), "expected NfsInstance for GCP with Ready state")
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
			By("And NfsVolume keeps Ready state", func() {
				Expect(meta.IsStatusConditionTrue(gcpNfsInstance.Status.Conditions, cloudcontrolv1beta1.ConditionTypeReady)).To(BeTrue())
			})
			By("And eventually KCP NfsVolume will get permanently removed", func() {
				Eventually(func() (exists bool, err error) {
					err = infra.KCP().Client().Get(infra.Ctx(), client.ObjectKeyFromObject(gcpNfsInstance), gcpNfsInstance)
					exists = apierrors.IsNotFound(err)
					return exists, client.IgnoreNotFound(err)
				}, timeout, interval).Should(BeTrue(), "expected NfsInstance for GCP to be deleted")
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
				kcpscope.Ignore.AddName(kymaName)
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
			kcpiprange.Ignore.AddName(kcpIpRangeName)
			By("And Given KCP IPRange exists", func() {
				Eventually(CreateKcpIpRange).
					WithArguments(
						infra.Ctx(), infra.KCP().Client(), kcpIpRange,
						WithName(kcpIpRangeName),
						WithScope(scope.Name),
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
						WithScope(scope.Name),
						WithIpRange(kcpIpRange.Name),
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
			By("Then KCP NfsVolume has Creating state", func() {
				Eventually(func() (exists bool, err error) {
					err = infra.KCP().Client().Get(infra.Ctx(), client.ObjectKeyFromObject(gcpNfsInstance), gcpNfsInstance)
					if err != nil {
						return false, err
					}
					exists = gcpNfsInstance.Status.State == gcpclient.Creating
					return exists, nil
				}, timeout, interval).
					Should(BeTrue(), "expected NfsInstance for GCP with Creating state")
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
			By("And KCP NfsVolume has Creating state", func() {
				Expect(gcpNfsInstance.Status.State).To(Equal(gcpclient.Creating))
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
			By("And KCP NfsVolume has Updating state", func() {
				Expect(gcpNfsInstance.Status.State).To(Equal(gcpclient.Updating))
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
				Expect(gcpNfsInstance.Status.State).To(Equal(gcpclient.Updating))
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
			By("And KCP NfsVolume has Deleting state", func() {
				Expect(gcpNfsInstance.Status.State).To(Equal(gcpclient.Deleting))
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
			By("And KCP NfsVolume has Deleting state", func() {
				Expect(gcpNfsInstance.Status.State).To(Equal(gcpclient.Deleting))
			})
			// Let it be deleted
			infra.GcpMock().SetOperationError(nil)
		})

		It("// cleanup: delete GCP NfsVolume for unhappy path", func() {
			infra.GcpMock().SetCreateError(nil)
			infra.GcpMock().SetPatchError(nil)
			infra.GcpMock().SetDeleteError(nil)
			infra.GcpMock().SetGetError(nil)
			infra.GcpMock().SetOperationError(nil)
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), gcpNfsInstance).
				Should(Succeed())
			Eventually(func(ctx context.Context) error {
				_, err := composed.PatchObjRemoveFinalizer(ctx, api.CommonFinalizerDeletionHook, gcpNfsInstance, infra.KCP().Client())
				return err
			}).
				WithContext(infra.Ctx()).
				Should(Succeed())
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), gcpNfsInstance).
				Should(Succeed())
		})
	})

	Context("Scenario: GCP NFSVolume With Restore Happy Path", Ordered, func() {
		gcpNfsInstance := &cloudcontrolv1beta1.NfsInstance{}
		It("Scenario: GCP NFSVolume Creation", func() {
			scope := &cloudcontrolv1beta1.Scope{}
			By("Given KCP Cluster", func() {

				// Tell Scope reconciler to ignore this kymaName
				kcpscope.Ignore.AddName(kymaName)
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
			kcpiprange.Ignore.AddName(kcpIpRangeName)
			By("And Given KCP IPRange exists", func() {
				Eventually(CreateKcpIpRange).
					WithArguments(
						infra.Ctx(), infra.KCP().Client(), kcpIpRange,
						WithName(kcpIpRangeName),
						WithScope(scope.Name),
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
						WithName("gcp-nfs-instance-3"),
						WithRemoteRef("gcp-nfs-instance-3"),
						WithScope(scope.Name),
						WithIpRange(kcpIpRange.Name),
						WithNfsInstanceGcp(scope.Spec.Region),
						WithSourceBackup("projects/kyma/locations/us-west1/backups/gcp-nfs-backup-1"),
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
				Expect(gcpNfsInstance.Status.State).To(Equal(cloudcontrolv1beta1.StateReady))
			})
			By("And KCP NfsVolume has correct status.Capacity", func() {
				Expect(gcpNfsInstance.Status.CapacityGb).To(Equal(DefaultGcpNfsInstanceCapacityGb))
				Expect(gcpNfsInstance.Status.Capacity).To(BeComparableTo(resource.MustParse(fmt.Sprintf("%dGi", DefaultGcpNfsInstanceCapacityGb))))
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
			By("And NfsVolume keeps Ready state", func() {
				Expect(meta.IsStatusConditionTrue(gcpNfsInstance.Status.Conditions, cloudcontrolv1beta1.ConditionTypeReady)).To(BeTrue())
			})
			By("And eventually KCP NfsVolume will get Ready state", func() {
				Eventually(func() (exists bool, err error) {
					err = infra.KCP().Client().Get(infra.Ctx(), client.ObjectKeyFromObject(gcpNfsInstance), gcpNfsInstance)
					if err != nil {
						return false, err
					}
					exists = gcpNfsInstance.Status.State == cloudcontrolv1beta1.StateReady
					return exists, nil
				}, timeout, interval).Should(BeTrue(), "expected NfsInstance for GCP with Ready state")
			})
			Eventually(func() (exists bool, err error) {
				err = infra.KCP().Client().Get(infra.Ctx(), client.ObjectKeyFromObject(gcpNfsInstance), gcpNfsInstance)
				if err != nil {
					return false, err
				}
				exists = gcpNfsInstance.Status.CapacityGb == 2*DefaultGcpNfsInstanceCapacityGb
				return exists, nil
			}, timeout, interval).Should(BeTrue(), "expected NfsInstance for GCP with updated capacity")
			By("And KCP NfsVolume has correct status.Capacity", func() {
				Expect(gcpNfsInstance.Status.CapacityGb).To(Equal(2 * DefaultGcpNfsInstanceCapacityGb))
				Expect(gcpNfsInstance.Status.Capacity).To(BeComparableTo(resource.MustParse(fmt.Sprintf("%dGi", 2*DefaultGcpNfsInstanceCapacityGb))))
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
			By("And NfsVolume keeps Ready state", func() {
				Expect(meta.IsStatusConditionTrue(gcpNfsInstance.Status.Conditions, cloudcontrolv1beta1.ConditionTypeReady)).To(BeTrue())
			})
			By("And eventually KCP NfsVolume will get permanently removed", func() {
				Eventually(func() (exists bool, err error) {
					err = infra.KCP().Client().Get(infra.Ctx(), client.ObjectKeyFromObject(gcpNfsInstance), gcpNfsInstance)
					exists = apierrors.IsNotFound(err)
					return exists, client.IgnoreNotFound(err)
				}, timeout, interval).Should(BeTrue(), "expected NfsInstance for GCP to be deleted")
			})
		})
	})

})

func sample500Error() *googleapi.Error {
	return &googleapi.Error{Code: 500, Message: "Internal Server Error"}
}
