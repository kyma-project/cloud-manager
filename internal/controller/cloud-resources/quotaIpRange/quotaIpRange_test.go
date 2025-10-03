package quotaIpRange

import (
	"fmt"
	"time"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/skr/iprange"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Feature: Quota SKR IpRange resource total count", func() {

	cleanup := func(nameList ...string) {
		Eventually(func() error {
			for _, name := range nameList {
				iprange.Ignore.AddName(name)
				skrIpRange := &cloudresourcesv1beta1.IpRange{}
				nsname := types.NamespacedName{
					Namespace: DefaultSkrNamespace,
					Name:      name,
				}
				if err := infra.SKR().Client().Get(infra.Ctx(), nsname, skrIpRange); err != nil {
					if apierrors.IsNotFound(err) {
						continue
					}
					return fmt.Errorf("error getting SKR IpRange %s: %w", name, err)
				}
				if len(skrIpRange.Status.Id) > 0 {
					kcpIpRange := &cloudcontrolv1beta1.IpRange{}
					if err := infra.KCP().Client().Get(infra.Ctx(), types.NamespacedName{
						Namespace: DefaultKcpNamespace,
						Name:      skrIpRange.Status.Id,
					}, kcpIpRange); client.IgnoreNotFound(err) != nil {
						return fmt.Errorf("error getting KCP IpRange for %s: %w", name, err)
					}
					if kcpIpRange.Name == skrIpRange.Status.Id {
						By(fmt.Sprintf("// cleanup: KCP IpRange deleting %s because of %s", kcpIpRange.Name, skrIpRange.Name), func() {})
						if err := infra.KCP().Client().Delete(infra.Ctx(), kcpIpRange); client.IgnoreNotFound(err) != nil {
							return fmt.Errorf("error deleting KCP IpRange for %s: %w", name, err)
						}
					}
				}
				if len(skrIpRange.Finalizers) > 0 {
					By(fmt.Sprintf("// cleanup: SKR IpRange removing finalizers %s", skrIpRange.Name), func() {})
					skrIpRange.Finalizers = nil
					if err := infra.SKR().Client().Update(infra.Ctx(), skrIpRange); client.IgnoreNotFound(err) != nil {
						return fmt.Errorf("error updating SKR IpRange %s: %w", name, err)
					}
					if err := infra.SKR().Client().Get(infra.Ctx(), nsname, skrIpRange); client.IgnoreNotFound(err) != nil {
						return fmt.Errorf("error getting SKR IpRange %s: %w", name, err)
					}
				}
				By(fmt.Sprintf("// cleanup: SKR IpRange deleting %s", skrIpRange.Name), func() {})
				if err := infra.SKR().Client().Delete(infra.Ctx(), skrIpRange); client.IgnoreNotFound(err) != nil {
					return fmt.Errorf("error deleting SKR IpRange %s: %w", name, err)
				}
			}

			return nil
		}).
			Should(Succeed(), "error cleaning up IpRanges")
	}

	setKcpIpRangeReadyAndWaitSkrIpRangeReady := func(skrIpRange *cloudresourcesv1beta1.IpRange) {
		kcpIpRange := &cloudcontrolv1beta1.IpRange{}

		Eventually(LoadAndCheck).
			WithArguments(
				infra.Ctx(),
				infra.SKR().Client(),
				skrIpRange,
				NewObjActions(),
				AssertSkrIpRangeHasId(),
			).
			Should(Succeed(), "expected first SKR IpRange to get status.id, but it didn't")

		Eventually(LoadAndCheck).
			WithArguments(
				infra.Ctx(),
				infra.KCP().Client(),
				kcpIpRange,
				NewObjActions(WithName(skrIpRange.Status.Id)),
			).
			Should(Succeed(), "expected first KCP IpRange to exists, but none found")

		Eventually(UpdateStatus).
			WithArguments(
				infra.Ctx(),
				infra.KCP().Client(),
				kcpIpRange,
				WithConditions(KcpReadyCondition()),
			).
			Should(Succeed(), "failed updating KCP IpRange status with Ready condition")
	}

	createSkrIpRangeWithReadyStatus := func(
		skrIpRangeName string,
		cidr string,
		skrIpRange *cloudresourcesv1beta1.IpRange,
	) {
		Eventually(CreateSkrIpRange).
			WithArguments(
				infra.Ctx(), infra.SKR().Client(),
				skrIpRange,
				WithName(skrIpRangeName),
				WithSkrIpRangeSpecCidr(cidr),
			).
			Should(Succeed(), "failed creating first SKR IpRange")

		setKcpIpRangeReadyAndWaitSkrIpRangeReady(skrIpRange)

		Eventually(LoadAndCheck).
			WithArguments(
				infra.Ctx(),
				infra.SKR().Client(),
				skrIpRange,
				NewObjActions(),
				HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
			).
			Should(Succeed(), "expected first SKR IpRange to has Ready condition, but it does not")
	}

	It("Scenario: Quota and IpRange CIDR overlap", func() {

		const (
			firstSkrIpRangeName  = "e4ff6f40-4015-4d08-9490-25e5e9a90e98"
			secondSkrIpRangeName = "287ddf66-388d-4ecc-b843-937082da44db"
			cidr                 = "10.1.0.0/24"
		)

		firstSkrIpRange := &cloudresourcesv1beta1.IpRange{}
		secondSkrIpRange := &cloudresourcesv1beta1.IpRange{}

		By("Given first SKR IpRange is created and has Ready condition", func() {
			createSkrIpRangeWithReadyStatus(firstSkrIpRangeName, cidr, firstSkrIpRange)
			time.Sleep(time.Second) // one second mast pass since obj.metadata.creationTimestamp has no nano part
		})

		By("When second SKR IpRange is created with same CIDR", func() {
			Eventually(CreateSkrIpRange).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), secondSkrIpRange,
					WithName(secondSkrIpRangeName),
					WithSkrIpRangeSpecCidr(cidr),
				).
				Should(Succeed(), "failed creating second SKR IpRange")
		})

		By("Then second SKR IpRange has Error condition with CidrOverlap reason", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					secondSkrIpRange,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeError),
				).
				Should(Succeed(), "expected second SKR IpRange to have error condition, but it does not")

			errCond := meta.FindStatusCondition(secondSkrIpRange.Status.Conditions, cloudresourcesv1beta1.ConditionTypeError)
			Expect(errCond).NotTo(BeNil())
			Expect(errCond.Reason).To(Equal(cloudresourcesv1beta1.ConditionReasonCidrOverlap))
		})

		By("When first SKR IpRange is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), firstSkrIpRange).
				Should(Succeed(), "failed deleting first SKR IpRange")
		})

		By("Then first SKR IpRange does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), firstSkrIpRange).
				Should(Succeed(), "expected first SKR IpRange not to exist anymore, but it does")
		})

		cleanup(firstSkrIpRangeName, secondSkrIpRangeName)
	})

	It("Scenario: Quota for IpRange allows older objects", func() {

		const (
			firstSkrIpRangeName  = "d8ba09b8-ef17-4199-9f49-ea83c7aee855"
			firstCidr            = "10.1.0.0/24"
			secondSkrIpRangeName = "80b9490c-b057-4f3c-8722-c21491487845"
			secondCidr           = "10.2.0.0/24"
			thirdSkrIpRangeName  = "65389b39-782d-48f8-bab3-2db7465df9c9"
			thirdCidr            = "10.3.0.0/24"
		)

		firstSkrIpRange := &cloudresourcesv1beta1.IpRange{}
		secondSkrIpRange := &cloudresourcesv1beta1.IpRange{}
		thirdSkrIpRange := &cloudresourcesv1beta1.IpRange{}

		By("Given first SKR IpRange is created and has Ready condition", func() {
			createSkrIpRangeWithReadyStatus(firstSkrIpRangeName, firstCidr, firstSkrIpRange)
			time.Sleep(time.Second)
		})

		By("When second SKR IpRange is created", func() {
			Eventually(CreateSkrIpRange).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(),
					secondSkrIpRange,
					WithName(secondSkrIpRangeName),
					WithSkrIpRangeSpecCidr(secondCidr),
				).
				Should(Succeed(), "failed creating second SKR IpRange")
			time.Sleep(time.Second)
		})

		By("Then second SKR IpRange has QuotaExceed condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					secondSkrIpRange,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeQuotaExceeded),
				).
				Should(Succeed(), "expected second SKR IpRange to have QuotaExceeded condition, but it does not")
			time.Sleep(time.Second)
		})

		By("When third SKR IpRange is created", func() {
			time.Sleep(100 * time.Millisecond)

			Eventually(CreateSkrIpRange).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(),
					thirdSkrIpRange,
					WithName(thirdSkrIpRangeName),
					WithSkrIpRangeSpecCidr(thirdCidr),
				).
				Should(Succeed(), "failed creating third SKR IpRange")
		})

		By("Then third SKR IpRange has QuotaExceed condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					thirdSkrIpRange,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeQuotaExceeded),
				).
				Should(Succeed(), "expected third SKR IpRange to have QuotaExceeded condition, but it does not")
		})

		By("When first SKR IpRange is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), firstSkrIpRange).
				Should(Succeed(), "failed deleting first SKR IpRange")
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), firstSkrIpRange).
				Should(Succeed(), "expected first SKR IpRange not to exist")
		})

		By("Then second SKR IpRange is ready", func() {
			setKcpIpRangeReadyAndWaitSkrIpRangeReady(secondSkrIpRange)

			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(),
					secondSkrIpRange,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
				).
				Should(Succeed(), "expected second SKR IpRange to have Ready condition")
		})

		By("When second SKR IpRange is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), secondSkrIpRange).
				Should(Succeed(), "failed deleting second SKR IpRange")
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), secondSkrIpRange).
				Should(Succeed(), "expected second SKR IpRange not to exist")
		})

		By("Then third SKR IpRange is ready", func() {
			setKcpIpRangeReadyAndWaitSkrIpRangeReady(thirdSkrIpRange)

			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(),
					thirdSkrIpRange,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
				).
				Should(Succeed(), "expected third SKR IpRange to have Ready condition")
		})

		cleanup(firstSkrIpRangeName, secondSkrIpRangeName, thirdSkrIpRangeName)
	})
})
