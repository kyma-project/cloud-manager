package awswebacl

import (
	"context"
	"strings"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
)

func deleteWebAcl(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	webAcl := state.ObjAsAwsWebAcl()

	// Skip if never created
	if webAcl.Status.Arn == "" {
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
	if err != nil {
		// If not found, consider it deleted
		if awsmeta.IsNotFound(err) {
			logger.Info("WebACL not found in AWS, considering as deleted")
			return nil, ctx
		}

		// If WebACL is still associated with resources, set DeleteWhileUsed condition
		if isWebAclAssociatedError(err) {
			logger.Error(err, "WebACL is still associated with AWS resources, cannot delete")
			return composed.NewStatusPatcherComposed(webAcl).
				MutateStatus(func(acl *cloudresourcesv1beta1.AwsWebAcl) {
					acl.SetStatusDeleteWhileUsed("WebACL is still associated with AWS resources. Remove all associations before deleting.")
				}).
				OnSuccess(composed.Requeue).
				OnStatusChanged(composed.Log("AwsWebAcl DeleteWhileUsed")).
				Run(ctx, state.Cluster().K8sClient())
		}

		logger.Error(err, "Error deleting WebACL")

		return composed.NewStatusPatcherComposed(webAcl).
			MutateStatus(func(acl *cloudresourcesv1beta1.AwsWebAcl) {
				acl.SetStatusProviderError(err.Error())
			}).
			OnSuccess(composed.Requeue).
			OnStatusChanged(composed.Log("AwsWebAcl ProviderError")).
			Run(ctx, state.Cluster().K8sClient())
	}

	// If deletion succeeded and we had a DeleteWhileUsed condition, remove it
	if webAcl.Status.State == cloudresourcesv1beta1.ReasonDeleteWhileUsed {
		logger.Info("WebACL is no longer associated, removing DeleteWhileUsed condition")
		return composed.NewStatusPatcherComposed(webAcl).
			MutateStatus(func(acl *cloudresourcesv1beta1.AwsWebAcl) {
				acl.RemoveStatusDeleteWhileUsed()
			}).
			OnSuccess(composed.Continue).
			OnFailure(composed.Log("Failed to remove DeleteWhileUsed condition")).
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
