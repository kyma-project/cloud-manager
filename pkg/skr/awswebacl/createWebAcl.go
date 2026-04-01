package awswebacl

import (
	"context"
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

	// Convert spec to AWS types
	defaultAction, err := convertDefaultAction(webAcl.Spec.DefaultAction)
	if err != nil {
		return composed.NewStatusPatcherComposed(webAcl).
			MutateStatus(func(acl *cloudresourcesv1beta1.AwsWebAcl) {
				acl.SetStatusProviderError(err.Error())
			}).
			OnSuccess(
				composed.LogError(err, "Provider error on SKR AWS WebACL create/update"),
				composed.Requeue).
			Run(ctx, state.Cluster().K8sClient())
	}

	rules, err := convertRules(webAcl.Spec.Rules)

	if err != nil {
		return composed.NewStatusPatcherComposed(webAcl).
			MutateStatus(func(acl *cloudresourcesv1beta1.AwsWebAcl) {
				acl.SetStatusProviderError(err.Error())
			}).
			OnSuccess(
				composed.LogError(err, "Provider error on SKR AWS WebACL create/update"),
				composed.Requeue).
			Run(ctx, state.Cluster().K8sClient())
	}

	visibilityConfig := convertVisibilityConfig(webAcl.Spec.VisibilityConfig, webAcl.Name)

	// Determine scope
	scope := ScopeRegional()
	if err != nil {
		return composed.NewStatusPatcherComposed(webAcl).
			MutateStatus(func(acl *cloudresourcesv1beta1.AwsWebAcl) {
				acl.SetStatusProviderError(err.Error())
			}).
			OnSuccess(
				composed.LogError(err, "Provider error on SKR AWS WebACL create/update"),
				composed.Requeue).
			Run(ctx, state.Cluster().K8sClient())
	}

	// Create WebACL
	createdWebACL, lockToken, err := state.awsClient.CreateWebACL(
		ctx,
		webAcl.Name,
		webAcl.Spec.Description,
		scope,
		defaultAction,
		rules,
		visibilityConfig,
		convertTags(webAcl),
	)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error creating WebACL", composed.StopWithRequeue, ctx)
	}

	// Store WebACL and lock token in state (transient, not persisted)
	state.awsWebAcl = createdWebACL
	state.lockToken = lockToken

	return composed.NewStatusPatcherComposed(webAcl).
		MutateStatus(func(acl *cloudresourcesv1beta1.AwsWebAcl) {
			acl.Status.Arn = ptr.Deref(createdWebACL.ARN, "")
			acl.Status.Capacity = createdWebACL.Capacity
			acl.SetStatusReady()
		}).
		OnSuccess(
			// log only if something was created/updated
			composed.Log("AWS SKR WebACL is successfully created"),
		).
		Run(ctx, state.Cluster().K8sClient())
}
