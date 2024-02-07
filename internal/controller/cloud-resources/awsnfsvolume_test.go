package cloudresources

import (
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	skriprange "github.com/kyma-project/cloud-manager/pkg/skr/iprange"
	"github.com/kyma-project/cloud-manager/pkg/util/debugged"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

var _ = Describe("KCP AwsNfsVolume", Focus, func() {

	const (
		awsNfsVolumeName = "aws-nfs-volume1"
		skrNamespace     = "test"

		duration = time.Second * 5
		interval = time.Millisecond * 250
	)

	var (
		timeout = time.Second * 10
	)

	if debugged.Debugged {
		timeout = time.Minute * 20
	}

	It("AwsNfsVolume creation", func() {

		Expect(infra.SKR().GivenNamespaceExists(skrNamespace)).
			NotTo(HaveOccurred())

		skriprange.Ignore.AddName(awsNfsVolumeName)

		Expect(infra.GivenSkrIpRangeExists(
			infra.Ctx(),
			skrNamespace,
			awsNfsVolumeName,
			"10.181.0.0/16",
			metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeReady,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionTypeReady,
				Message: "Ready",
			},
		)).
			NotTo(HaveOccurred())

		// When AwsNfsVolume is created
		nfsVol := &cloudresourcesv1beta1.AwsNfsVolume{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: skrNamespace,
				Name:      awsNfsVolumeName,
			},
			Spec: cloudresourcesv1beta1.AwsNfsVolumeSpec{
				IpRange: cloudresourcesv1beta1.IpRangeRef{
					Namespace: skrNamespace,
					Name:      awsNfsVolumeName,
				},
				PerformanceMode: "",
				Throughput:      "",
			},
		}
		Expect(infra.SKR().Client().Create(infra.Ctx(), nfsVol)).
			NotTo(HaveOccurred())

		// When SKR source is reloaded and SKR loop triggered
		Expect(infra.SkrRunner().Reloader().ReloadObjKind(infra.Ctx(), &cloudresourcesv1beta1.IpRange{})).
			NotTo(HaveOccurred())
		Expect(infra.SkrRunner().Reloader().ReloadObjKind(infra.Ctx(), &cloudresourcesv1beta1.AwsNfsVolume{})).
			NotTo(HaveOccurred())

		// Then AwsNfsVolume will eventually get status condition
		var readyCond *metav1.Condition
		var errorCond *metav1.Condition
		Eventually(func() (ok bool, err error) {
			err = infra.SKR().Client().Get(infra.Ctx(), client.ObjectKeyFromObject(nfsVol), nfsVol)
			if err != nil {
				return false, client.IgnoreNotFound(err)
			}

			readyCond = meta.FindStatusCondition(*nfsVol.Conditions(), cloudresourcesv1beta1.ConditionTypeReady)
			errorCond = meta.FindStatusCondition(*nfsVol.Conditions(), cloudresourcesv1beta1.ConditionTypeError)
			return len(*nfsVol.Conditions()) > 0, nil
		}, timeout, interval).
			Should(BeTrue(), "expected SKR AwsNfsVolume to be loaded and have a status condition")

		Expect(errorCond).To(BeNil(), "expected SKR AwsNfsVolume not to have Error condition")
		Expect(readyCond).NotTo(BeNil(), "expected SKR AwsNfsVolume to have Ready condition")
	})

})
