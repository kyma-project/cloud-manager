package wafpolicy

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/wafv2"
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

	// Parse JSON from spec.data directly into AWS SDK CreateWebACLInput
	var input wafv2.CreateWebACLInput
	err := json.Unmarshal([]byte(webAcl.Spec.Data), &input)
	if err != nil {
		// JSON unmarshal error is always a configuration error (user must fix spec.data)
		logger.Error(err, "Invalid JSON in spec.data")
		return composed.NewStatusPatcherComposed(webAcl).
			MutateStatus(func(acl *cloudresourcesv1beta1.WafPolicy) {
				acl.SetStatusConfigurationError("Invalid JSON in spec.data: " + err.Error())
			}).
			OnSuccess(composed.Forget).
			OnStatusChanged(composed.Log("WafPolicy ConfigurationError")).
			Run(ctx, state.Cluster().K8sClient())
	}

	// Override immutable fields
	input.Name = aws.String(webAcl.Name)
	input.Scope = ScopeRegional()

	// Add Cloud Manager tags
	input.Tags = convertTags(webAcl, state.Scope())

	// Create WebACL
	err = state.awsClient.CreateWebACL(ctx, &input)
	if err == nil {
		// WebACL created successfully - requeue to reload full details in next loop
		logger.Info("AWS WebACL created successfully, requeuing to reload")
		return composed.StopWithRequeue, ctx
	}

	// Handle AWS API errors
	logger.Error(err, "Error creating WebACL")

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
