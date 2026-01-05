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

	// Skip if instance doesn't exist
	if instance == nil {
		return nil, nil
	}

	// Check instance state
	switch instance.State {
	case filestorepb.Instance_READY:
		// Instance is ready, continue
		return nil, nil

	case filestorepb.Instance_CREATING:
		// Still creating, requeue
		logger.Info("Instance is creating, waiting")
		return composed.StopWithRequeueDelay(config.GcpConfig.GcpRetryWaitTime), nil

	case filestorepb.Instance_ERROR:
		// Instance in error state, continue to status update
		logger.Info("Instance in error state")
		return nil, nil

	default:
		// Unknown state, requeue
		logger.Info("Instance in unknown state, waiting", "state", instance.State)
		return composed.StopWithRequeueDelay(config.GcpConfig.GcpRetryWaitTime), nil
	}
}
