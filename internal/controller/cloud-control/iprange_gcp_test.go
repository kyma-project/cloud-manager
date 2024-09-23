package cloudcontrol

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/meta"
	kcpscope "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"strconv"
	"strings"
)

var _ = Describe("Feature: KCP IpRange for GCP", func() {

	It("Scenario: KCP IpRange with specified CIDR is created and deleted", func() {
		const (
			kymaName    = "570c0d27-d67a-44cc-908a-2c151d50303e"
			ipRangeName = "81dbff76-33ce-4a2b-be48-a48ea4f2a25b"
		)

		scope := &cloudcontrolv1beta1.Scope{}
		ipRange := &cloudcontrolv1beta1.IpRange{}

		By("Given Scope exists", func() {
			// Tell Scope reconciler to ignore this kymaName
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
			Eventually(CreateKcpIpRange).
				WithArguments(infra.Ctx(), infra.KCP().Client(), ipRange,
					WithName(ipRangeName),
					WithKcpIpRangeRemoteRef(ipRangeName),
					WithKcpIpRangeSpecCidr("10.20.30.0/24"),
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

		By("And Then KCP IpRange has status cidr equals to spec cidr", func() {
			Expect(ipRange.Status.Cidr).To(Equal(ipRange.Spec.Cidr))
		})

		By("And Then GCP global address is created", func() {
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

		By("And Then GCP PSA is created", func() {
			list, err := infra.GcpMock().ListServiceConnections(infra.Ctx(), scope.Spec.Scope.Gcp.Project, scope.Spec.Scope.Gcp.VpcNetwork)
			Expect(err).NotTo(HaveOccurred())
			Expect(list).To(HaveLen(1))
			Expect(list[0].Service).To(Equal(gcpclient.ServiceNetworkingServicePath))
			Expect(list[0].ReservedPeeringRanges).To(HaveLen(1))
			Expect(list[0].ReservedPeeringRanges[0]).To(Equal("cm-" + ipRange.Name))
		})

		// Delete

		By("When KCP IpRange is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), ipRange).
				Should(Succeed(), "failed deleting KCP IpRange")
		})

		By("Then KCP IpRange does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), ipRange).
				Should(Succeed(), "expected KCP IpRange to be deleted, but it exists")
		})

		By("And Then GCP PSA does not exist", func() {
			list, err := infra.GcpMock().ListServiceConnections(infra.Ctx(), scope.Spec.Scope.Gcp.Project, scope.Spec.Scope.Gcp.VpcNetwork)
			Expect(err).NotTo(HaveOccurred())
			Expect(list).To(HaveLen(0))
		})

		By("And Then GCP global address does not exist", func() {
			gcpGlobalAddress, err := infra.GcpMock().GetIpRange(infra.Ctx(), scope.Spec.Scope.Gcp.Project, "cm-"+ipRange.Name)
			Expect(gcpGlobalAddress).To(BeNil())
			Expect(gcpmeta.IsNotFound(err)).To(BeTrue())
		})

	})

	It("Scenario: KCP IpRange without CIDR is created and deleted", func() {
		const (
			kymaName    = "89668d12-0132-4748-b175-fe66dd1c4d93"
			ipRangeName = "6a9123da-dacb-4803-abf9-6d238e903d40"
		)

		scope := &cloudcontrolv1beta1.Scope{}
		ipRange := &cloudcontrolv1beta1.IpRange{}

		By("Given Scope exists", func() {
			// Tell Scope reconciler to ignore this kymaName
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
			Eventually(CreateKcpIpRange).
				WithArguments(infra.Ctx(), infra.KCP().Client(), ipRange,
					WithName(ipRangeName),
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

		By("And Then KCP IpRange has status cidr", func() {
			Expect(ipRange.Status.Cidr).To(Equal(common.DefaultCloudManagerCidr))
		})

		By("And Then GCP global address is created", func() {
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

		By("And Then GCP PSA is created", func() {
			list, err := infra.GcpMock().ListServiceConnections(infra.Ctx(), scope.Spec.Scope.Gcp.Project, scope.Spec.Scope.Gcp.VpcNetwork)
			Expect(err).NotTo(HaveOccurred())
			Expect(list).To(HaveLen(1))
			Expect(list[0].Service).To(Equal(gcpclient.ServiceNetworkingServicePath))
			Expect(list[0].ReservedPeeringRanges).To(HaveLen(1))
			Expect(list[0].ReservedPeeringRanges[0]).To(Equal("cm-" + ipRange.Name))
		})

		// Delete

		By("When KCP IpRange is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), ipRange).
				Should(Succeed(), "failed deleting KCP IpRange")
		})

		By("Then KCP IpRange does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), ipRange).
				Should(Succeed(), "expected KCP IpRange to be deleted, but it exists")
		})

		By("And Then GCP PSA does not exist", func() {
			list, err := infra.GcpMock().ListServiceConnections(infra.Ctx(), scope.Spec.Scope.Gcp.Project, scope.Spec.Scope.Gcp.VpcNetwork)
			Expect(err).NotTo(HaveOccurred())
			Expect(list).To(HaveLen(0))
		})

		By("And Then GCP global address does not exist", func() {
			gcpGlobalAddress, err := infra.GcpMock().GetIpRange(infra.Ctx(), scope.Spec.Scope.Gcp.Project, "cm-"+ipRange.Name)
			Expect(gcpGlobalAddress).To(BeNil())
			Expect(gcpmeta.IsNotFound(err)).To(BeTrue())
		})

	})

})
