package nfsinstance

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func subnetLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	subnetId, _ := state.ObjAsNfsInstance().GetStateData(StateDataSubnetId)

	if subnetId == "" && len(state.network.Subnets) > 1 {
		arr, err := state.sapClient.ListSubnets(ctx, state.network.ID)
		if err != nil {
			return composed.LogErrorAndReturn(err, "Error listing SAP subnets", composed.StopWithRequeue, ctx)
		}
		for _, subnet := range arr {
			if subnet.CIDR == state.Scope().Spec.Scope.OpenStack.Network.Nodes {
				state.subnet = &subnet
				break
			}
		}
	} else {
		id := subnetId
		if id == "" {
			id = state.network.Subnets[0]
		}
		subnet, err := state.sapClient.GetSubnet(ctx, id)
		if err != nil {
			return composed.LogErrorAndReturn(err, "Error getting SAP subnet", composed.StopWithRequeue, ctx)
		}
		state.subnet = subnet
	}

	if state.subnet != nil {
		logger = logger.WithValues("sapSubnetId", state.subnet.ID)
		ctx = composed.LoggerIntoCtx(ctx, logger)
		logger.Info("SAP subnet loaded")
	}

	if state.subnet != nil && subnetId == "" {
		state.ObjAsNfsInstance().SetStateData(StateDataSubnetId, state.subnet.ID)

		return composed.PatchStatus(state.ObjAsNfsInstance()).
			ErrorLogMessage("Error updating SAP NfsInstance state data with subnetId").
			FailedError(composed.StopWithRequeue).
			SuccessErrorNil().
			Run(ctx, state)
	}

	return nil, ctx
}
