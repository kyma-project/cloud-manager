package v2

import (
	"context"

	"cloud.google.com/go/filestore/apiv1/filestorepb"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/config"
)

// waitInstanceReady waits for the Filestore instance to reach READY state.
func waitInstanceReady(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	instance := state.GetInstance()

	if instance == nil {
		return nil, ctx
	}

	switch instance.State {
	case filestorepb.Instance_READY:
		return nil, ctx

	case filestorepb.Instance_CREATING:
		logger.Info("Instance is creating, waiting")
		return composed.StopWithRequeueDelay(config.GcpConfig.GcpRetryWaitTime), nil

	case filestorepb.Instance_ERROR:
		logger.Info("Instance in error state")
		return nil, ctx

	default:
		logger.Info("Instance in unknown state, waiting", "state", instance.State)
		return composed.StopWithRequeueDelay(config.GcpConfig.GcpRetryWaitTime), nil
	}
}
