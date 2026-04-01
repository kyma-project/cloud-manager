package awswebacl

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
)

func loadWebAcl(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	webAcl := state.ObjAsAwsWebAcl()

	// Skip if no ARN yet (not created)
	if webAcl.Status.Arn == "" {
		return nil, ctx
	}

	logger.Info("Loading AWS WebACL")

	scope := ScopeRegional()

	// List WebACLs to find by name
	summaries, err := state.awsClient.ListWebACLs(ctx, scope)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error listing WebACLs", composed.StopWithRequeue, ctx)
	}

	// Find the WebACL by name
	var foundId string
	for _, summary := range summaries {
		if summary.Name != nil && *summary.Name == webAcl.Name {
			if summary.Id != nil {
				foundId = *summary.Id
			}
			break
		}
	}

	// If not found in list, clear status so it can be recreated
	if foundId == "" {
		logger.Info("WebACL not found in AWS, will recreate")
		webAcl.Status.Arn = ""
		webAcl.Status.Capacity = 0
		state.awsWebAcl = nil
		return nil, ctx
	}

	// Load full WebACL details
	awsWebACL, lockToken, err := state.awsClient.GetWebACL(ctx, webAcl.Name, foundId, scope)
	if err != nil {
		// If not found, clear status so it can be recreated
		if awsmeta.IsNotFound(err) {
			logger.Info("WebACL not found in AWS, will recreate")
			webAcl.Status.Arn = ""
			webAcl.Status.Capacity = 0
			state.awsWebAcl = nil
			return nil, ctx
		}
		return composed.LogErrorAndReturn(err, "Error loading WebACL", composed.StopWithRequeue, ctx)
	}

	// Store in state
	state.awsWebAcl = awsWebACL
	state.lockToken = lockToken
	webAcl.Status.Capacity = awsWebACL.Capacity

	return nil, ctx
}
