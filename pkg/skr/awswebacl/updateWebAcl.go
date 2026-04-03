package awswebacl

import (
	"context"
	"fmt"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"

	"github.com/kyma-project/cloud-manager/pkg/composed"
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

	scope := ScopeRegional()

	// Get ID from loaded WebACL in state
	if state.awsWebAcl == nil || state.awsWebAcl.Id == nil {
		return composed.LogErrorAndReturn(
			fmt.Errorf("WebACL not loaded in state"),
			"Cannot update WebACL without loaded state",
			composed.StopWithRequeue,
			ctx,
		)
	}

	// Update WebACL using config from state
	err := state.awsClient.UpdateWebACL(
		ctx,
		webAcl.Name,
		*state.awsWebAcl.Id,
		scope,
		state.defaultAction,
		state.rules,
		state.visibilityConfig,
		state.lockToken,
	)
	if err != nil {
		logger.Error(err, "Error updating WebACL")
		if err != nil {
			return composed.NewStatusPatcherComposed(webAcl).
				MutateStatus(func(acl *cloudresourcesv1beta1.AwsWebAcl) {
					acl.SetStatusProviderError(err.Error())
				}).
				OnSuccess(composed.Requeue).
				Run(ctx, state.Cluster().K8sClient())
		}
	}

	logger.Info("WebACL updated successfully")

	return nil, ctx
}
