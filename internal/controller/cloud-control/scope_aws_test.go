package cloudresources

//
//import (
//	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/components/kcp/api/cloud-control/v1beta1"
//	"github.com/kyma-project/cloud-manager/components/kcp/pkg/util"
//	. "github.com/onsi/ginkgo/v2"
//	. "github.com/onsi/gomega"
//	"k8s.io/apimachinery/pkg/types"
//	"sigs.k8s.io/controller-runtime/pkg/client"
//	"time"
//)
//
//var _ = Describe("Scope AWS", func() {
//
//	const (
//		kymaName = "5d60be8c-e422-48ff-bd0a-166b0e09dc58"
//
//		timeout  = time.Second * 5
//		duration = time.Second * 5
//		interval = time.Millisecond * 250
//	)
//
//	It("Scope lifecycle", func() {
//		kcpObjKey := types.NamespacedName{
//			Namespace: infra.KCP().Namespace(),
//			Name:      kymaName,
//		}
//
//		Expect(infra.GivenGardenShootAwsExists(kymaName)).
//			NotTo(HaveOccurred(), "failed creating garden shoot for aws")
//
//		Expect(infra.GivenKymaCRExists(kymaName)).
//			NotTo(HaveOccurred(), "failed creating kyma cr")
//
//		scope := &cloudcontrolv1beta1.Scope{}
//
//		Consistently(func() (exists bool, err error) {
//			err = infra.KCP().Client().Get(infra.Ctx(), kcpObjKey, scope)
//			exists = err == nil
//			return exists, client.IgnoreNotFound(err)
//		}, duration, interval).
//			Should(BeFalse(), "expected Scope not to exist")
//
//		Expect(infra.WhenKymaModuleStateUpdates(kymaName, util.KymaModuleStateReady)).
//			NotTo(HaveOccurred())
//
//		Eventually(func() (exists bool, err error) {
//			err = infra.KCP().Client().Get(infra.Ctx(), kcpObjKey, scope)
//			exists = err == nil
//			return exists, client.IgnoreNotFound(err)
//		}, timeout, interval).
//			Should(BeTrue(), "expected Scope to be created")
//
//		Expect(scope).NotTo(BeNil())
//		Expect(scope.Spec.Provider).To(Equal(string(cloudcontrolv1beta1.ProviderAws)))
//	})
//
//})
