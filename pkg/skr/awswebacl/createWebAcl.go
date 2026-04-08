package awswebacl

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
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

	// Determine scope
	scope := ScopeRegional()

	// Create WebACL using config from state
	createdWebACL, lockToken, err := state.awsClient.CreateWebACL(
		ctx,
		webAcl.Name,
		webAcl.Spec.Description,
		scope,
		state.defaultAction,
		state.rules,
		state.visibilityConfig,
		convertTags(webAcl),
	)
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
