package wafpolicy

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
)

func loadLoggingConfig(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.awsWebAcl == nil || state.awsWebAcl.ARN == nil {
		return nil, ctx
	}

	cfg, err := state.awsClient.GetLoggingConfiguration(ctx, *state.awsWebAcl.ARN)
	if err != nil {
		if awsmeta.IsNotFound(err) {
			state.loggingConfig = nil
			return nil, ctx
		}
		return composed.LogErrorAndReturn(err, "Error loading WebACL logging configuration", composed.StopWithRequeue, ctx)
	}

	state.loggingConfig = cfg
	return nil, ctx
}
