package cloudcontrol

// Run with: FF_IP_RANGE_V2=true go test

import (
	"context"

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
		if !feature.IpRangeV2.Value(context.Background()) {
			Skip("IpRange refactored implementation is disabled")
		}

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
			Eventually(CreateKcpIpRange).
				WithArguments(infra.Ctx(), infra.KCP().Client(), ipRange,
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

		By("And Then GCP PSA connection is created", func() {
			Eventually(func() error {
				connections, err := infra.GcpMock().ListServiceConnections(infra.Ctx(), scope.Spec.Scope.Gcp.Project, scope.Spec.Scope.Gcp.VpcNetwork)
				if err != nil {
					return err
				}
				if len(connections) == 0 {
					return nil
				}
				return nil
			}).Should(Succeed())
		})

		By("And Then GCP PSA connection includes the IP range", func() {
			connections, err := infra.GcpMock().ListServiceConnections(infra.Ctx(), scope.Spec.Scope.Gcp.Project, scope.Spec.Scope.Gcp.VpcNetwork)
			Expect(err).NotTo(HaveOccurred())
			Expect(connections).NotTo(BeEmpty())
			Expect(connections[0].ReservedPeeringRanges).To(ContainElement("cm-" + ipRange.Name))
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

		By("And Then GCP PSA connection no longer includes the IP range", func() {
			connections, err := infra.GcpMock().ListServiceConnections(infra.Ctx(), scope.Spec.Scope.Gcp.Project, scope.Spec.Scope.Gcp.VpcNetwork)
			Expect(err).NotTo(HaveOccurred())
			if len(connections) > 0 {
				Expect(connections[0].ReservedPeeringRanges).NotTo(ContainElement("cm-" + ipRange.Name))
			}
		})
	})

	It("Scenario: KCP GCP IpRange lifecycle without CIDR (refactored)", func() {
		if !feature.IpRangeV2.Value(context.Background()) {
			Skip("IpRange refactored implementation is disabled")
		}

		const (
			kymaName    = "4d8c28c0-d1dd-4711-833d-aa2c559c9cc8"
			ipRangeName = "d0590de2-aae5-42c2-b11d-f9506f14e5c6"
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

		By("When KCP IpRange is created WITHOUT CIDR with refactored implementation", func() {
			Eventually(CreateKcpIpRange).
				WithArguments(infra.Ctx(), infra.KCP().Client(), ipRange,
					WithName(ipRangeName),
					// Note: NO WithKcpIpRangeSpecCidr() - testing defaulting
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

		By("And Then KCP IpRange has status.cidr with a default value", func() {
			Expect(ipRange.Status.Cidr).NotTo(BeEmpty())
			Expect(ipRange.Status.Cidr).To(MatchRegexp(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}/\d{1,2}$`))
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

		By("And Then GCP PSA connection is created", func() {
			Eventually(func() error {
				connections, err := infra.GcpMock().ListServiceConnections(infra.Ctx(), scope.Spec.Scope.Gcp.Project, scope.Spec.Scope.Gcp.VpcNetwork)
				if err != nil {
					return err
				}
				if len(connections) == 0 {
					return nil
				}
				return nil
			}).Should(Succeed())
		})

		By("And Then GCP PSA connection includes the IP range", func() {
			connections, err := infra.GcpMock().ListServiceConnections(infra.Ctx(), scope.Spec.Scope.Gcp.Project, scope.Spec.Scope.Gcp.VpcNetwork)
			Expect(err).NotTo(HaveOccurred())
			Expect(connections).NotTo(BeEmpty())
			Expect(connections[0].ReservedPeeringRanges).To(ContainElement("cm-" + ipRange.Name))
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

		By("And Then GCP PSA connection no longer includes the IP range", func() {
			connections, err := infra.GcpMock().ListServiceConnections(infra.Ctx(), scope.Spec.Scope.Gcp.Project, scope.Spec.Scope.Gcp.VpcNetwork)
			Expect(err).NotTo(HaveOccurred())
			if len(connections) > 0 {
				Expect(connections[0].ReservedPeeringRanges).NotTo(ContainElement("cm-" + ipRange.Name))
			}
		})
	})
})
