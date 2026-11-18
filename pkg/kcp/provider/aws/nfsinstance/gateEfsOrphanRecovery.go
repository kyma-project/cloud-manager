package nfsinstance

import (
	"context"
	"time"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/config"
)

// gateEfsOrphanRecovery parks the NfsInstance reconciler for the configured shoot while EFS
// orphan recovery is active (mode dry-run|on). This frees the shared worker pool from the stuck
// shoot's throttled AWS calls so the healthy fleet is unaffected. When mode is off (the shoot id
// is gone from config) the gate stops matching and normal create/wait resumes idempotently.
//
// Disposable: remove together with the recovery garbage collector after close-out.
func gateEfsOrphanRecovery(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	rc := awsconfig.AwsConfig.EfsOrphanRecovery

	if rc.Mode == awsconfig.EfsOrphanRecoveryModeOff {
		return nil, ctx
	}
	if state.Scope().Spec.ShootName != rc.ShootId {
		return nil, ctx
	}

	composed.LoggerFromCtx(ctx).Info(
		"EFS orphan recovery active for this shoot; deferring NfsInstance create/wait",
		"mode", rc.Mode,
	)
	// Long requeue; util.Timing tops out at 5m, so a raw 1h is used deliberately here (recovery
	// is an operator-driven window, not a tight loop).
	return composed.StopWithRequeueDelay(time.Hour), ctx
}
