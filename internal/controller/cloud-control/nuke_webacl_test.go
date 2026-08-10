package cloudcontrol

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/wafv2"
	wafv2types "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	"github.com/kyma-project/cloud-manager/api"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	kcpscope "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	kcpvpcpeering "github.com/kyma-project/cloud-manager/pkg/kcp/vpcpeering"

	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Feature: Nuke AWS WebACL", func() {

	It("Scenario: KCP Nuke deletes AWS WebACL resources", func() {
		const kymaName = "6ff4e8g3-ge8f-5829-bce3-gg0g6894g1ge"

		awsAccount := infra.AwsMock().NewAccount()

		scope := &cloudcontrolv1beta1.Scope{}

		By("Given AWS Scope exists", func() {
			kcpscope.Ignore.AddName(kymaName)

			Expect(CreateScopeAws(infra.Ctx(), infra, scope, awsAccount.AccountId(), WithName(kymaName))).
				To(Succeed(), "failed creating AWS scope")
		})

		// Nuke reconciler is not triggered if there are no resources to delete see [resourceStatusDeleting]
		vpcPeering := cloudcontrolv1beta1.NewVpcPeeringBuilder().
			WithName("ad249d39-795a-44c8-9997-2344a3fe16f5").
			WithScope(kymaName).
			WithRemoteRef(DefaultSkrNamespace, "name").
			WithAzurePeering("remotePeeringName", "remoteVNet", "remoteResourceGroup", false).
			Build()

		By("And Given VpcPeering exists", func() {
			kcpvpcpeering.Ignore.AddName(vpcPeering.Name)

			Expect(CreateObj(infra.Ctx(), infra.KCP().Client(), vpcPeering,
				AddFinalizer(api.CommonFinalizerDeletionHook),
			)).To(Succeed(), "failed creating VpcPeering")
		})

		awsRegion := awsAccount.Region(scope.Spec.Region)
		aclName := "my-web-acl"

		var webAcl *wafv2types.WebACL
		By("And Given WebACL exists in AWS mock with scope tags", func() {
			err := awsRegion.CreateWebACL(infra.Ctx(), &wafv2.CreateWebACLInput{
				Name: &aclName,
				Tags: []wafv2types.Tag{
					{
						Key:   aws.String("Name"),
						Value: &aclName,
					},
					{
						Key:   aws.String("ManagedBy"),
						Value: aws.String("cloud-manager"),
					},
					{
						Key:   aws.String(common.TagScope),
						Value: aws.String(scope.Name),
					},
					{
						Key:   aws.String(common.TagShoot),
						Value: aws.String(scope.Spec.ShootName),
					},
				},
			})

			Expect(err).To(Succeed())

			// Retrieve the created WebACL
			webAcl, _, err = awsRegion.GetWebACL(infra.Ctx(), aclName, "", wafv2types.ScopeRegional)
			Expect(err).To(Succeed())
		})

		nuke := &cloudcontrolv1beta1.Nuke{}

		By("When Nuke for the Scope is created", func() {
			Expect(CreateObj(infra.Ctx(), infra.KCP().Client(), nuke,
				WithName("nuke-aws-webacl-test"),
				WithScope(kymaName),
			)).To(Succeed())
		})

		By("Then Nuke status state is Deleting", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), nuke, NewObjActions(),
					HavingState("Deleting"),
				).
				Should(Succeed())
		})

		By("When resource VpcPeering finalizer is removed", func() {
			removed, err := composed.PatchObjRemoveFinalizer(infra.Ctx(), api.CommonFinalizerDeletionHook, vpcPeering, infra.KCP().Client())
			Expect(err).To(Succeed())
			Expect(removed).To(BeTrue())
		})

		By("Then resource VpcPeering does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), vpcPeering).
				Should(Succeed())
		})

		By("And Then Nuke status shows AwsWebAcl resource discovered", func() {
			Eventually(func() error {
				if err := LoadAndCheck(infra.Ctx(), infra.KCP().Client(), nuke, NewObjActions()); err != nil {
					return err
				}
				sk := nuke.Status.GetKindNoCreate("AwsWebAcl")
				if sk == nil {
					return fmt.Errorf("AwsWebAcl kind not found in Nuke status")
				}
				if len(sk.Objects) == 0 {
					return fmt.Errorf("no AwsWebAcl objects found")
				}
				return nil
			}).Should(Succeed())
		})

		By("And Then WebACL is deleted from AWS mock", func() {
			Eventually(func() bool {

				_, _, err := awsRegion.GetWebACL(infra.Ctx(), *webAcl.Name, *webAcl.ARN, wafv2types.ScopeRegional)

				if awsmeta.IsNotFound(err) {
					return true
				}

				return false

			}).Should(BeTrue())
		})

		By("And Then Nuke status shows AwsWebAcl as Deleted", func() {
			Eventually(func() error {
				if err := LoadAndCheck(infra.Ctx(), infra.KCP().Client(), nuke, NewObjActions()); err != nil {
					return err
				}
				sk := nuke.Status.GetKindNoCreate("AwsWebAcl")
				if sk == nil {
					return fmt.Errorf("AwsWebAcl kind not found in Nuke status")
				}
				status := sk.Objects[*webAcl.ARN]
				if status != cloudcontrolv1beta1.NukeResourceStatusDeleted {
					return fmt.Errorf("expected AwsWebAcl status Deleted, got %s", status)
				}
				return nil
			}).Should(Succeed())
		})

		By("And Then Nuke status state is Completed", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), nuke, NewObjActions(),
					HavingState("Completed"),
				).Should(Succeed())
		})

		By("And Then Scope is deleted", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), scope).
				Should(Succeed())
		})

		By("// cleanup: Delete Nuke", func() {
			Expect(Delete(infra.Ctx(), infra.KCP().Client(), nuke)).
				To(Succeed())
		})
	})
})
