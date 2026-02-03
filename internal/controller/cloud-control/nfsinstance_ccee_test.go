package cloudcontrol

import (
	"fmt"

	"github.com/gophercloud/gophercloud/v2/openstack/sharedfilesystems/v2/shares"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	kcpiprange "github.com/kyma-project/cloud-manager/pkg/kcp/iprange"
	kcpscope "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Feature: KCP NfsInstance SAP", func() {

	It("Scenario: KCP SAP NfsInstance is created and deleted", func() {
		name := "f7db16a8-0fb4-4f9d-b055-e360a10e1f36"
		scope := &cloudcontrolv1beta1.Scope{}

		sapMock := infra.SapMock().NewProject()

		By("Given OpenStack Scope exists", func() {
			// Tell Scope reconciler to ignore this Scope
			kcpscope.Ignore.AddName(name)

			Expect(CreateScopeOpenStack(infra.Ctx(), infra, scope, sapMock.ProviderParams(), WithName(name))).
				To(Succeed(), "failed creating Scope")
		})

		//var sapGardenInfra *SapGardenerInfra

		By("And Given SAP infra exists", func() {
			sapMock.AddNetwork(
				"wrong1-"+name,
				"wrong1-"+name,
			)

			_, err := CreateSapGardenerResources(infra.Ctx(), sapMock, infra.Garden().Namespace(), scope.Spec.ShootName, "10.250.0.0/16")
			Expect(err).NotTo(HaveOccurred())
			//sapGardenInfra = sgi

			sapMock.AddNetwork(
				"wrong2-"+name,
				"wrong2-"+name,
			)
		})

		var kcpNetworkKyma *cloudcontrolv1beta1.Network

		By("And Given KCP Kyma Network exists in Ready state", func() {
			kcpNetworkKyma = cloudcontrolv1beta1.NewNetworkBuilder().
				WithScope(name).
				WithName(common.KcpNetworkKymaCommonName(name)).
				WithOpenStackRef(scope.Spec.Scope.OpenStack.DomainName, scope.Spec.Scope.OpenStack.TenantName, "", scope.Spec.Scope.OpenStack.VpcNetwork).
				WithType(cloudcontrolv1beta1.NetworkTypeKyma).
				Build()

			Expect(CreateObj(infra.Ctx(), infra.KCP().Client(), kcpNetworkKyma)).
				To(Succeed())

			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpNetworkKyma, NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady)).
				Should(Succeed())
		})

		kcpIpRange := &cloudcontrolv1beta1.IpRange{}
		iprangeCidr := "10.251.0.0/16"

		By("And Given KCP IpRange exists in Ready state", func() {
			kcpiprange.Ignore.AddName(name)

			Expect(CreateKcpIpRange(infra.Ctx(), infra.KCP().Client(), kcpIpRange,
				WithName(name),
				WithKcpIpRangeRemoteRef("some-remote-ref"),
				WithKcpIpRangeNetwork(kcpNetworkKyma.Name),
				WithKcpIpRangeSpecCidr(iprangeCidr),
				WithScope(name),
			)).
				To(Succeed())

			Expect(UpdateStatus(infra.Ctx(), infra.KCP().Client(), kcpIpRange,
				WithKcpIpRangeStatusCidr(iprangeCidr),
				WithConditions(KcpReadyCondition()),
			)).
				To(Succeed())
		})

		nfsInstance := &cloudcontrolv1beta1.NfsInstance{}

		By("When NfsInstance is created", func() {
			Eventually(CreateNfsInstance).
				WithArguments(infra.Ctx(), infra.KCP().Client(), nfsInstance,
					WithName(name),
					WithRemoteRef("foo"),
					WithScope(name),
					WithIpRange(kcpIpRange.Name),
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

			x, err := sapMock.GetShare(infra.Ctx(), nfsInstance.Status.Id)
			Expect(err).NotTo(HaveOccurred())
			theShare = x
		})

		By("When Share is available", func() {
			sapMock.SetShareStatus(theShare.ID, "available")
		})

		By("Then NfsInstance has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), nfsInstance,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
				).
				Should(Succeed(), "expected NfsInstance to have Ready state, but it didn't")

			// reload share
			x, err := sapMock.GetShare(infra.Ctx(), nfsInstance.Status.Id)
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
			arr, err := sapMock.ListShareAccessRules(infra.Ctx(), theShare.ID)
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
			x, err := sapMock.GetShare(infra.Ctx(), theShare.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(x).To(BeNil())
		})

		By("And Then SAP Share Network does not exist", func() {
			x, err := sapMock.GetShareNetwork(infra.Ctx(), theShare.ShareNetworkID)
			Expect(err).NotTo(HaveOccurred())
			Expect(x).To(BeNil())
		})

		By("// cleanup: delete KCP IpRange", func() {
			Expect(Delete(infra.Ctx(), infra.KCP().Client(), kcpIpRange)).
				To(Succeed())
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpIpRange).
				Should(Succeed())
		})

		By("// cleanup: delete Scope", func() {
			Expect(Delete(infra.Ctx(), infra.KCP().Client(), scope)).
				To(Succeed())
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), scope).
				Should(Succeed())
		})
	})

})
