package awswebacl

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
)

func deleteWebAcl(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	webAcl := state.ObjAsAwsWebAcl()

	// Skip if never created
	if webAcl.Status.Arn == "" {
		return nil, ctx
	}

	logger.Info("Deleting AWS WebACL")

	scope := ScopeRegional()

	// Delete WebACL
	id := extractIdFromArn(webAcl.Status.Arn)
	err := state.awsClient.DeleteWebACL(ctx, webAcl.Name, id, scope, state.lockToken)
	if err != nil {
		// If not found, consider it deleted
		if awsmeta.IsNotFound(err) {
			logger.Info("WebACL not found in AWS, considering as deleted")
			return nil, ctx
		}
		return composed.LogErrorAndReturn(err, "Error deleting WebACL", composed.StopWithRequeue, ctx)
	}

	logger.Info("WebACL deleted successfully")

	return nil, ctx
}
