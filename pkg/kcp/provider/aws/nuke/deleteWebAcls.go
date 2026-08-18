package nuke

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"k8s.io/utils/ptr"
)

func deleteWebAcls(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	for _, rks := range state.ProviderResources {
		if rks.Kind == "AwsWebAcl" && rks.Provider == cloudcontrolv1beta1.ProviderAws {
			for _, obj := range rks.Objects {
				resource := obj.(AwsWebAclResource)

				// Get full WebACL details to obtain lock token
				name := ptr.Deref(resource.Summary.Name, "")
				id := ptr.Deref(resource.Summary.Id, "")

				// Note: types.ScopeRegional is the WAFv2 scope type (ALB/API Gateway), not related to state.Scope()
				_, lockToken, err := state.awsClient.GetWebACL(ctx, name, id, types.ScopeRegional)
				if err != nil {
					if awsmeta.IsNotFound(err) {
						logger.Info("WebACL already deleted", "id", resource.GetId())
						continue
					}
					logger.Error(err, "Error getting WebACL for deletion", "id", resource.GetId())
					continue // non-fatal
				}

				// Delete WebACL with lock token
				err = state.awsClient.DeleteWebACL(ctx, name, id, types.ScopeRegional, lockToken)
				if err != nil {
					if awsmeta.IsNotFound(err) {
						logger.Info("WebACL already deleted during deletion call", "id", resource.GetId())
						continue
					}
					logger.Error(err, "Error requesting AWS WebACL deletion", "id", resource.GetId())
					// Continue anyway (non-fatal error pattern)
				} else {
					logger.Info("Requested AWS WebACL deletion", "id", resource.GetId(), "name", name)
				}
			}
		}
	}

	return nil, nil
}
