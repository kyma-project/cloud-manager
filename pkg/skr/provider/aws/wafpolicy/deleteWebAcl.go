package wafpolicy

import (
	"context"
	"strings"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"k8s.io/apimachinery/pkg/api/meta"
)

func deleteWebAcl(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	webAcl := state.ObjAsWafPolicy()

	// Skip if never created
	if webAcl.Status.ProviderId == "" {
		return nil, ctx
	}

	logger.Info("Deleting AWS WebACL")

	scope := ScopeRegional()

	// Get ID from loaded WebACL in state
	var id string
	if state.awsWebAcl != nil && state.awsWebAcl.Id != nil {
		id = *state.awsWebAcl.Id
	} else {
		// If WebACL not loaded (e.g., manual deletion), try to find it by listing
		summaries, err := state.awsClient.ListWebACLs(ctx, scope)
		if err != nil {
			if awsmeta.IsNotFound(err) {
				logger.Info("WebACL not found in AWS, considering as deleted")
				return nil, ctx
			}
			return composed.LogErrorAndReturn(err, "Error listing WebACLs for deletion", composed.StopWithRequeue, ctx)
		}

		// Find by name
		for _, summary := range summaries {
			if summary.Name != nil && *summary.Name == webAcl.Name && summary.Id != nil {
				id = *summary.Id
				break
			}
		}

		if id == "" {
			logger.Info("WebACL not found in AWS, considering as deleted")
			return nil, ctx
		}
	}

	// Delete WebACL
	err := state.awsClient.DeleteWebACL(ctx, webAcl.Name, id, scope, state.lockToken)

	// If not found, consider it deleted
	if awsmeta.IsNotFound(err) {
		logger.Info("WebACL not found in AWS, considering as deleted")
		return nil, ctx
	}

	// If WebACL is still associated with resources, set DeleteWhileUsed condition
	if isWebAclAssociatedError(err) {
		logger.Error(err, "WebACL is still associated with AWS resources, cannot delete")
		return composed.NewStatusPatcherComposed(webAcl).
			MutateStatus(func(acl *cloudresourcesv1beta1.WafPolicy) {
				acl.SetStatusDeleteWhileUsed("WebACL is still associated with AWS resources. Remove all associations before deleting.")
			}).
			OnSuccess(composed.Requeue).
			OnStatusChanged(composed.Log("WafPolicy DeleteWhileUsed")).
			Run(ctx, state.Cluster().K8sClient())
	}

	// Handle other errors
	if err != nil {
		logger.Error(err, "Error deleting WebACL")

		// Configuration errors (permissions, invalid state)
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

	// Deletion succeeded - if we had a DeleteWhileUsed reason, clear it
	readyCondition := meta.FindStatusCondition(webAcl.Status.Conditions, cloudresourcesv1beta1.ConditionTypeReady)
	if readyCondition != nil && readyCondition.Reason == cloudresourcesv1beta1.ReasonDeleteWhileUsed {
		logger.Info("WebACL is no longer associated, clearing DeleteWhileUsed state")
		return composed.NewStatusPatcherComposed(webAcl).
			MutateStatus(func(acl *cloudresourcesv1beta1.WafPolicy) {
				acl.RemoveStatusDeleteWhileUsed()
			}).
			OnSuccess(composed.Continue).
			OnFailure(composed.Log("Failed to clear DeleteWhileUsed state")).
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
