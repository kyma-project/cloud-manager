package nfsinstance

import (
	"context"

	"github.com/kyma-project/cloud-manager/components/kcp/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/provider/gcp/client"
	"github.com/kyma-project/cloud-manager/components/lib/composed"
	"k8s.io/apimachinery/pkg/api/meta"
)

func checkNUpdateState(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	nfsInstance := state.ObjAsNfsInstance()
	logger.WithValues("NfsInstance :", nfsInstance.Name).Info("Updating State Info")

	//Compute State Info
	//Check and see whether the desiredState == actualState
	deleting := !state.Obj().GetDeletionTimestamp().IsZero()

	state.operation = focal.NONE
	if deleting {
		//If the address exists, delete it.
		if state.fsInstance != nil {
			state.operation = focal.DELETE
			state.curState = client.SyncFilestore
		} else {
			state.curState = client.Deleted
		}
	} else {
		if state.fsInstance == nil {
			//If filestore doesn't exist, add it.
			state.operation = focal.ADD
			state.curState = client.SyncFilestore
		} else if !state.doesFilestoreMatch() {
			//If the address exists, but does not match, update it.
			state.operation = focal.MODIFY
			state.curState = client.SyncFilestore
		} else {
			state.curState = v1beta1.ReadyState
		}
	}

	//Update State Info
	state.ObjAsNfsInstance().Status.State = state.curState

	if state.curState == v1beta1.ReadyState {
		meta.RemoveStatusCondition(state.ObjAsNfsInstance().Conditions(), v1beta1.ConditionTypeError)
		state.AddReadyCondition(ctx, "Filestore Instance provisioned in GCP.")
	}

	err := state.UpdateObjStatus(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating IpRange success status", composed.StopWithRequeue, nil)
	}

	if state.curState == v1beta1.ReadyState {
		return composed.StopAndForget, nil
	}

	return nil, nil
}
