package iprange

import (
	"context"

	"github.com/kyma-project/cloud-manager/components/kcp/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-manager/components/lib/composed"
)

func compareStates(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	//Check and see whether the desiredState == actualState
	deleting := !state.Obj().GetDeletionTimestamp().IsZero()
	gcpOptions := state.ObjAsIpRange().Spec.Options.Gcp
	ipRange := state.ObjAsIpRange()

	state.addressOp = focal.NONE
	state.connectionOp = focal.NONE
	if deleting {
		//If the address exists, delete it.
		if state.address != nil {
			state.addressOp = focal.DELETE
		}
		//If service connection exists,
		//delete or update it based on whether IP range is present in it.
		index := state.doesConnectionIncludeRange()
		if state.serviceConnection != nil && index >= 0 {
			if len(state.serviceConnection.ReservedPeeringRanges) > 1 {
				state.connectionOp = focal.MODIFY
				state.ipRanges = append(state.serviceConnection.ReservedPeeringRanges[:index],
					state.serviceConnection.ReservedPeeringRanges[index+1:]...)
			} else {
				state.connectionOp = focal.DELETE
			}
		}
	} else {
		if state.address == nil {
			//If address doesn't exist, add it.
			state.addressOp = focal.ADD
		} else if !state.doesAddressMatch() {
			//If the address exists, but does not match, update it.
			state.addressOp = focal.MODIFY
		}

		//Check whether the IPRange is created for PSA.
		//If yes, identify the operation to be performed.
		//If no, serviceConnection object is not needed, so ignore.
		if gcpOptions == nil || gcpOptions.Purpose == v1beta1.GcpPurposePSA {
			if state.serviceConnection == nil {
				//If serviceConnection doesn't exist, add it.
				state.connectionOp = focal.ADD
				state.ipRanges = []string{ipRange.Name}
			} else if index := state.doesConnectionIncludeRange(); index < 0 {
				//If connection exists, but the ipRange is not part of it, include it.
				state.connectionOp = focal.MODIFY
				state.ipRanges = append(state.serviceConnection.ReservedPeeringRanges, ipRange.Name)
			}
		}
	}
	state.inSync = state.addressOp == focal.NONE && state.connectionOp == focal.NONE

	return nil, nil
}
