package nfsinstance

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
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
	prevState := nfsInstance.Status.State
	nfsInstance.Status.State = state.curState

	if state.curState == v1beta1.ReadyState {
		nfsInstance.Status.Hosts = state.fsInstance.Networks[0].IpAddresses
		nfsInstance.Status.CapacityGb = int(state.fsInstance.FileShares[0].CapacityGb)
		return composed.UpdateStatus(nfsInstance).
			SetExclusiveConditions(metav1.Condition{
				Type:    v1beta1.ConditionTypeReady,
				Status:  metav1.ConditionTrue,
				Reason:  v1beta1.ReasonReady,
				Message: "Filestore instance provisioned in GCP.",
			}).
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	} else if prevState != state.curState {
		return composed.UpdateStatus(nfsInstance).SuccessError(composed.StopWithRequeue).Run(ctx, state)
	}

	return nil, nil
}
