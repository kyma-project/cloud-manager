package criprange

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"time"
)

func deleteKcpIpRange(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if !composed.MarkedForDeletionPredicate(ctx, st) {
		// SKR IpRange is NOT marked for deletion, do not delete mirror in KCP
		return nil, nil
	}

	if state.KcpIpRange == nil {
		// SKR IpRange is marked for deletion, but none found in KCP, probably already deleted
		return nil, nil
	}

	err := state.KcpCluster.K8sClient().Delete(ctx, state.KcpIpRange)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error deleting KCP IpRange", composed.StopWithRequeue, nil)
	}

	// give some time to cloud-control and cloud providers to delete it, and then run again
	return composed.StopWithRequeueDelay(3 * time.Second), nil
}
