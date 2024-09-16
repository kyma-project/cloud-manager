package cloudcontrol

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	"github.com/kyma-project/cloud-manager/pkg/util/debugged"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

var _ = Describe("Feature: KCP IpRange for GCP", func() {

	const (
		kymaName = "a179cd2d-5cb6-42b0-86bc-bc8720bccbc8"

		interval = time.Millisecond * 250
	)
	var (
		timeout = time.Second * 20
	)

	if debugged.Debugged {
		timeout = time.Minute * 20
	}

	It("Scenario: IpRange GCP Creation", func() {

		// Tell Scope reconciler to ignore this kymaName
		scope.Ignore.AddName(kymaName)

		// Given Scope exists
		Expect(
			infra.GivenScopeGcpExists(kymaName),
		).NotTo(HaveOccurred())
		time.Sleep(time.Second)

		// Load created scope
		scope := &cloudcontrolv1beta1.Scope{}
		Eventually(func() (exists bool, err error) {
			err = infra.KCP().Client().Get(infra.Ctx(), infra.KCP().ObjKey(kymaName), scope)
			exists = err == nil
			return exists, client.IgnoreNotFound(err)
		}, timeout, interval).
			Should(BeTrue(), "expected Scope to get created")

		// When IpRange is created
		iprange := &cloudcontrolv1beta1.IpRange{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: infra.KCP().Namespace(),
				Name:      "gcp-ip-range",
			},
			Spec: cloudcontrolv1beta1.IpRangeSpec{
				RemoteRef: cloudcontrolv1beta1.RemoteRef{
					Namespace: "skr-namespace",
					Name:      "skr-ip-range",
				},
				Scope: cloudcontrolv1beta1.ScopeRef{
					Name: kymaName,
				},
				Cidr: "10.20.30.0/24",
			},
		}
		Expect(
			infra.KCP().Client().Create(infra.Ctx(), iprange),
		).NotTo(HaveOccurred())

		// Then IpRange will get Ready condition
		Eventually(func() (exists bool, err error) {
			err = infra.KCP().Client().Get(infra.Ctx(), client.ObjectKeyFromObject(iprange), iprange)
			if err != nil {
				return false, err
			}
			exists = meta.IsStatusConditionTrue(iprange.Status.Conditions, cloudcontrolv1beta1.ConditionTypeReady)
			return exists, nil
		}, timeout, interval).
			Should(BeTrue(), "expected IpRange with Ready condition")

		Expect(iprange.Status.Cidr).To(Equal(iprange.Spec.Cidr))
		Expect(iprange.Status.State).To(Equal(cloudcontrolv1beta1.ReadyState))

	})

	It("Scenario: IpRange GCP Deletion", func() {

		// Tell Scope reconciler to ignore this kymaName
		scope.Ignore.AddName(kymaName)

		// Given Scope exists
		Expect(
			infra.GivenScopeGcpExists(kymaName),
		).NotTo(HaveOccurred())
		time.Sleep(time.Second)

		// Load created scope
		scope := &cloudcontrolv1beta1.Scope{}
		Eventually(func() (exists bool, err error) {
			err = infra.KCP().Client().Get(infra.Ctx(), infra.KCP().ObjKey(kymaName), scope)
			exists = err == nil
			return exists, client.IgnoreNotFound(err)
		}, timeout, interval).
			Should(BeTrue(), "expected Scope to get created")

		// Check IpRange exists
		iprange := &cloudcontrolv1beta1.IpRange{}
		Eventually(func() (exists bool, err error) {
			err = infra.KCP().Client().Get(infra.Ctx(), infra.KCP().ObjKey("gcp-ip-range"), iprange)
			exists = err == nil
			return exists, client.IgnoreNotFound(err)
		}, timeout, interval).Should(BeTrue(), "expected IpRange to alreay exist")
		// When IpRange is created
		Eventually(CreateKcpIpRange).
			WithArguments(
				infra.Ctx(), infra.KCP().Client(), iprange,
				WithName("gcp-ip-range"),
				WithNamespace(infra.KCP().Namespace()),
				WithKcpIpRangeRemoteRef("skr-ip-range"),
				WithScope(kymaName),
				WithKcpIpRangeSpecCidr("10.20.30.0/24"),
			).
			Should(Succeed())

		// Then IpRange will get Ready condition
		Eventually(func() (exists bool, err error) {
			err = infra.KCP().Client().Get(infra.Ctx(), client.ObjectKeyFromObject(iprange), iprange)
			if err != nil {
				return false, err
			}
			exists = meta.IsStatusConditionTrue(iprange.Status.Conditions, cloudcontrolv1beta1.ConditionTypeReady)
			return exists, nil
		}, timeout, interval).
			Should(BeTrue(), "expected IpRange with Ready condition")

		Expect(iprange.Status.Cidr).To(Equal(iprange.Spec.Cidr))
		Expect(iprange.Status.State).To(Equal(cloudcontrolv1beta1.ReadyState))

		By("When KCP IpRange is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), iprange).
				Should(Succeed(), "failed deleting KCP IpRange")
		})

		By("Then KCP IpRange does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), iprange).
				Should(Succeed(), "expected KCP IpRange to be deleted, but it exists")
		})

	})

	It("Scenario: IpRange CIDR is automatically allocated", func() {
		// Tell Scope reconciler to ignore this kymaName
		scope.Ignore.AddName(kymaName)

		// Given Scope exists
		Expect(
			infra.GivenScopeGcpExists(kymaName),
		).NotTo(HaveOccurred())
		time.Sleep(time.Second)

		// Load created scope
		scope := &cloudcontrolv1beta1.Scope{}
		Eventually(func() (exists bool, err error) {
			err = infra.KCP().Client().Get(infra.Ctx(), infra.KCP().ObjKey(kymaName), scope)
			exists = err == nil
			return exists, client.IgnoreNotFound(err)
		}, timeout, interval).
			Should(BeTrue(), "expected Scope to get created")

		// When IpRange is created without cidr
		iprange := &cloudcontrolv1beta1.IpRange{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: infra.KCP().Namespace(),
				Name:      "gcp-ip-range",
			},
			Spec: cloudcontrolv1beta1.IpRangeSpec{
				RemoteRef: cloudcontrolv1beta1.RemoteRef{
					Namespace: "skr-namespace",
					Name:      "skr-ip-range",
				},
				Scope: cloudcontrolv1beta1.ScopeRef{
					Name: kymaName,
				},
				Cidr: "",
			},
		}
		Expect(
			infra.KCP().Client().Create(infra.Ctx(), iprange),
		).NotTo(HaveOccurred())

		// Then IpRange will get Ready condition
		Eventually(func() (exists bool, err error) {
			err = infra.KCP().Client().Get(infra.Ctx(), client.ObjectKeyFromObject(iprange), iprange)
			if err != nil {
				return false, err
			}
			exists = meta.IsStatusConditionTrue(iprange.Status.Conditions, cloudcontrolv1beta1.ConditionTypeReady)
			return exists, nil
		}, timeout, interval).
			Should(BeTrue(), "expected IpRange with Ready condition")

		Expect(len(iprange.Status.Cidr)).To(BeNumerically(">", 0))
		Expect(iprange.Status.State).To(Equal(cloudcontrolv1beta1.ReadyState))

	})

})
