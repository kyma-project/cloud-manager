package cloudcontrol

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	scopePkg "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var _ = Describe("Feature: KCP Network reference", func() {

	It("Scenario: Network reference is created and deleted", func() {

		kymaName := "0b10bb9c-a727-4f5b-8d64-803110402586"
		scope := &cloudcontrolv1beta1.Scope{}
		netObjName := "bacd3153-ecd4-4666-86de-067f94aaae53"
		var net *cloudcontrolv1beta1.Network

		awsAccount := "acc-123"
		awsRegion := "us-east-1"
		netId := "net-345"
		netName := "my-net"

		By("Given Scope exists", func() {
			scopePkg.Ignore.AddName(kymaName)
			Eventually(CreateScopeAws).
				WithArguments(infra.Ctx(), infra, scope, WithName(kymaName)).
				Should(Succeed())
		})

		By("When Network reference is created", func() {
			net = cloudcontrolv1beta1.NewNetworkBuilder().WithAwsRef(awsAccount, awsRegion, netId, netName).Build()
			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.KCP().Client(), net, WithName(netObjName), WithScope(kymaName)).
				Should(Succeed())
		})

		By("Then Network state is Ready", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), net, NewObjActions(), HavingState(string(cloudcontrolv1beta1.ReadyState))).
				Should(Succeed())
		})

		By("And Then Network has Ready condition", func() {
			cond := meta.FindStatusCondition(*net.Conditions(), cloudcontrolv1beta1.ConditionTypeReady)
			Expect(cond).ToNot(BeNil())
			Expect(cond.Status).To(Equal(metav1.ConditionTrue))
		})

		By("And Then Network reference in status equals to its spec value", func() {
			Expect(net.Status.Network).To(Not(BeNil()))
			Expect(net.Spec.Network.Reference.Equals(net.Status.Network)).To(BeTrue())
		})

		By("And Then Network has finalizer", func() {
			Expect(controllerutil.ContainsFinalizer(net, cloudcontrolv1beta1.FinalizerName)).To(BeTrue())
		})
	})

})
