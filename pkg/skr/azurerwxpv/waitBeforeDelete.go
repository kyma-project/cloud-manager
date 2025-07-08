package azurerwxpv

import (
	"context"
	"fmt"
	"time"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	azureconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/config"
	corev1 "k8s.io/api/core/v1"
)

func waitBeforeDelete(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	pv := state.ObjAsPV()

	if pv.Status.Phase != corev1.VolumeReleased {
		composed.LoggerFromCtx(ctx).Info("PV not in Released state. Stopping reconciliation", "pv", pv.Name)
		return composed.StopAndForget, ctx
	}

	//Wait for the configured time, if the elapsed time is less than it.
	elapsed := time.Since(pv.Status.LastPhaseTransitionTime.Time)
	timeLeft := azureconfig.AzureConfig.AzureFileShareDeletionWaitDuration - elapsed
	if timeLeft > 0 {
		composed.LoggerFromCtx(ctx).Info(fmt.Sprintf("Waiting for %v to reconcile PV %v", timeLeft, pv.Name))
		return composed.StopWithRequeueDelay(timeLeft), ctx
	}

	//Continue with the reconciliation
	return nil, ctx
}
