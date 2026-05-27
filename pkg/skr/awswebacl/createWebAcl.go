package awswebacl

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/wafv2"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func createWebAcl(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	webAcl := state.ObjAsAwsWebAcl()

	// Skip if already exists
	if state.awsWebAcl != nil {
		return nil, ctx
	}

	logger.Info("Creating AWS WebACL")

	// Convert spec to AWS types
	defaultAction, err := convertDefaultAction(webAcl.Spec.DefaultAction)
	if err != nil {
		return composed.NewStatusPatcherComposed(webAcl).
			MutateStatus(func(acl *cloudresourcesv1beta1.AwsWebAcl) {
				acl.SetStatusProviderError(err.Error())
			}).
			OnSuccess(composed.Requeue).
			Run(ctx, state.Cluster().K8sClient())
	}

	rules, err := convertRules(webAcl.Spec.Rules)
	if err != nil {
		return composed.NewStatusPatcherComposed(webAcl).
			MutateStatus(func(acl *cloudresourcesv1beta1.AwsWebAcl) {
				acl.SetStatusProviderError(err.Error())
			}).
			OnSuccess(composed.Requeue).
			Run(ctx, state.Cluster().K8sClient())
	}

	visibilityConfig := convertVisibilityConfig(webAcl.Spec.VisibilityConfig, webAcl.Name)

	// Build CreateWebACLInput
	input := &wafv2.CreateWebACLInput{
		Name:             aws.String(webAcl.Name),
		Description:      aws.String(webAcl.Spec.Description),
		Scope:            ScopeRegional(),
		DefaultAction:    defaultAction,
		Rules:            rules,
		VisibilityConfig: visibilityConfig,
		Tags:             convertTags(webAcl, state.Scope()),
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

	// Create WebACL
	err = state.awsClient.CreateWebACL(ctx, input)
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

	// WebACL created successfully - requeue to reload full details in next loop
	logger.Info("AWS WebACL created successfully, requeuing to reload")

	return composed.StopWithRequeue, ctx
}
