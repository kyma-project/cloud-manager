package cloudresources

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/util/debugged"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"time"
)

var _ = Describe("IpRange", func() {

	const (
		ipRangeName      = "range1"
		ipRangeNamespace = "test"

		duration = time.Second * 5
		interval = time.Millisecond * 250
	)

	var (
		timeout = time.Second * 5
	)

	if debugged.Debugged {
		timeout = time.Minute * 20
	}

	It("IpRange lifecycle", func() {

		Expect(infra.SKR().GivenNamespaceExists(ipRangeNamespace)).
			NotTo(HaveOccurred())

		// When SKR IpRange is created
		skrIpRange := &cloudresourcesv1beta1.IpRange{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ipRangeNamespace,
				Name:      ipRangeName,
			},
			Spec: cloudresourcesv1beta1.IpRangeSpec{
				Cidr: "10.181.0.0/16",
			},
		}
		Expect(infra.SKR().Client().Create(infra.Ctx(), skrIpRange)).
			NotTo(HaveOccurred())

		// When SKR source is reloaded and SKR loop triggered
		Expect(infra.SkrRunner().Reloader().ReloadObjKind(infra.Ctx(), &cloudresourcesv1beta1.IpRange{})).
			NotTo(HaveOccurred())

		// Then Kcp IpRage will get created
		kcpIpRangeList := &cloudcontrolv1beta1.IpRangeList{}
		Eventually(func() (exists bool, err error) {
			err = infra.KCP().Client().List(
				infra.Ctx(),
				kcpIpRangeList,
				client.MatchingLabels{
					cloudcontrolv1beta1.LabelKymaName:        infra.SkrKymaRef().Name,
					cloudcontrolv1beta1.LabelRemoteName:      skrIpRange.Name,
					cloudcontrolv1beta1.LabelRemoteNamespace: skrIpRange.Namespace,
				},
			)
			exists = len(kcpIpRangeList.Items) > 0
			return
		}, timeout, interval).
			Should(BeTrue(), "expected KCP IpRange to be created")
		Expect(kcpIpRangeList.Items).
			To(HaveLen(1))

		// Then Assert KCP IpRange =====================================================
		kcpIpRange := &kcpIpRangeList.Items[0]
		// has labels
		Expect(kcpIpRange.Labels).To(HaveKeyWithValue(cloudcontrolv1beta1.LabelKymaName, infra.SkrKymaRef().Name))
		Expect(kcpIpRange.Labels).To(HaveKeyWithValue(cloudcontrolv1beta1.LabelRemoteName, skrIpRange.Name))
		Expect(kcpIpRange.Labels).To(HaveKeyWithValue(cloudcontrolv1beta1.LabelRemoteNamespace, skrIpRange.Namespace))

		// has scope specified with name equal to SKR's kymaRef
		Expect(kcpIpRange.Spec.Scope.Name).To(Equal(infra.SkrKymaRef().Name))

		// has cidr same as in SKR
		Expect(kcpIpRange.Spec.Cidr).To(Equal(skrIpRange.Spec.Cidr))

		// has remote ref set with SKR IpRange name
		Expect(kcpIpRange.Spec.RemoteRef.Name).To(Equal(skrIpRange.Name))
		Expect(kcpIpRange.Spec.RemoteRef.Namespace).To(Equal(skrIpRange.Namespace))

		// Then Assert SKR IpRange =====================================================
		Expect(infra.SKR().Client().Get(infra.Ctx(), client.ObjectKeyFromObject(skrIpRange), skrIpRange)).
			NotTo(HaveOccurred())
		// has cidr copy in status
		Expect(skrIpRange.Status.Cidr).To(Equal(skrIpRange.Spec.Cidr))
		// has finalizer
		Expect(controllerutil.ContainsFinalizer(skrIpRange, cloudresourcesv1beta1.Finalizer))
		// does not have ready condition type, since KCP IpRange has none so far
		Expect(meta.FindStatusCondition(skrIpRange.Status.Conditions, cloudresourcesv1beta1.ConditionTypeReady)).To(BeNil())
		// does not have error condition type, since KCP IpRange has none so far
		Expect(meta.FindStatusCondition(skrIpRange.Status.Conditions, cloudresourcesv1beta1.ConditionTypeError)).To(BeNil())
	})
})
