package wafpolicy

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	wafv2types "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
)

func createWebAcl(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	webAcl := state.ObjAsWafPolicy()

	// Skip if already exists
	if state.awsWebAcl != nil {
		return nil, ctx
	}

	logger.Info("Creating AWS WebACL")

	input := *state.parsedInput
	input.Name = aws.String(webAcl.Status.Id)
	input.Scope = wafv2types.ScopeRegional
	input.Tags = state.convertTags()

	// Create WebACL
	err := state.awsClient.CreateWebACL(ctx, &input)
	if err == nil {
		logger.Info("AWS WebACL created successfully, requeuing to reload")
		return composed.StopWithRequeue, ctx
	}

	// User-actionable configuration errors (invalid rules, permissions, etc)
	if isConfigurationError(err) {
		return composed.NewStatusPatcherComposed(webAcl).
			MutateStatus(func(acl *cloudresourcesv1beta1.WafPolicy) {
				acl.SetStatusConfigurationError(err.Error())
			}).
			OnSuccess(composed.Forget).
			OnStatusChanged(composed.Log("WafPolicy ConfigurationError")).
			Run(ctx, state.Cluster().K8sClient())
	}

	// Retryable errors (throttling, temporary issues)
	if awsmeta.IsErrorRetryable(err) {
		return composed.NewStatusPatcherComposed(webAcl).
			MutateStatus(func(acl *cloudresourcesv1beta1.WafPolicy) {
				acl.SetStatusProviderError(err.Error())
			}).
			OnSuccess(composed.Requeue).
			OnStatusChanged(composed.Log("WafPolicy Error (retryable)")).
			Run(ctx, state.Cluster().K8sClient())
	}

	// Terminal non-retryable errors
	return composed.NewStatusPatcherComposed(webAcl).
		MutateStatus(func(acl *cloudresourcesv1beta1.WafPolicy) {
			acl.SetStatusFailure(err.Error())
		}).
		OnSuccess(composed.Forget).
		OnStatusChanged(composed.Log("WafPolicy Failure")).
		Run(ctx, state.Cluster().K8sClient())
}
