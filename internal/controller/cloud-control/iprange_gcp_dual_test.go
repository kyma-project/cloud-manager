package cloudcontrol

import (
	"strconv"
	"strings"

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

var _ = Describe("Feature: KCP IpRange for GCP - Dual Implementation Testing", func() {

	// Helper to test IpRange lifecycle with specific implementation
	testIpRangeLifecycleWithCidr := func(useRefactored bool, implName string) {
		const (
			kymaName    = "570c0d27-d67a-44cc-908a-2c151d50303e"
			ipRangeName = "81dbff76-33ce-4a2b-be48-a48ea4f2a25b"
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

		By("When KCP IpRange is created", func() {
			// Set feature flag context before creating IpRange
			ctx := infra.Ctx()
			if useRefactored {
				ctx = feature.ContextBuilderFromCtx(ctx).Custom("ipRangeRefactored", true).Build(ctx)
			} else {
				ctx = feature.ContextBuilderFromCtx(ctx).Custom("ipRangeRefactored", false).Build(ctx)
			}

			Eventually(CreateKcpIpRange).
				WithArguments(ctx, infra.KCP().Client(), ipRange,
					WithName(ipRangeName),
					WithKcpIpRangeRemoteRef(ipRangeName),
					WithKcpIpRangeSpecCidr("10.20.30.0/24"),
					WithScope(kymaName),
				).
				Should(Succeed())
		})

		By("Then KCP IpRange has Ready condition ("+implName+")", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), ipRange,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
				).
				Should(Succeed())
		})

		By("And Then KCP IpRange has status cidr equals to spec cidr", func() {
			Expect(ipRange.Status.Cidr).To(Equal(ipRange.Spec.Cidr))
		})

		By("And Then GCP global address is created ("+implName+")", func() {
			gcpGlobalAddress, err := infra.GcpMock().GetIpRange(infra.Ctx(), scope.Spec.Scope.Gcp.Project, "cm-"+ipRange.Name)
			Expect(err).NotTo(HaveOccurred())
			Expect(gcpGlobalAddress).NotTo(BeNil())
			chunks := strings.Split(ipRange.Status.Cidr, "/")
			expectedAddress := chunks[0]
			expectedPrefixLen, err := strconv.ParseInt(chunks[1], 10, 32)
			Expect(err).NotTo(HaveOccurred())
			Expect(gcpGlobalAddress.Address).To(Equal(expectedAddress))
			Expect(gcpGlobalAddress.PrefixLength).To(Equal(expectedPrefixLen))
		})

		By("And Then GCP PSA is created ("+implName+")", func() {
			list, err := infra.GcpMock().ListServiceConnections(infra.Ctx(), scope.Spec.Scope.Gcp.Project, scope.Spec.Scope.Gcp.VpcNetwork)
			Expect(err).NotTo(HaveOccurred())
			Expect(list).To(HaveLen(1))
			Expect(list[0].Service).To(Equal(gcpclient.ServiceNetworkingServicePath))
			Expect(list[0].ReservedPeeringRanges).To(HaveLen(1))
			Expect(list[0].ReservedPeeringRanges[0]).To(Equal("cm-" + ipRange.Name))
		})

		// Delete

		By("When KCP IpRange is deleted ("+implName+")", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), ipRange).
				Should(Succeed(), "failed deleting KCP IpRange")
		})

		By("Then KCP IpRange does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), ipRange).
				Should(Succeed(), "expected KCP IpRange to be deleted, but it exists")
		})

		By("And Then GCP PSA does not exist ("+implName+")", func() {
			list, err := infra.GcpMock().ListServiceConnections(infra.Ctx(), scope.Spec.Scope.Gcp.Project, scope.Spec.Scope.Gcp.VpcNetwork)
			Expect(err).NotTo(HaveOccurred())
			Expect(list).To(HaveLen(0))
		})

		By("And Then GCP global address does not exist ("+implName+")", func() {
			gcpGlobalAddress, err := infra.GcpMock().GetIpRange(infra.Ctx(), scope.Spec.Scope.Gcp.Project, "cm-"+ipRange.Name)
			Expect(gcpGlobalAddress).To(BeNil())
			Expect(gcpmeta.IsNotFound(err)).To(BeTrue())
		})
	}

	Context("with legacy implementation (v2)", func() {
		It("Scenario: IpRange lifecycle with specified CIDR - legacy v2", func() {
			testIpRangeLifecycleWithCidr(false, "legacy")
		})
	})

	Context("with refactored implementation", func() {
		PIt("Scenario: IpRange lifecycle with specified CIDR - refactored", func() {
			// TODO: Enable once refactored implementation is stable
			testIpRangeLifecycleWithCidr(true, "refactored")
		})
	})

	Context("comparing both implementations", func() {
		PIt("should produce identical results for specified CIDR", func() {
			// TODO: Create two IpRanges with different implementations
			// and verify they produce identical GCP resources and status
			Skip("Comparison test not yet implemented")
		})

		PIt("should produce identical results for allocated CIDR", func() {
			// TODO: Test auto-allocation with both implementations
			Skip("Comparison test not yet implemented")
		})
	})
})
