package cloudcontrol

import (
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/privatedns/armprivatedns"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	kcpnetwork "github.com/kyma-project/cloud-manager/pkg/kcp/network"
	azurecommon "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/common"
	azureutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	scopePkg "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	kcpvpcpeering "github.com/kyma-project/cloud-manager/pkg/kcp/vpcpeering"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

var _ = Describe("Feature: KCP IpRange for Azure", func() {

	It("Scenario: KCP Azure IpRange is created and deleted", func() {

		const (
			kymaName = "6aa2a9a4-4c7b-481d-b078-a6eddf7b440e"
		)

		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			// Tell Scope reconciler to ignore this kymaName
			scopePkg.Ignore.AddName(kymaName)

			Eventually(CreateScopeAzure).
				WithArguments(infra.Ctx(), infra, scope, WithName(kymaName)).
				Should(Succeed())
		})

		var kcpNetworkKyma *cloudcontrolv1beta1.Network

		By("And Given KCP Kyma Network exists in Ready state", func() {
			kcpNetworkKyma = cloudcontrolv1beta1.NewNetworkBuilder().
				WithScope(kymaName).
				WithName(common.KcpNetworkKymaCommonName(kymaName)).
				WithAzureRef(scope.Spec.Scope.Azure.TenantId, scope.Spec.Scope.Azure.SubscriptionId, scope.Spec.Scope.Azure.VpcNetwork, scope.Spec.Scope.Azure.VpcNetwork).
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

		azureMock := infra.AzureMock().MockConfigs(kcpNetworkKyma.Spec.Network.Reference.Azure.SubscriptionId, kcpNetworkKyma.Spec.Network.Reference.Azure.TenantId)

		By("And Given Azure Kyma Shoot VNet exist", func() {
			err := azureMock.CreateNetwork(
				infra.Ctx(),
				kcpNetworkKyma.Spec.Network.Reference.Azure.ResourceGroup,
				kcpNetworkKyma.Spec.Network.Reference.Azure.NetworkName,
				scope.Spec.Region,
				"10.250.0.0/22",
				nil,
			)
			Expect(err).NotTo(HaveOccurred())
		})

		kcpIpRangeName := "d87e3449-777d-4d0e-bfec-95ef8dda436e"
		kcpIpRange := &cloudcontrolv1beta1.IpRange{}

		kcpNetworkCm := cloudcontrolv1beta1.NewNetworkBuilder().WithName(common.KcpNetworkCMCommonName(kymaName)).Build()
		kcpVpcPeering := cloudcontrolv1beta1.NewVpcPeeringBuilder().WithName(kymaName).Build()

		By("When KCP IpRange is created", func() {
			kcpnetwork.Ignore.AddName(kcpNetworkCm.Name)
			kcpvpcpeering.Ignore.AddName(kcpVpcPeering.Name)

			Eventually(CreateKcpIpRange).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpIpRange,
					WithName(kcpIpRangeName),
					WithKcpIpRangeRemoteRef("some-remote-ref"),
					WithKcpIpRangeNetwork(kcpNetworkCm.Name),
					WithScope(kymaName),
				).
				Should(Succeed())
		})

		By("Then KCP IpRange has no Error condition", func() {
			Consistently(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpIpRange, NewObjActions(),
					NotHavingConditionTrue(cloudcontrolv1beta1.ConditionTypeError)).
				Should(Succeed())
		})

		By("Then KCP IpRange has allocated CIDR in status", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpIpRange, NewObjActions(),
					HavingKcpIpRangeStatusCidr(common.DefaultCloudManagerCidr)).
				Should(Succeed())
		})

		By("And Then KCP CM Network is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpNetworkCm, NewObjActions()).
				Should(Succeed())
			added, err := composed.PatchObjAddFinalizer(infra.Ctx(), cloudcontrolv1beta1.FinalizerName, kcpNetworkCm, infra.KCP().Client())
			Expect(added).To(BeTrue())
			Expect(err).NotTo(HaveOccurred())
		})

		By("And Then KCP CM Network has scope as KCP IpRange", func() {
			Expect(kcpNetworkCm.Spec.Scope.Name).To(Equal(kcpIpRange.Spec.Scope.Name))
		})

		By("And Then KCP CM Network is managed", func() {
			Expect(kcpNetworkCm.Spec.Network.Managed).NotTo(BeNil())
		})

		By("And Then KCP CM Network has cidr same as KCP IpRange status cidr", func() {
			Expect(kcpNetworkCm.Spec.Network.Managed.Cidr).To(Equal(kcpIpRange.Status.Cidr))
		})

		By("And Then KCP CM Network has location same as scope region", func() {
			Expect(kcpNetworkCm.Spec.Network.Managed.Location).To(Equal(scope.Spec.Region))
		})

		cmCommonName := azurecommon.AzureCloudManagerResourceGroupName(scope.Spec.Scope.Azure.VpcNetwork)

		By("When Azure CM Network is provisioned", func() {
			err := azureMock.CreateNetwork(infra.Ctx(), cmCommonName, cmCommonName, scope.Spec.Region, kcpIpRange.Status.Cidr, nil)
			Expect(err).NotTo(HaveOccurred())
		})

		By("When KCP CM Network has Ready condition", func() {
			kcpNetworkCm.Status.State = string(cloudcontrolv1beta1.ReadyState)
			meta.SetStatusCondition(&kcpNetworkCm.Status.Conditions, metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeReady,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ConditionTypeReady,
				Message: "Ready",
			})
			err := composed.PatchObjStatus(infra.Ctx(), kcpNetworkCm, infra.KCP().Client())
			Expect(err).NotTo(HaveOccurred())
		})

		By("Then KCP VpcPeering is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpVpcPeering, NewObjActions()).
				Should(Succeed())
			added, err := composed.PatchObjAddFinalizer(infra.Ctx(), cloudcontrolv1beta1.FinalizerName, kcpVpcPeering, infra.KCP().Client())
			Expect(added).To(BeTrue())
			Expect(err).NotTo(HaveOccurred())
		})

		By("And Then KCP VpcPeering has scope as KCP IpRange", func() {
			Expect(kcpVpcPeering.Spec.Scope.Name).To(Equal(kcpIpRange.Spec.Scope.Name))
		})

		By("And Then KCP VpcPeering local network ref to KCP Kyma Network", func() {
			Expect(kcpVpcPeering.Spec.Details.LocalNetwork.Name).To(Equal(kcpNetworkKyma.Name))
		})

		By("And Then KCP VpcPeering remote network ref to KCP CM Network", func() {
			Expect(kcpVpcPeering.Spec.Details.RemoteNetwork.Name).To(Equal(kcpNetworkCm.Name))
		})

		By("And Then KCP VpcPeering remote peering name is kyma--SHOOT", func() {
			Expect(kcpVpcPeering.Spec.Details.PeeringName).To(Equal("kyma--" + scope.Spec.ShootName))
		})

		By("And Then KCP VpcPeering local peering name is cm--SHOOT", func() {
			Expect(kcpVpcPeering.Spec.Details.LocalPeeringName).To(Equal("cm--" + scope.Spec.ShootName))
		})

		By("When KCP VpcPeering has Ready condition", func() {
			kcpVpcPeering.Status.State = string(cloudcontrolv1beta1.ReadyState)
			meta.SetStatusCondition(&kcpVpcPeering.Status.Conditions, metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeReady,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ConditionTypeReady,
				Message: "Ready",
			})
			err := composed.PatchObjStatus(infra.Ctx(), kcpVpcPeering, infra.KCP().Client())
			Expect(err).NotTo(HaveOccurred())
		})

		var azureSecurityGroup *armnetwork.SecurityGroup

		By("Then Azure Security group is created", func() {
			Eventually(func() error {
				sg, err := azureMock.GetSecurityGroup(infra.Ctx(), cmCommonName, cmCommonName)
				if err != nil {
					return fmt.Errorf("error loading azure security group: %w", err)
				}
				azureSecurityGroup = sg
				return nil
			}).
				Should(Succeed())
			Expect(azureSecurityGroup).NotTo(BeNil())
		})

		var virtualNetworkLink *armprivatedns.VirtualNetworkLink
		resourceGroupName := azurecommon.AzureCloudManagerResourceGroupName(scope.Spec.Scope.Azure.VpcNetwork)
		privateDnsZoneName := azureutil.NewPrivateDnsZoneName()
		By("Then Azure Virtual Network Link is created", func() {
			Eventually(func() error {
				vnl, err := azureMock.GetVirtualNetworkLink(infra.Ctx(), resourceGroupName, privateDnsZoneName, kcpIpRangeName)
				if err != nil {
					return fmt.Errorf("error loading azure Virtual Network Link: %w", err)
				}
				virtualNetworkLink = vnl
				return nil
			}).
				Should(Succeed())
			Expect(virtualNetworkLink).NotTo(BeNil())
		})

		var privateDnsZone *armprivatedns.PrivateZone
		By("Then Azure Private Dns Zone is created", func() {
			Eventually(func() error {
				pdz, err := azureMock.GetPrivateDnsZone(infra.Ctx(), resourceGroupName, privateDnsZoneName)
				if err != nil {
					return fmt.Errorf("error loading azure Private Dns Zone: %w", err)
				}
				privateDnsZone = pdz
				return nil
			}).
				Should(Succeed())
			Expect(privateDnsZone).NotTo(BeNil())
		})

		var azureSubnet *armnetwork.Subnet

		By("And Then Azure Subnet is created", func() {
			Eventually(func() error {
				sn, err := azureMock.GetSubnet(infra.Ctx(), cmCommonName, cmCommonName, cmCommonName)
				if err != nil {
					return fmt.Errorf("error loading azure subnet: %w", err)
				}
				azureSubnet = sn
				return nil
			}).Should(Succeed())
			Expect(azureSubnet).NotTo(BeNil())
		})

		By("And Then Azure Subnet has address range as KCP IpRange", func() {
			Expect(ptr.Deref(azureSubnet.Properties.AddressPrefix, "")).To(Equal(kcpIpRange.Status.Cidr))
		})

		By("And Then Azure Subnet has security group", func() {
			Expect(azureSubnet.Properties.NetworkSecurityGroup).NotTo(BeNil())
			Expect(ptr.Deref(azureSubnet.Properties.NetworkSecurityGroup.ID, "")).To(Equal(ptr.Deref(azureSecurityGroup.ID, "")))
		})

		By("And Then KCP IpRange has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpIpRange, NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady)).
				Should(Succeed())
		})

		// Delete

		By("When KCP IpRange is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpIpRange).
				Should(Succeed())
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpIpRange, NewObjActions(),
					HavingDeletionTimestamp()).
				Should(Succeed())
		})

		By("Then KCP VpcPeering is deleted", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpVpcPeering, NewObjActions(),
					HavingDeletionTimestamp()).
				Should(Succeed())
		})

		By("And Then KCP CM Network is deleted", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpNetworkCm, NewObjActions(),
					HavingDeletionTimestamp()).
				Should(Succeed())
		})

		By("When KCP VpcPeering finalizer is removed", func() {
			removed, err := composed.PatchObjRemoveFinalizer(infra.Ctx(), cloudcontrolv1beta1.FinalizerName, kcpVpcPeering, infra.KCP().Client())
			Expect(removed).To(BeTrue())
			Expect(err).NotTo(HaveOccurred())
		})

		By("Then KCP VpcPeering does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpVpcPeering).
				Should(Succeed())
		})

		By("Then Azure Virtual Network Link does not exists", func() {
			virtualNetworkLink, err := azureMock.GetVirtualNetworkLink(infra.Ctx(), resourceGroupName, privateDnsZoneName, kcpIpRangeName)
			Expect(virtualNetworkLink).To(BeNil())
			Expect(err).NotTo(HaveOccurred())
		})

		By("Then Azure Private Dns Zone does not exists", func() {
			privateDnsZone, err := azureMock.GetPrivateDnsZone(infra.Ctx(), resourceGroupName, privateDnsZoneName)
			Expect(privateDnsZone).To(BeNil())
			Expect(err).NotTo(HaveOccurred())
		})

		By("When KCP CM Network finalizer is removed", func() {
			removed, err := composed.PatchObjRemoveFinalizer(infra.Ctx(), cloudcontrolv1beta1.FinalizerName, kcpNetworkCm, infra.KCP().Client())
			Expect(removed).To(BeTrue())
			Expect(err).NotTo(HaveOccurred())
		})

		By("Then KCP CM Network does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpNetworkCm).
				Should(Succeed())
		})

		By("And Then KCP IpRange does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpIpRange).
				Should(Succeed())
		})
	})

})
