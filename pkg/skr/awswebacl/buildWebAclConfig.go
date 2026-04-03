package awswebacl

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func buildWebAclConfig(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	webAcl := state.ObjAsAwsWebAcl()

	// Convert spec to AWS types
	defaultAction, err := convertDefaultAction(webAcl.Spec.DefaultAction)
	if err != nil {
		return composed.NewStatusPatcherComposed(webAcl).
			MutateStatus(func(acl *cloudresourcesv1beta1.AwsWebAcl) {
				acl.SetStatusProviderError(err.Error())
			},
			).Run(ctx, state.Cluster().K8sClient())
	}

	rules, err := convertRules(webAcl.Spec.Rules)
	if err != nil {
		return composed.NewStatusPatcherComposed(webAcl).
			MutateStatus(func(acl *cloudresourcesv1beta1.AwsWebAcl) {
				acl.SetStatusProviderError(err.Error())
			},
			).Run(ctx, state.Cluster().K8sClient())
	}

	visibilityConfig := convertVisibilityConfig(webAcl.Spec.VisibilityConfig, webAcl.Name)

	// Store in state
	state.defaultAction = defaultAction
	state.rules = rules
	state.visibilityConfig = visibilityConfig

	return nil, ctx
}
