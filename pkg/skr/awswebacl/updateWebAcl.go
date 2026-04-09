package awswebacl

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/wafv2"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/utils/ptr"
)

func updateWebAcl(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	webAcl := state.ObjAsAwsWebAcl()

	// Skip if not created yet
	if webAcl.Status.Arn == "" {
		return nil, ctx
	}

	// Check if update is needed
	if !state.updateNeeded {
		return nil, ctx
	}

	logger.Info("Updating AWS WebACL")

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

	// Build UpdateWebACLInput
	input := &wafv2.UpdateWebACLInput{
		Name:             ptr.To(webAcl.Name),
		Id:               state.awsWebAcl.Id,
		Scope:            ScopeRegional(),
		DefaultAction:    defaultAction,
		Rules:            rules,
		VisibilityConfig: visibilityConfig,
		LockToken:        ptr.To(state.lockToken),
	}

	// Add optional fields from spec
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

	// Update WebACL
	err = state.awsClient.UpdateWebACL(ctx, input)

	if err != nil {
		logger.Error(err, "Error updating WebACL")
		return composed.NewStatusPatcherComposed(webAcl).
			MutateStatus(func(acl *cloudresourcesv1beta1.AwsWebAcl) {
				acl.SetStatusProviderError(err.Error())
			}).
			OnSuccess(composed.Requeue).
			Run(ctx, state.Cluster().K8sClient())
	}

	logger.Info("WebACL updated successfully, requeueing to reload")

	// Requeue to reload WebACL with fresh state from AWS
	// On next reconciliation, loadWebAcl will load the updated WebACL
	// and checkUpdateNeeded will return false since AWS now matches spec
	return composed.StopWithRequeue, ctx
}
