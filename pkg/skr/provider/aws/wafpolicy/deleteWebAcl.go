package wafpolicy

import (
	"context"
	"strings"

	wafv2types "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
)

func deleteWebAcl(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.awsWebAcl == nil || state.awsWebAcl.Id == nil {
		return nil, ctx
	}

	// Delete WebACL
	err := state.awsClient.DeleteWebACL(ctx, *state.awsWebAcl.Name, *state.awsWebAcl.Id, wafv2types.ScopeRegional, state.lockToken)

	if awsmeta.IsNotFound(err) {
		return nil, ctx
	}

	// If WebACL is still associated with resources, set DeletionBlockedByDependents condition
	if isWebAclAssociatedError(err) {
		return composed.NewStatusPatcherComposed(state.ObjAsWafPolicy()).
			MutateStatus(func(acl *cloudresourcesv1beta1.WafPolicy) {
				acl.SetStatusDeletionBlockedByDependents("WebACL is still associated with AWS resources. Remove all associations before deleting.")
			}).
			OnSuccess(composed.Requeue).
			OnStatusChanged(composed.Log("WafPolicy DeletionBlockedByDependents")).
			OnFailure(composed.Log("WafPolicy failed to patch status DeletionBlockedByDependents")).
			Run(ctx, state.Cluster().K8sClient())
	}

	// Handle other errors
	if err != nil {
		logger.Error(err, "Error deleting WebACL")

		// Configuration errors (permissions, invalid state)
		if isConfigurationError(err) {
			return composed.NewStatusPatcherComposed(state.ObjAsWafPolicy()).
				MutateStatus(func(acl *cloudresourcesv1beta1.WafPolicy) {
					acl.SetStatusConfigurationError(err.Error())
				}).
				OnSuccess(composed.Forget).
				OnStatusChanged(composed.Log("WafPolicy ConfigurationError")).
				OnFailure(composed.Log("WafPolicy failed to patch ConfigurationError")).
				Run(ctx, state.Cluster().K8sClient())
		}

		// Retryable errors (throttling, temporary issues)
		if awsmeta.IsErrorRetryable(err) {
			return composed.NewStatusPatcherComposed(state.ObjAsWafPolicy()).
				MutateStatus(func(acl *cloudresourcesv1beta1.WafPolicy) {
					acl.SetStatusProviderError(err.Error())
				}).
				OnSuccess(composed.Requeue).
				OnStatusChanged(composed.Log("WafPolicy Error (retryable)")).
				OnFailure(composed.Log("WafPolicy failed to patch status ProviderError")).
				Run(ctx, state.Cluster().K8sClient())
		}

		// Terminal non-retryable errors
		return composed.NewStatusPatcherComposed(state.ObjAsWafPolicy()).
			MutateStatus(func(acl *cloudresourcesv1beta1.WafPolicy) {
				acl.SetStatusFailure(err.Error())
			}).
			OnSuccess(composed.Forget).
			OnStatusChanged(composed.Log("WafPolicy Failure")).
			OnFailure(composed.Log("WafPolicy failed to patch status Failure")).
			Run(ctx, state.Cluster().K8sClient())
	}

	logger.Info("WebACL deleted successfully")
	return nil, ctx
}

func isWebAclAssociatedError(err error) bool {
	if err == nil {
		return false
	}
	// AWS WAFv2 returns errors with "associated" in message when WebACL is still in use
	return strings.Contains(err.Error(), "associated") ||
		strings.Contains(err.Error(), "WAFAssociatedItemException")
}
