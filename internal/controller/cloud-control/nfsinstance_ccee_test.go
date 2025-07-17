package cloudcontrol

import (
	"fmt"

	"github.com/gophercloud/gophercloud/v2/openstack/sharedfilesystems/v2/shares"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	kcpscope "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Feature: KCP NfsInstance SAP", func() {

	It("Scenario: KCP SAP NfsInstance is created and deleted", func() {
		name := "f7db16a8-0fb4-4f9d-b055-e360a10e1f36"
		scope := &cloudcontrolv1beta1.Scope{}

		By("Given OpenStack Scope exists", func() {
			// Tell Scope reconciler to ignore this Scope
			kcpscope.Ignore.AddName(name)

			Eventually(CreateScopeOpenStack).
				WithArguments(infra.Ctx(), infra, scope, WithName(name)).
				Should(Succeed(), "failed creating Scope")
		})

		networkId := "273c6391-b934-4d14-9ebc-2da48c364bf4"
		subnetId := "031a4b51-146b-455f-9243-770b791b1b28"

		By("And Given SAP network exists", func() {
			infra.SapMock().AddNetwork(
				"wrong1",
				"wrong1",
			)
			infra.SapMock().AddNetwork(
				networkId,
				scope.Spec.Scope.OpenStack.VpcNetwork,
			)
			infra.SapMock().AddNetwork(
				"wrong2",
				"wrong2",
			)

			infra.SapMock().AddSubnet(
				subnetId,
				networkId,
				scope.Spec.Scope.OpenStack.VpcNetwork,
				"10.250.0.0/22",
			)
		})

		nfsInstance := &cloudcontrolv1beta1.NfsInstance{}

		By("When NfsInstance is created", func() {
			Eventually(CreateNfsInstance).
				WithArguments(infra.Ctx(), infra.KCP().Client(), nfsInstance,
					WithName(name),
					WithRemoteRef("foo"),
					WithScope(name),
					WithNfsInstanceSap(10),
				).
				Should(Succeed(), "failed creating NfsInstance")
		})

		var theShare *shares.Share

		By("Then SAP share is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), nfsInstance,
					NewObjActions(),
					HavingNfsInstanceStatusId()).
				Should(Succeed(), "expected NfsInstance to get status.id")

			x, err := infra.SapMock().GetShare(infra.Ctx(), nfsInstance.Status.Id)
			Expect(err).NotTo(HaveOccurred())
			theShare = x
		})

		By("When Share is available", func() {
			infra.SapMock().SetShareStatus(theShare.ID, "available")
		})

		By("Then NfsInstance has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), nfsInstance,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
				).
				Should(Succeed(), "expected NfsInstance to have Ready state, but it didn't")

			// reload share
			x, err := infra.SapMock().GetShare(infra.Ctx(), nfsInstance.Status.Id)
			Expect(err).NotTo(HaveOccurred())
			Expect(x).NotTo(BeNil())
			theShare = x
		})

		By("And Then NfsInstance has status.host set", func() {
			Expect(nfsInstance.Status.Path).To(Equal(fmt.Sprintf("%s-1", theShare.ID)))
			Expect(nfsInstance.Status.Host).To(Equal("10.100.0.10"))
		})

		By("And Then NfsInstance has status.capacity set", func() {
			Expect(nfsInstance.Status.Capacity.String()).To(Equal("10Gi"))
		})

		By("And Then Share has access granted", func() {
			arr, err := infra.SapMock().ListShareAccessRules(infra.Ctx(), theShare.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(arr).To(HaveLen(1), "expected one access right")
			Expect(arr[0].AccessTo).To(Equal(scope.Spec.Scope.OpenStack.Network.Nodes))
			Expect(arr[0].AccessLevel).To(Equal("rw"))
			Expect(arr[0].AccessType).To(Equal("ip"))
		})

		// DELETE

		By("When NfsInstance is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), nfsInstance).
				Should(Succeed(), "failed deleting NfsInstance")
		})

		By("Then NfsInstance does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), nfsInstance).
				Should(Succeed(), "expected NfsInstance not to exist (be deleted), but it still exists")
		})

		By("And Then SAP Share does not exist", func() {
			x, err := infra.SapMock().GetShare(infra.Ctx(), theShare.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(x).To(BeNil())
		})

		By("And Then SAP Share Network does not exist", func() {
			x, err := infra.SapMock().GetShareNetwork(infra.Ctx(), theShare.ShareNetworkID)
			Expect(err).NotTo(HaveOccurred())
			Expect(x).To(BeNil())
		})
	})

})
