package cloudcontrol

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/meta"
	kcpscope "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Feature: KCP IpRange for GCP - Refactored Implementation", func() {

	It("Scenario: KCP GCP IpRange lifecycle with CIDR (refactored)", func() {
		const (
			kymaName    = "cbd30a68-eecd-4d3f-8a9b-2e60ed609115"
			ipRangeName = "b9446f79-70f1-443f-bde4-adfa7f1b2d0d"
		)

		scope := &cloudcontrolv1beta1.Scope{}
		ipRange := &cloudcontrolv1beta1.IpRange{}

		By("Given Scope exists", func() {
			kcpscope.Ignore.AddName(kymaName)

			Eventually(CreateScopeGcp).
				WithArguments(infra.Ctx(), infra, scope, WithName(kymaName)).
				Should(Succeed())
		})

		var kcpNetworkKyma *cloudcontrolv1beta1.Network

		By("And Given KCP Kyma Network exists in Ready state", func() {
			kcpNetworkKyma = cloudcontrolv1beta1.NewNetworkBuilder().
				WithScope(kymaName).
				WithName(common.KcpNetworkKymaCommonName(kymaName)).
				WithGcpRef(scope.Spec.Scope.Gcp.Project, scope.Spec.Scope.Gcp.VpcNetwork).
				WithType(cloudcontrolv1beta1.NetworkTypeKyma).
				Build()

			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpNetworkKyma).
				Should(Succeed())

			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpNetworkKyma, NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady)).
				Should(Succeed())
		})

		By("When KCP IpRange is created with refactored implementation", func() {
			// Create context with ipRangeRefactored feature flag enabled
			ctx := feature.ContextBuilderFromCtx(infra.Ctx()).Custom("ipRangeRefactored", true).Build(infra.Ctx())

			Eventually(CreateKcpIpRange).
				WithArguments(ctx, infra.KCP().Client(), ipRange,
					WithName(ipRangeName),
					WithKcpIpRangeSpecCidr("10.250.0.0/22"),
					WithKcpIpRangeRemoteRef(ipRangeName),
					WithScope(kymaName),
				).
				Should(Succeed())
		})

		By("Then KCP IpRange has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), ipRange,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
				).
				Should(Succeed())
		})

		By("And Then KCP IpRange has status.id", func() {
			Expect(ipRange.Status.Id).NotTo(BeEmpty())
		})

		By("And Then KCP IpRange has status.cidr equal to spec.cidr", func() {
			Expect(ipRange.Status.Cidr).To(Equal(ipRange.Spec.Cidr))
		})

		By("And Then GCP global address is created", func() {
			Eventually(func() error {
				_, err := infra.GcpMock().GetIpRangeDiscovery(infra.Ctx(), scope.Spec.Scope.Gcp.Project, "cm-"+ipRange.Name)
				return err
			}).Should(Succeed())
		})

		By("And Then GCP global address has correct properties", func() {
			gcpGlobalAddress, err := infra.GcpMock().GetIpRangeDiscovery(infra.Ctx(), scope.Spec.Scope.Gcp.Project, "cm-"+ipRange.Name)
			Expect(err).NotTo(HaveOccurred())
			Expect(gcpGlobalAddress).NotTo(BeNil())
			Expect(gcpGlobalAddress.Name).To(Equal("cm-" + ipRange.Name))
			Expect(gcpGlobalAddress.AddressType).To(Equal(string(gcpclient.AddressTypeInternal)))
		})

		// DELETE

		By("When KCP IpRange is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), ipRange).
				Should(Succeed())
		})

		By("Then KCP IpRange does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), ipRange).
				Should(Succeed())
		})

		By("And Then GCP global address does not exist", func() {
			gcpGlobalAddress, err := infra.GcpMock().GetIpRangeDiscovery(infra.Ctx(), scope.Spec.Scope.Gcp.Project, "cm-"+ipRange.Name)
			Expect(gcpGlobalAddress).To(BeNil())
			Expect(gcpmeta.IsNotFound(err)).To(BeTrue())
		})
	})
})
