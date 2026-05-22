package nuke

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	nuketypes "github.com/kyma-project/cloud-manager/pkg/kcp/nuke/types"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/ptr"
)

func loadWebAcls(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	// Get Cloud Manager Scope (Kyma cluster metadata)
	scopeName := state.Scope().Name

	// List WebACLs with REGIONAL scope type (AWS WAFv2 API parameter - means ALB/API Gateway, not CloudFront)
	// Note: types.ScopeRegional is the WAFv2 scope type, NOT related to state.Scope()
	summaries, err := state.awsClient.ListWebACLs(ctx, types.ScopeRegional)
	if err != nil {
		logger.Error(err, "Error listing AWS WebACLs")
		return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
	}

	// Create resource kind state
	rks := &nuketypes.ProviderResourceKindState{
		Kind:     "AwsWebAcl",
		Provider: cloudcontrolv1beta1.ProviderAws,
		Objects:  []nuketypes.ProviderResourceObject{},
	}

	// Filter and add WebACLs belonging to this scope
	for _, summary := range summaries {
		arn := ptr.Deref(summary.ARN, "")
		if arn == "" {
			continue
		}

		// Get tags for this WebACL
		tags, err := state.awsClient.ListTagsForWebACL(ctx, arn)
		if err != nil {
			logger.Error(err, "Error getting tags for WebACL", "arn", arn)
			continue // Skip this WebACL but continue processing others
		}

		// Check if this WebACL belongs to our scope
		if hasTag(tags, common.TagScope, scopeName) {
			rks.Objects = append(rks.Objects, AwsWebAclResource{Summary: summary})
			logger.Info("Found WebACL for scope", "arn", arn, "scope", scopeName)
		}
	}

	state.ProviderResources = append(state.ProviderResources, rks)

	logger.Info("Loaded AWS WebACLs for nuke", "count", len(rks.Objects), "scope", scopeName)
	return nil, nil
}
