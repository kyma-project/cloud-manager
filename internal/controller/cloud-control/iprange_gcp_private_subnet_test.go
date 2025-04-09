package cloudcontrol

import (
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	v3 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange/v3/client"
	gcpmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/meta"
	kcpscope "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Feature: KCP IpRange for GCP Private Subnet", func() {

	It("Scenario: KCP IpRange with specified CIDR is created and deleted", func() {
		const (
			kymaName    = "bbf533b8-31e5-44d1-ae8a-f6cffd67bdfb"
			ipRangeName = "9778c07f-ae9c-4eb6-ad05-5be95f1652b2"
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
					WithKcpIpRangeSpecCidr("10.20.60.0/24"),
					WithKcpIpRangeGcpType(cloudcontrolv1beta1.GcpIpRangeTypeSUBNET),
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

		By("And Then GCP Private Subnet is created", func() {
			privateSubnet, err := infra.GcpMock().GetSubnet(infra.Ctx(), v3.GetSubnetRequest{
				ProjectId: scope.Spec.Scope.Gcp.Project,
				Name:      "cm-" + ipRange.Name,
				Region:    scope.Spec.Region,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(privateSubnet).NotTo(BeNil())
		})

		By("And Then GCP Connection Policy is created", func() {
			policyName := fmt.Sprintf("projects/%s/locations/%s/serviceConnectionPolicies/cm-%s-%s-rc",
				scope.Spec.Scope.Gcp.Project, scope.Spec.Region, scope.Spec.Scope.Gcp.VpcNetwork, scope.Spec.Region,
			)
			connectionPolicy, err := infra.GcpMock().GetServiceConnectionPolicy(infra.Ctx(), policyName)
			Expect(err).NotTo(HaveOccurred())
			Expect(connectionPolicy).NotTo(BeNil())
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

		By("And Then GCP Connection Policy does not exist", func() {
			privateSubnet, err := infra.GcpMock().GetSubnet(infra.Ctx(), v3.GetSubnetRequest{
				ProjectId: scope.Spec.Scope.Gcp.Project,
				Name:      "cm-" + ipRange.Name,
				Region:    scope.Spec.Region,
			})
			Expect(privateSubnet).To(BeNil())
			Expect(gcpmeta.IsNotFound(err)).To(BeTrue())
		})

		By("And Then GCP Private Subnet does not exist", func() {
			policyName := fmt.Sprintf("projects/%s/locations/%s/serviceConnectionPolicies/cm-%s-%s-redis-cluster",
				scope.Spec.Scope.Gcp.Project, scope.Spec.Region, scope.Spec.Scope.Gcp.VpcNetwork, scope.Spec.Region,
			)
			connectionPolicy, err := infra.GcpMock().GetServiceConnectionPolicy(infra.Ctx(), policyName)
			Expect(connectionPolicy).To(BeNil())
			Expect(gcpmeta.IsNotFound(err)).To(BeTrue())
		})

	})

	It("Scenario: KCP IpRange without CIDR is created and deleted", func() {
		const (
			kymaName    = "9f99fbf7-09da-4cc6-b286-010df44f473b"
			ipRangeName = "8f533b18-ccb1-417c-8a68-512a7d4859a5"
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
					WithKcpIpRangeGcpType(cloudcontrolv1beta1.GcpIpRangeTypeSUBNET),
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

		By("And Then GCP Private Subnet is created", func() {
			privateSubnet, err := infra.GcpMock().GetSubnet(infra.Ctx(), v3.GetSubnetRequest{
				ProjectId: scope.Spec.Scope.Gcp.Project,
				Name:      "cm-" + ipRange.Name,
				Region:    scope.Spec.Region,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(privateSubnet).NotTo(BeNil())
		})

		By("And Then GCP Connection Policy is created", func() {
			policyName := fmt.Sprintf("projects/%s/locations/%s/serviceConnectionPolicies/cm-%s-%s-rc",
				scope.Spec.Scope.Gcp.Project, scope.Spec.Region, scope.Spec.Scope.Gcp.VpcNetwork, scope.Spec.Region,
			)
			connectionPolicy, err := infra.GcpMock().GetServiceConnectionPolicy(infra.Ctx(), policyName)
			Expect(err).NotTo(HaveOccurred())
			Expect(connectionPolicy).NotTo(BeNil())
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

		By("And Then GCP Connection Policy does not exist", func() {
			gcpGlobalAddress, err := infra.GcpMock().GetIpRange(infra.Ctx(), scope.Spec.Scope.Gcp.Project, "cm-"+ipRange.Name)
			Expect(gcpGlobalAddress).To(BeNil())
			Expect(gcpmeta.IsNotFound(err)).To(BeTrue())
		})

		By("And Then GCP Connection Policy does not exist", func() {
			privateSubnet, err := infra.GcpMock().GetSubnet(infra.Ctx(), v3.GetSubnetRequest{
				ProjectId: scope.Spec.Scope.Gcp.Project,
				Name:      "cm-" + ipRange.Name,
				Region:    scope.Spec.Region,
			})
			Expect(privateSubnet).To(BeNil())
			Expect(gcpmeta.IsNotFound(err)).To(BeTrue())
		})

		By("And Then GCP Private Subnet does not exist", func() {
			policyName := fmt.Sprintf("projects/%s/locations/%s/serviceConnectionPolicies/cm-%s-%s-redis-cluster",
				scope.Spec.Scope.Gcp.Project, scope.Spec.Region, scope.Spec.Scope.Gcp.VpcNetwork, scope.Spec.Region,
			)
			connectionPolicy, err := infra.GcpMock().GetServiceConnectionPolicy(infra.Ctx(), policyName)
			Expect(connectionPolicy).To(BeNil())
			Expect(gcpmeta.IsNotFound(err)).To(BeTrue())
		})

	})

})
