package cloudcontrol

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	awsmock "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/mock"
	awsutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/util"
	"github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	"github.com/kyma-project/cloud-manager/pkg/util/debugged"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

var _ = Describe("IpRange AWS", func() {

	const (
		kymaName = "d87cfa6d-ff74-47e9-a3f6-c6efc637ce2a"
		vpcId    = "b1d68fc4-1bd4-4ad6-b81c-3d86de54f4f9"

		duration = time.Second * 5
		interval = time.Millisecond * 250
	)
	var (
		timeout = time.Second * 5
	)

	if debugged.Debugged {
		timeout = time.Minute * 20
	}

	It("IpRange AWS", func() {

		// Tell Scope reconciler to ignore this kymaName
		scope.Ignore.AddName(kymaName)

		// Given Scope exists
		Expect(
			infra.GivenScopeAwsExists(kymaName),
		).NotTo(HaveOccurred())

		// Load created scope
		scope := &cloudcontrolv1beta1.Scope{}
		Eventually(func() (exists bool, err error) {
			err = infra.KCP().Client().Get(infra.Ctx(), infra.KCP().ObjKey(kymaName), scope)
			exists = err == nil
			return exists, client.IgnoreNotFound(err)
		}, timeout, interval).
			Should(BeTrue(), "expected Scope to get created")

		// Given AWS VPC exists for this kymaName
		infra.AwsMock().AddVpc(vpcId, "10.180.0.0/16", awsutil.Ec2Tags("Name", scope.Spec.Scope.Aws.VpcNetwork), awsmock.VpcSubnetsFromScope(scope))

		// When IpRange is created
		iprange := &cloudcontrolv1beta1.IpRange{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: infra.KCP().Namespace(),
				Name:      "some-ip-range",
			},
			Spec: cloudcontrolv1beta1.IpRangeSpec{
				RemoteRef: cloudcontrolv1beta1.RemoteRef{
					Namespace: "skr-namespace",
					Name:      "skr-ip-range",
				},
				Scope: cloudcontrolv1beta1.ScopeRef{
					Name: kymaName,
				},
				Cidr: "10.181.0.0/16",
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

		Expect(iprange.Status.Cidr).To(Equal(iprange.Spec.Cidr), "expected IpRange status.cidr to be equal to spec.cidr")
		Expect(iprange.Status.Ranges).To(HaveLen(3), "expected three IpRange status.ranges")

		Expect(iprange.Status.Ranges).To(ContainElement("10.181.0.0/18"))
		Expect(iprange.Status.Ranges).To(ContainElement("10.181.64.0/18"))
		Expect(iprange.Status.Ranges).To(ContainElement("10.181.128.0/18"))

		Expect(iprange.Status.VpcId).To(Equal(vpcId))
		Expect(iprange.Status.Subnets).To(HaveLen(3))

		Expect(iprange.Status.Subnets).To(HaveLen(3))
		expectedZones := map[string]struct{}{
			"eu-west-1a": {},
			"eu-west-1b": {},
			"eu-west-1c": {},
		}
		for _, subnet := range iprange.Status.Subnets {
			Expect(subnet.Id).NotTo(BeEmpty())
			Expect(iprange.Status.Ranges).To(ContainElement(subnet.Range))
			Expect(expectedZones).To(HaveKey(subnet.Zone))
			delete(expectedZones, subnet.Zone)
		}

	})

})
