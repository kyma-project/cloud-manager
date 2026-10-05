package wafpolicy

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/service/wafv2"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func parsePayload(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	webAcl := state.ObjAsWafPolicy()

	var input wafv2.CreateWebACLInput
	if err := json.Unmarshal([]byte(webAcl.Spec.Payload), &input); err != nil {
		return composed.NewStatusPatcherComposed(webAcl).
			MutateStatus(func(acl *cloudresourcesv1beta1.WafPolicy) {
				acl.SetStatusConfigurationError("Invalid JSON in spec.payload: " + err.Error())
			}).
			OnSuccess(composed.Forget).
			OnStatusChanged(composed.LogError(err, "WafPolicy ConfigurationError")).
			Run(ctx, state.Cluster().K8sClient())
	}

	state.parsedInput = &input
	return nil, ctx
}
