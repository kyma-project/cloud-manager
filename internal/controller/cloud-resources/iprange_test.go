package cloudresources

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/util/debugged"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	})
})
