package wafpolicy

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/wafv2"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
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
		return composed.NewStatusPatcherComposed(webAcl).
			MutateStatus(func(acl *cloudresourcesv1beta1.WafPolicy) {
				acl.SetStatusProviderError(err.Error())
			}).
			OnSuccess(composed.Requeue).
			Run(ctx, state.Cluster().K8sClient())
	}

	// Override immutable fields
	input.Name = aws.String(webAcl.Name)
	input.Scope = ScopeRegional()

	// Add Cloud Manager tags
	input.Tags = convertTags(webAcl, state.Scope())

	// Create WebACL
	err = state.awsClient.CreateWebACL(ctx, &input)
	if err != nil {
		logger.Error(err, "Error creating WebACL")

		return composed.NewStatusPatcherComposed(webAcl).
			MutateStatus(func(acl *cloudresourcesv1beta1.WafPolicy) {
				acl.SetStatusProviderError(err.Error())
			}).
			OnSuccess(composed.Requeue).
			OnStatusChanged(composed.Log("WafPolicy ProviderError")).
			Run(ctx, state.Cluster().K8sClient())
	}

	// WebACL created successfully - requeue to reload full details in next loop
	logger.Info("AWS WebACL created successfully, requeuing to reload")

	return composed.StopWithRequeue, ctx
}
