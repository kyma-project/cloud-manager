package rediscluster

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

// enableSsl ensures that SSL/TLS encryption is enabled on the AliCloud
// r-kvstore cluster. AliCloud instances are created with SSL disabled by
// default; this action calls ModifyInstanceSSL and requeues — the instance
// transitions to SSLModifying, and waitRedisAvailable (next pipeline step)
// gates further progress until it returns to Normal.
func enableSsl(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.instance == nil {
		return nil, ctx
	}

	sslEnabled, err := state.client.DescribeInstanceSSL(ctx, state.instance.InstanceId)
	if err != nil {
		return composed.LogErrorAndReturn(err,
			"Error describing AliCloud r-kvstore cluster SSL",
			composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
	}

	if sslEnabled {
		return nil, ctx
	}

	if err := state.client.ModifyInstanceSSL(ctx, state.instance.InstanceId, true); err != nil {
		return composed.LogErrorAndReturn(err,
			"Error enabling SSL on AliCloud r-kvstore cluster",
			composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
	}

	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
}
