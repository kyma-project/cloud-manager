package cloudcontrol

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	kcpscope "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Feature: KCP NfsInstance for Alicloud", func() {

	It("Scenario: KCP Alicloud NfsInstance (NAS) is created and deleted", func() {
		const (
			kymaName    = "ac-nfs-01"
			iprangeName = "ac-nfs-iprange-01"
			nfsName     = "ac-nfs-instance-01"
			iprangeCidr = "10.181.0.0/22"
			region      = "ap-southeast-1"
		)

		alicloudAccount := infra.AlicloudMock().NewAccount()
		defer alicloudAccount.Delete()
		alicloudRegion := alicloudAccount.Region(region)

		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			kcpscope.Ignore.AddName(kymaName)
			Eventually(CreateScopeAlicloud).
				WithArguments(infra.Ctx(), infra, scope, alicloudAccount.Credentials().AccessKeyId, WithName(kymaName)).
				Should(Succeed())
		})

		vpcName := common.GardenerVpcName(scope.Namespace, kymaName)
		alicloudVpcName := vpcName + "-vpc"

		By("And Given Alicloud VPC exists in mock (Gardener naming convention)", func() {
			alicloudRegion.AddVpc("vpc-ac-nfs", alicloudVpcName, "10.180.0.0/16")
			alicloudRegion.AddZone("ap-southeast-1a")
		})

		var kcpNetworkKyma *cloudcontrolv1beta1.Network

		By("And Given KCP Kyma Network exists in Ready state", func() {
			kcpNetworkKyma = cloudcontrolv1beta1.NewNetworkBuilder().
				WithScope(kymaName).
				WithName(common.KcpNetworkKymaCommonName(kymaName)).
				WithAlicloudRef(scope.Spec.Scope.Alicloud.AccountId, scope.Spec.Region, "vpc-ac-nfs", scope.Spec.Scope.Alicloud.VpcNetwork).
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

		iprange := &cloudcontrolv1beta1.IpRange{}

		By("And Given KCP IpRange exists in Ready state", func() {
			Eventually(CreateKcpIpRange).
				WithArguments(infra.Ctx(), infra.KCP().Client(), iprange,
					WithName(iprangeName),
					WithKcpIpRangeRemoteRef(iprangeName),
					WithScope(kymaName),
					WithKcpIpRangeSpecCidr(iprangeCidr),
				).
				Should(Succeed())

			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), iprange,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
				).
				Should(Succeed())
		})

		nfsInstance := &cloudcontrolv1beta1.NfsInstance{}

		By("When KCP NfsInstance is created", func() {
			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.KCP().Client(), nfsInstance,
					WithName(nfsName),
					WithRemoteRef(nfsName),
					WithScope(kymaName),
					WithIpRange(iprangeName),
					WithNfsInstanceAlicloud(),
				).
				Should(Succeed(), "failed creating NfsInstance")
		})

		By("Then KCP NfsInstance has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), nfsInstance,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
				).
				Should(Succeed(), "expected NfsInstance to have Ready condition")
		})

		By("And Then KCP NfsInstance has status.id (NAS file system id)", func() {
			Expect(nfsInstance.Status.Id).NotTo(BeEmpty())
		})

		By("And Then KCP NfsInstance has status.host and status.path", func() {
			Expect(nfsInstance.Status.Host).NotTo(BeEmpty())
			Expect(nfsInstance.Status.Path).To(Equal("/"))
		})

		By("And Then NAS file system exists in mock", func() {
			fs, err := alicloudRegion.NfsInstanceClient().DescribeFileSystem(infra.Ctx(), nfsInstance.Status.Id)
			Expect(err).NotTo(HaveOccurred())
			Expect(fs).NotTo(BeNil())
			Expect(fs.Status).To(Equal("Running"))
		})

		By("And Then NAS mount target exists in mock", func() {
			mts, err := alicloudRegion.NfsInstanceClient().DescribeMountTargets(infra.Ctx(), nfsInstance.Status.Id)
			Expect(err).NotTo(HaveOccurred())
			Expect(mts).To(HaveLen(1))
			Expect(mts[0].VpcId).To(Equal("vpc-ac-nfs"))
			Expect(mts[0].VSwitchId).To(Equal(iprange.Status.Subnets[0].Id))
		})

		// DELETE ======================================================

		By("When KCP NfsInstance is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), nfsInstance).
				Should(Succeed())
		})

		By("Then KCP NfsInstance does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), nfsInstance).
				Should(Succeed())
		})

		By("And Then NAS mount targets are deleted from mock", func() {
			mts, err := alicloudRegion.NfsInstanceClient().DescribeMountTargets(infra.Ctx(), nfsInstance.Status.Id)
			Expect(err).NotTo(HaveOccurred())
			Expect(mts).To(BeEmpty())
		})

		By("And Then NAS file system is deleted from mock", func() {
			fs, err := alicloudRegion.NfsInstanceClient().DescribeFileSystem(infra.Ctx(), nfsInstance.Status.Id)
			Expect(err).NotTo(HaveOccurred())
			Expect(fs).To(BeNil())
		})

		By("// cleanup: delete KCP IpRange", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), iprange).
				Should(Succeed())
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), iprange).
				Should(Succeed())
		})

		By("// cleanup: delete KCP Kyma Network", func() {
			Expect(infra.KCP().Client().Delete(infra.Ctx(), kcpNetworkKyma)).To(Succeed())
		})

		By("// cleanup: delete Scope", func() {
			Expect(infra.KCP().Client().Delete(infra.Ctx(), scope)).To(Succeed())
		})
	})
})
