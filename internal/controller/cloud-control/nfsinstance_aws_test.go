package cloudcontrol

import (
	"fmt"
	"k8s.io/utils/ptr"
	"time"

	efstypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	"github.com/elliotchance/pie/v2"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	kcpiprange "github.com/kyma-project/cloud-manager/pkg/kcp/iprange"
	awsmock "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/mock"
	awsutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/util"
	kcpscope "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Feature: KCP NfsInstance AWS", func() {

	It("Scenario: KCP AWS NfsInstance is created and deleted", func() {

		name := "5338ac8f-4927-40ee-a51d-e22e2334bd19"
		scope := &cloudcontrolv1beta1.Scope{}

		By("Given AWS Scope exists", func() {
			// Tell Scope reconciler to ignore this Scope
			kcpscope.Ignore.AddName(name)

			Eventually(CreateScopeAws).
				WithArguments(infra.Ctx(), infra, scope, WithName(name)).
				Should(Succeed(), "failed creating Scope")
		})

		vpcId := "85b43d7c-6488-4e15-9782-fff7bc31c286"

		awsMock := infra.AwsMock().MockConfigs(scope.Spec.Scope.Aws.AccountId, scope.Spec.Region)

		By("And Given AWS VPC exists", func() {
			awsMock.AddVpc(
				"wrong1",
				"10.200.0.0/16",
				awsutil.Ec2Tags("Name", "wrong1"),
				nil,
			)
			awsMock.AddVpc(
				vpcId,
				"10.180.0.0/16",
				awsutil.Ec2Tags("Name", scope.Spec.Scope.Aws.VpcNetwork),
				awsmock.VpcSubnetsFromScope(scope),
			)
			awsMock.AddVpc(
				"wrong2",
				"10.200.0.0/16",
				awsutil.Ec2Tags("Name", "wrong2"),
				nil,
			)
		})

		iprange := &cloudcontrolv1beta1.IpRange{}
		iprangeCidr := "10.181.0.0/16"

		By("And Given KCP IpRange exists", func() {
			// Tell IpRange reconciler to ignore this IpRange
			kcpiprange.Ignore.AddName(name)

			Eventually(CreateAwsIpRangeWithSubnets).
				WithArguments(infra.Ctx(), infra.KCP().Client(), awsMock, iprange, vpcId, name, iprangeCidr).
				Should(Succeed(), "failed creating IpRange")
		})

		nfsInstance := &cloudcontrolv1beta1.NfsInstance{}

		By("When NfsInstance is created", func() {
			Eventually(CreateNfsInstance).
				WithArguments(infra.Ctx(), infra.KCP().Client(), nfsInstance,
					WithName(name),
					WithRemoteRef("foo"),
					WithScope(name),
					WithIpRange(name),
					WithNfsInstanceAws(),
				).
				Should(Succeed(), "failed creating NfsInstance")
		})

		var theEfs *efstypes.FileSystemDescription
		By("Then AWS EFS is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), nfsInstance,
					NewObjActions(),
					HavingNfsInstanceStatusId()).
				Should(Succeed(), "expected NfsInstance to get status.id")
			theEfs = awsMock.GetFileSystemById(nfsInstance.Status.Id)
		})

		By("When EFS is Available", func() {
			awsMock.SetFileSystemLifeCycleState(*theEfs.FileSystemId, efstypes.LifeCycleStateAvailable)
		})

		By("Then NfsInstance has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), nfsInstance,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
				).
				Should(Succeed(), "expected NfsInstance to have Ready state, but it didn't")
		})

		By("And Then NfsInstance has status.host set", func() {
			Expect(nfsInstance.Status.Hosts).To(HaveLen(1), "expected one host in NfsInstance.status.hosts")
			Expect(nfsInstance.Status.Hosts[0]).To(Equal(fmt.Sprintf("%s.efs.%s.amazonaws.com", *theEfs.FileSystemId, scope.Spec.Region)))
		})

		By("And Then EFS has mount targets", func() {
			list, err := awsMock.DescribeMountTargets(infra.Ctx(), ptr.Deref(theEfs.FileSystemId, ""))
			Expect(err).NotTo(HaveOccurred(), "failed listing EFS mount targets")
			Expect(list).To(HaveLen(3), "expected 3 EFS mount targets to exist")
			subnetList := pie.Sort(pie.Map(list, func(x efstypes.MountTargetDescription) string {
				return ptr.Deref(x.SubnetId, "")
			}))
			for _, subnet := range iprange.Status.Subnets {
				Expect(subnetList).Should(ContainElement(subnet.Id), fmt.Sprintf("expected mount target in %s subnet with id %s, but its missing", subnet.Zone, subnet.Id))
			}
		})

		// DELETE

		By("When NfsInstance is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), nfsInstance).
				Should(Succeed(), "failed deleting NfsInstance")
		})

		By("And When AWS EFS state is deleted", func() {
			awsMock.SetFileSystemLifeCycleState(ptr.Deref(theEfs.FileSystemId, ""), efstypes.LifeCycleStateDeleted)
		})

		By("Then NfsInstance does not exist", func() {
			Eventually(IsDeleted, 5*time.Second).
				WithArguments(infra.Ctx(), infra.KCP().Client(), nfsInstance).
				Should(Succeed(), "expected NfsInstance not to exist (be deleted), but it still exists")
		})
	})

})
