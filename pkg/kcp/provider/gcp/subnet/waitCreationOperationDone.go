package subnet

import (
	"context"

	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/ptr"
)

func waitCreationOperationDone(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.subnetCreationOperation == nil {
		return nil, ctx
	}

	actual := ptr.Deref(state.subnetCreationOperation.Status, computepb.Operation_PENDING).String()
	done := computepb.Operation_DONE.String()

	if actual == done {
		return nil, ctx
	}

	logger.Info("Waiting KCP GcpSubnet creation operation to be done")

	return composed.StopWithRequeueDelay(5 * util.Timing.T1000ms()), ctx
}
