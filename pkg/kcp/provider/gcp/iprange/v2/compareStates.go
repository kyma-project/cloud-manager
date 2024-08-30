package v2

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"

	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func compareStates(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	//Check and see whether the desiredState == actualState
	deleting := !state.Obj().GetDeletionTimestamp().IsZero()
	gcpOptions := state.ObjAsIpRange().Spec.Options.Gcp

	state.addressOp = client.NONE
	state.connectionOp = client.NONE
	if deleting {
		//If the address exists, delete it.
		if state.address != nil {
			state.addressOp = client.DELETE
		}
		//If service connection exists,
		//delete or update it based on whether IP range is present in it.
		index := state.doesConnectionIncludeRange()
		if state.serviceConnection != nil && index >= 0 {
			if len(state.serviceConnection.ReservedPeeringRanges) > 1 {
				state.connectionOp = client.MODIFY
			} else {
				state.connectionOp = client.DELETE
			}
		}

		//Set the State value.
		if state.connectionOp != client.NONE {
			state.curState = client.DeletePsaConnection
		} else if state.addressOp != client.NONE {
			state.curState = client.DeleteAddress
		} else {
			state.curState = client.Deleted
		}
	} else {
		if state.address == nil {
			//If address doesn't exist, add it.
			state.addressOp = client.ADD
		} else if !state.doesAddressMatch() {
			//If the address exists, but does not match, update it.
			state.addressOp = client.MODIFY
		}

		//Check whether the IPRange is created for PSA.
		//If yes, identify the operation to be performed.
		//If no, serviceConnection object is not needed, so ignore.
		if state.address == nil {
			state.connectionOp = client.NONE
		} else if gcpOptions == nil || gcpOptions.Purpose == v1beta1.GcpPurposePSA {
			if state.serviceConnection == nil {
				//If serviceConnection doesn't exist, add it.
				state.connectionOp = client.ADD
			} else if index := state.doesConnectionIncludeRange(); index < 0 {
				//If connection exists, but the ipRange is not part of it, include it.
				state.connectionOp = client.MODIFY
			}
		}

		//Set the State value.
		if state.addressOp != client.NONE {
			state.curState = client.SyncAddress
		} else if state.connectionOp != client.NONE {
			state.curState = client.SyncPsaConnection
		} else {
			state.curState = v1beta1.ReadyState
		}
	}
	state.inSync = state.addressOp == client.NONE && state.connectionOp == client.NONE

	return nil, nil
}
