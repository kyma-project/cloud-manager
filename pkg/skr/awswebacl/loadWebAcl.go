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

	// Load WebACL from AWS
	id := extractIdFromArn(webAcl.Status.Arn)
	awsWebACL, lockToken, err := state.awsClient.GetWebACL(ctx, webAcl.Name, id, scope)
	if err != nil {
		// If not found, clear status so it can be recreated
		if awsmeta.IsNotFound(err) {
			logger.Info("WebACL not found in AWS, will recreate")
			webAcl.Status.Arn = ""
			webAcl.Status.Capacity = 0
			return nil, ctx
		}
		return composed.LogErrorAndReturn(err, "Error loading WebACL", composed.StopWithRequeue, ctx)
	}

	// Store lock token in state (transient, not persisted)
	state.lockToken = lockToken
	webAcl.Status.Capacity = awsWebACL.Capacity

	return nil, ctx
}
