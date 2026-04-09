package awswebacl

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/wafv2"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/utils/ptr"
)

func createWebAcl(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	webAcl := state.ObjAsAwsWebAcl()

	// Skip if already exists
	if webAcl.Status.Arn != "" {
		return nil, ctx
	}

	logger.Info("Creating AWS WebACL")

	// Build CreateWebACLInput
	input := &wafv2.CreateWebACLInput{
		Name:             ptr.To(webAcl.Name),
		Description:      ptr.To(webAcl.Spec.Description),
		Scope:            ScopeRegional(),
		DefaultAction:    state.defaultAction,
		Rules:            state.rules,
		VisibilityConfig: state.visibilityConfig,
		Tags:             convertTags(webAcl),
	}

	// Add optional fields from state
	if len(webAcl.Spec.TokenDomains) > 0 {
		input.TokenDomains = webAcl.Spec.TokenDomains
	}

	if len(webAcl.Spec.CustomResponseBodies) > 0 {
		input.CustomResponseBodies = convertCustomResponseBodies(webAcl.Spec.CustomResponseBodies)
	}

	if webAcl.Spec.CaptchaConfig != nil {
		input.CaptchaConfig = convertCaptchaConfig(webAcl.Spec.CaptchaConfig)
	}

	if webAcl.Spec.ChallengeConfig != nil {
		input.ChallengeConfig = convertChallengeConfig(webAcl.Spec.ChallengeConfig)
	}

	if webAcl.Spec.AssociationConfig != nil {
		input.AssociationConfig = convertAssociationConfig(webAcl.Spec.AssociationConfig)
	}

	// Create WebACL
	createdWebACL, lockToken, err := state.awsClient.CreateWebACL(ctx, input)
	if err != nil {
		logger.Error(err, "Error creating WebACL")

		return composed.NewStatusPatcherComposed(webAcl).
			MutateStatus(func(acl *cloudresourcesv1beta1.AwsWebAcl) {
				acl.SetStatusProviderError(err.Error())
			}).
			OnSuccess(composed.Requeue).
			OnStatusChanged(composed.Log("AwsWebAcl ProviderError")).
			Run(ctx, state.Cluster().K8sClient())
	}

	// Store WebACL and lock token in state (transient, not persisted)
	state.awsWebAcl = createdWebACL
	state.lockToken = lockToken

	logger.Info("AWS WebACL created successfully")

	return nil, ctx
}
