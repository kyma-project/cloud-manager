package subnet

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/pkg/common/leases"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func acquireLease(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	subnet := state.ObjAsGcpSubnet()

	leaseName := GetLeaseName(state.Scope().Name)
	leaseNamespace := subnet.Namespace
	holderName := fmt.Sprintf("%s/%s", subnet.Namespace, subnet.Name)
	leaseDurationSec := int32(5 * 60) // 5min

	res, err := leases.Acquire(
		ctx,
		state.Cluster(),
		leaseName,
		leaseNamespace,
		holderName,
		leaseDurationSec,
	)

	switch res {
	case leases.AcquiredLease, leases.RenewedLease:
		return nil, nil
	case leases.LeasingFailed:
		return composed.LogErrorAndReturn(err, "Error acquiring lease", composed.StopWithRequeueDelay(util.Timing.T100ms()), ctx)
	case leases.OtherLeased:
		logger.Info("Another subnet leased the connection policy. Waiting for it to release the lease.")
		return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
	default:
		return composed.LogErrorAndReturn(err, "Unknown lease result", composed.StopAndForget, ctx)
	}
}
