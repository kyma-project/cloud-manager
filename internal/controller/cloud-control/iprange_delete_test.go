package cloudcontrol

import (
	"fmt"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/elliotchance/pie/v2"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	kcpnfsinstance "github.com/kyma-project/cloud-manager/pkg/kcp/nfsinstance"
	awsmock "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/mock"
	awsutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/util"
	kcpredisinstance "github.com/kyma-project/cloud-manager/pkg/kcp/redisinstance"
	scopePkg "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// It is not important what's the underlying cloud provider since the SUT is in the common KCP IpRange flow.
// The AWS provider is just randomly picked as the most convenient one to write the test.

var _ = Describe("Feature: KCP IpRange deletion with dependant objects", func() {

	It("Scenario: KCP AWS IpRange is deleted with existing NfsInstance", func() {
		const (
			kymaName        = "8ca09bf7-0ca2-447e-9466-9b1c0d44ecba"
			vpcId           = "27d608ae-a953-4dc5-8642-44da6d626eee"
			vpcCidr         = "10.180.0.0/16"
			iprangeName     = "6a37cbb9-ae17-4ff8-9ec5-2d198ff9dc42"
			iprangeCidr     = "10.181.0.0/16"
			nfsInstanceName = "40133258-d90b-497a-9b43-1281df9ed82d"
		)

		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			// Tell Scope reconciler to ignore this kymaName
			scopePkg.Ignore.AddName(kymaName)

			Eventually(CreateScopeAws).
				WithArguments(infra.Ctx(), infra, scope, WithName(kymaName)).
				Should(Succeed())
		})

		awsMock := infra.AwsMock().MockConfigs(scope.Spec.Scope.Aws.AccountId, scope.Spec.Region)

		By("And Given AWS VPC exists", func() {
			var _ *ec2Types.Vpc
			_ = awsMock.AddVpc(
				vpcId,
				vpcCidr,
				awsutil.Ec2Tags("Name", scope.Spec.Scope.Aws.VpcNetwork),
				awsmock.VpcSubnetsFromScope(scope),
			)
		})

		var kcpNetworkKyma *cloudcontrolv1beta1.Network

		By("And Given KCP Kyma Network exists in Ready state", func() {
			kcpNetworkKyma = cloudcontrolv1beta1.NewNetworkBuilder().
				WithScope(kymaName).
				WithName(common.KcpNetworkKymaCommonName(kymaName)).
				WithAwsRef(scope.Spec.Scope.Aws.AccountId, scope.Spec.Region, vpcId, scope.Spec.Scope.Aws.VpcNetwork).
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

		By("And Given KCP IpRange is created", func() {
			Eventually(CreateKcpIpRange).
				WithArguments(infra.Ctx(), infra.KCP().Client(), iprange,
					WithName(iprangeName),
					WithKcpIpRangeRemoteRef("skr-aws-ip-range"),
					WithScope(kymaName),
					WithKcpIpRangeSpecCidr(iprangeCidr),
				).
				Should(Succeed())
		})

		By("And Given KCP IpRange has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), iprange,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
				).
				Should(Succeed())
		})

		nfsInstance := &cloudcontrolv1beta1.NfsInstance{}

		By("And Given NfsInstance using KCP IpRange exists", func() {
			kcpnfsinstance.Ignore.AddName(nfsInstanceName)
			Expect(CreateNfsInstance(infra.Ctx(), infra.KCP().Client(), nfsInstance,
				WithName(nfsInstanceName),
				WithRemoteRef("foo"),
				WithScope(kymaName),
				WithIpRange(iprangeName),
				WithNfsInstanceAws(),
			)).
				To(Succeed())
		})

		By("When IpRange is marked for deletion", func() {
			Expect(Delete(infra.Ctx(), infra.KCP().Client(), iprange)).
				To(Succeed())
		})

		By("Then IpRange has warning state", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), iprange,
					NewObjActions(),
					HavingState(string(cloudcontrolv1beta1.WarningState)),
				).
				Should(Succeed())
		})

		By("And Then IpRange has DeleteWhileUsed Warning condition", func() {
			cond := meta.FindStatusCondition(iprange.Status.Conditions, cloudcontrolv1beta1.ConditionTypeWarning)
			Expect(cond).ToNot(BeNil(), fmt.Sprintf(
				"Expected Warning condition, but found: %v",
				pie.Map(iprange.Status.Conditions, func(c metav1.Condition) string {
					return c.Type
				}),
			))
			Expect(cond.Reason).To(Equal(cloudcontrolv1beta1.ReasonDeleteWhileUsed),
				fmt.Sprintf("Expected Reason to equal %s, but found %s", cloudcontrolv1beta1.ReasonDeleteWhileUsed, cond.Reason))
			Expect(cond.Status).To(Equal(metav1.ConditionTrue), fmt.Sprintf("Expected True status, but found: %s", cond.Status))
		})

		By("When NfsInstance is deleted", func() {
			Expect(Delete(infra.Ctx(), infra.KCP().Client(), nfsInstance)).
				To(Succeed())
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), nfsInstance).
				Should(Succeed())
		})

		By("Then IpRange is deleted", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), iprange).
				Should(Succeed())
		})

		By("// cleanup: delete KCP Kyma Network", func() {
			Expect(Delete(infra.Ctx(), infra.KCP().Client(), kcpNetworkKyma)).
				To(Succeed())
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpNetworkKyma).
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

	It("Scenario: KCP AWS IpRange is deleted with existing RedisInstance", func() {
		const (
			kymaName          = "9d88aa77-0f11-4885-b3f9-d284d9312dae"
			vpcId             = "aab35d3c-0a0d-4b0c-91b9-2948888b0d65"
			vpcCidr           = "10.180.0.0/16"
			iprangeName       = "29dfe162-d63f-4768-8bf2-c3d336731297"
			iprangeCidr       = "10.181.0.0/16"
			redisInstanceName = "82a3bd14-6806-43ae-b702-0a77f811a918"
		)

		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			// Tell Scope reconciler to ignore this kymaName
			scopePkg.Ignore.AddName(kymaName)

			Eventually(CreateScopeAws).
				WithArguments(infra.Ctx(), infra, scope, WithName(kymaName)).
				Should(Succeed())
		})

		awsMock := infra.AwsMock().MockConfigs(scope.Spec.Scope.Aws.AccountId, scope.Spec.Region)

		By("And Given AWS VPC exists", func() {
			var _ *ec2Types.Vpc
			_ = awsMock.AddVpc(
				vpcId,
				vpcCidr,
				awsutil.Ec2Tags("Name", scope.Spec.Scope.Aws.VpcNetwork),
				awsmock.VpcSubnetsFromScope(scope),
			)
		})

		var kcpNetworkKyma *cloudcontrolv1beta1.Network

		By("And Given KCP Kyma Network exists in Ready state", func() {
			kcpNetworkKyma = cloudcontrolv1beta1.NewNetworkBuilder().
				WithScope(kymaName).
				WithName(common.KcpNetworkKymaCommonName(kymaName)).
				WithAwsRef(scope.Spec.Scope.Aws.AccountId, scope.Spec.Region, vpcId, scope.Spec.Scope.Aws.VpcNetwork).
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

		By("And Given KCP IpRange is created", func() {
			Eventually(CreateKcpIpRange).
				WithArguments(infra.Ctx(), infra.KCP().Client(), iprange,
					WithName(iprangeName),
					WithKcpIpRangeRemoteRef("skr-aws-ip-range"),
					WithScope(kymaName),
					WithKcpIpRangeSpecCidr(iprangeCidr),
				).
				Should(Succeed())
		})

		By("And Given KCP IpRange has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), iprange,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
				).
				Should(Succeed())
		})

		redisInstance := &cloudcontrolv1beta1.RedisInstance{}

		By("And Given RedisInstance using KCP IpRange exists", func() {
			kcpredisinstance.Ignore.AddName(redisInstanceName)
			Expect(CreateRedisInstance(infra.Ctx(), infra.KCP().Client(), redisInstance,
				WithName(redisInstanceName),
				WithRemoteRef("foo"),
				WithScope(kymaName),
				WithIpRange(iprangeName),
				WithRedisInstanceAws(),
			)).
				To(Succeed())
		})

		By("When IpRange is marked for deletion", func() {
			Expect(Delete(infra.Ctx(), infra.KCP().Client(), iprange)).
				To(Succeed())
		})

		By("Then IpRange has warning state", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), iprange,
					NewObjActions(),
					HavingState(string(cloudcontrolv1beta1.WarningState)),
				).
				Should(Succeed())
		})

		By("And Then IpRange has DeleteWhileUsed Warning condition", func() {
			cond := meta.FindStatusCondition(iprange.Status.Conditions, cloudcontrolv1beta1.ConditionTypeWarning)
			Expect(cond).ToNot(BeNil(), fmt.Sprintf(
				"Expected Warning condition, but found: %v",
				pie.Map(iprange.Status.Conditions, func(c metav1.Condition) string {
					return c.Type
				}),
			))
			Expect(cond.Reason).To(Equal(cloudcontrolv1beta1.ReasonDeleteWhileUsed),
				fmt.Sprintf("Expected Reason to equal %s, but found %s", cloudcontrolv1beta1.ReasonDeleteWhileUsed, cond.Reason))
			Expect(cond.Status).To(Equal(metav1.ConditionTrue), fmt.Sprintf("Expected True status, but found: %s", cond.Status))
		})

		By("When RedisInstance is deleted", func() {
			Expect(Delete(infra.Ctx(), infra.KCP().Client(), redisInstance)).
				To(Succeed())
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance).
				Should(Succeed())
		})

		By("Then IpRange is deleted", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), iprange).
				Should(Succeed())
		})

		By("// cleanup: delete KCP Kyma Network", func() {
			Expect(Delete(infra.Ctx(), infra.KCP().Client(), kcpNetworkKyma)).
				To(Succeed())
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpNetworkKyma).
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
