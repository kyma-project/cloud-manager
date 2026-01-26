/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cloudcontrol

import (
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/elliotchance/pie/v2"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	kcpsubscription "github.com/kyma-project/cloud-manager/pkg/kcp/subscription"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/utils/ptr"
)

var _ = Describe("Feature: VpcNetwork", func() {

	It("Scenario: VpcNetwork AWS is created", func() {
		const subscriptionName = "dd48fd32-7ae9-4fe3-aa24-d66cb1ea06df"
		const vpcNetworkName = "3262da42-3fa7-485f-9487-bc66a5fcacc2"
		const region = "us-east-1"

		subscription := &cloudcontrolv1beta1.Subscription{}
		var vpcNetwork *cloudcontrolv1beta1.VpcNetwork

		awsAccount := infra.AwsMock().NewAccount()

		By("Given AWS Subscription exists", func() {
			kcpsubscription.Ignore.AddName(subscriptionName)
			Expect(
				CreateSubscription(infra.Ctx(), infra, subscription,
					WithName(subscriptionName),
					WithSubscriptionSpecGarden("binding-name")),
			).To(Succeed())

			Expect(
				SubscriptionPatchStatusReadyAws(infra.Ctx(), infra, subscription, awsAccount.AccountId()),
			).To(Succeed())
		})

		By("When VpcNetwork is created", func() {
			vpcNetwork = cloudcontrolv1beta1.NewVpcNetworkBuilder().
				WithRegion(region).
				WithSubscription(subscriptionName).
				WithCidrBlocks("10.250.0.0/16").
				Build()

			Expect(
				CreateObj(infra.Ctx(), infra.KCP().Client(), vpcNetwork, WithName(vpcNetworkName)),
			).To(Succeed())
		})

		By("Then VpcNetwork is ready", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), vpcNetwork, NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady)).
				Should(Succeed())
			cond := meta.FindStatusCondition(*vpcNetwork.Conditions(), cloudcontrolv1beta1.ReasonReady)
			Expect(cond).ToNot(BeNil())
			Expect(cond.Reason).To(Equal(cloudcontrolv1beta1.ReasonProvisioned))
		})

		By("Then VpcNetwork has subscription label", func() {
			Expect(vpcNetwork.Labels[cloudcontrolv1beta1.SubscriptionLabel]).To(Equal(subscriptionName))
		})

		By("Then VpcNetwork has provider label", func() {
			Expect(vpcNetwork.Labels[cloudcontrolv1beta1.SubscriptionLabelProvider]).To(Equal(string(cloudcontrolv1beta1.ProviderAws)))
		})

		By("Then VpcNetwork status has vpcID", func() {
			Expect(vpcNetwork.Status.Identifiers.Vpc).NotTo(BeEmpty())
		})

		awsMock := awsAccount.Region(region)

		var vpc *ec2types.Vpc

		By("Then AWS VPC Network exists", func() {
			aVpc, err := awsMock.DescribeVpc(infra.Ctx(), vpcNetwork.Status.Identifiers.Vpc)
			Expect(err).ToNot(HaveOccurred())
			Expect(aVpc).ToNot(BeNil())
			vpc = aVpc
		})

		By("Then AWS VPC Network has correct CIDR block", func() {
			Expect(pie.Map(vpc.CidrBlockAssociationSet, func(x ec2types.VpcCidrBlockAssociation) string {
				return ptr.Deref(x.CidrBlock, "")
			})).To(Equal(vpcNetwork.Status.CidrBlocks))
		})

		// DELETE ===============================================

		By("When KCP VpcNetwork is deleted", func() {
			Expect(Delete(infra.Ctx(), infra.KCP().Client(), vpcNetwork)).
				To(Succeed())
		})

		By("Then KCP VpcNetwork does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), vpcNetwork).
				Should(Succeed())
		})

		By("Then AWS VPC Network does not exist", func() {
			vpc, err := awsMock.DescribeVpc(infra.Ctx(), ptr.Deref(vpc.VpcId, ""))
			Expect(err).ToNot(HaveOccurred())
			Expect(vpc).To(BeNil())
		})

		By("// cleanup: delete Subscription", func() {
			Expect(infra.KCP().Client().Delete(infra.Ctx(), subscription)).To(Succeed())
		})
	})

})
