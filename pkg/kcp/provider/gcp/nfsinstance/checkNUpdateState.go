package nfsinstance

import (
	"context"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/kcp/provider/gcp/client"

	"github.com/kyma-project/cloud-manager/components/kcp/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/composed"
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

	state.operation = client.NONE
	if deleting {
		if state.fsInstance == nil {
			state.curState = client.Deleted
		} else if state.fsInstance.State != string(client.DELETING) {
			//If the filestore exists and not DELETING, delete it.
			state.operation = client.DELETE
			state.curState = client.SyncFilestore
		}
	} else {
		if state.fsInstance == nil {
			//If filestore doesn't exist, add it.
			state.operation = client.ADD
			state.curState = client.SyncFilestore
		} else if !state.doesFilestoreMatch() {
			//If the filestore exists, but does not match, update it.
			state.operation = client.MODIFY
			state.curState = client.SyncFilestore
		} else if state.fsInstance.State == string(client.READY) {
			state.curState = v1beta1.ReadyState
		}
	}

	//Update State Info
	state.ObjAsNfsInstance().Status.State = state.curState

	if state.curState == v1beta1.ReadyState {
		state.ObjAsNfsInstance().Status.Hosts = state.fsInstance.Networks[0].IpAddresses
		meta.RemoveStatusCondition(state.ObjAsNfsInstance().Conditions(), v1beta1.ConditionTypeError)
		state.AddReadyCondition(ctx, "Filestore Instance provisioned in GCP.")
	}

	err := state.UpdateObjStatus(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating NfsInstance success status", composed.StopWithRequeue, nil)
	}

	if state.curState == v1beta1.ReadyState {
		return composed.StopAndForget, nil
	}

	return nil, nil
}
