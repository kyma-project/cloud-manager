package redisinstance

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

// enableSsl ensures that SSL/TLS encryption is enabled on the AliCloud
// r-kvstore instance. AliCloud instances are created with SSL disabled by
// default; this action enables it and waits for the async operation to
// complete by requeueing when the instance is not yet in Normal status.
func enableSsl(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.instance == nil {
		return nil, ctx
	}

	kcp := state.ObjAsRedisInstance()
	if kcp.Status.Id == "" {
		return nil, ctx
	}

	sslEnabled, err := state.client.DescribeInstanceSSL(ctx, kcp.Status.Id)
	if err != nil {
		return composed.LogErrorAndReturn(err,
			"Error describing AliCloud r-kvstore instance SSL",
			composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
	}

	if sslEnabled {
		return nil, ctx
	}

	if err := state.client.ModifyInstanceSSL(ctx, kcp.Status.Id, true); err != nil {
		return composed.LogErrorAndReturn(err,
			"Error enabling SSL on AliCloud r-kvstore instance",
			composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
	}

	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
}
