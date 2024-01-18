package iprange

import (
	"context"
	"errors"

	"github.com/kyma-project/cloud-manager/components/kcp/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-manager/components/lib/composed"
)

func syncAddress(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	ipRange := state.ObjAsIpRange()
	logger.WithValues("ipRange :", ipRange.Name).Info("Saving GCP Address")

	gcpScope := state.Scope().Spec.Scope.Gcp
	project := gcpScope.Project
	vpc := gcpScope.VpcNetwork

	switch state.addressOp {
	case focal.ADD:
		_, err := state.computeClient.CreatePscIpRange(ctx, project, vpc, ipRange.Name, ipRange.Name, state.ipAddress, int64(state.prefix))
		if err != nil {
			state.AddErrorCondition(ctx, v1beta1.ReasonGcpError, err)
			return composed.LogErrorAndReturn(err, "Error creating Address object in GCP", composed.StopWithRequeue, nil)
		}
	case focal.MODIFY:
		err := errors.New("IpRange update not supported.")
		state.AddErrorCondition(ctx, v1beta1.ReasonNotSupported, err)
		return composed.LogErrorAndReturn(err, "IpRange update not supported.", composed.StopAndForget, nil)
	case focal.DELETE:
		_, err := state.computeClient.DeleteIpRange(ctx, project, ipRange.Name)
		if err != nil {
			state.AddErrorCondition(ctx, v1beta1.ReasonGcpError, err)
			return composed.LogErrorAndReturn(err, "Error deleting address object in GCP", composed.StopWithRequeue, nil)
		}
	}

	return composed.StopWithRequeue, nil
}
