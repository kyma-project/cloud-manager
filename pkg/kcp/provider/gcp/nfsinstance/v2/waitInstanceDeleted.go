package v2

import (
	"context"

	"cloud.google.com/go/filestore/apiv1/filestorepb"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/config"
)

// waitInstanceDeleted waits for the Filestore instance to be deleted.
func waitInstanceDeleted(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	instance := state.GetInstance()

	// If instance doesn't exist, deletion is complete
	if instance == nil {
		return nil, nil
	}

	// If instance is deleting, wait
	if instance.State == filestorepb.Instance_DELETING {
		logger.Info("Instance is deleting, waiting")
		return composed.StopWithRequeueDelay(config.GcpConfig.GcpRetryWaitTime), nil
	}

	// Instance exists but not deleting - this shouldn't happen
	// Continue anyway, deleteInstance will be called again next reconcile
	return nil, nil
}
