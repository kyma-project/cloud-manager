package v1

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/config"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/elliotchance/pie/v2"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func checkNUpdateState(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	nfsInstance := state.ObjAsNfsInstance()
	logger.WithValues("NfsInstance", nfsInstance.Name).Info("Updating State Info")

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
			state.curState = client.Deleting
		}
	} else {
		if state.fsInstance == nil {
			//If filestore doesn't exist, add it.
			state.operation = client.ADD
			state.curState = client.Creating
		} else if state.fsInstance.State == string(client.ERROR) {
			state.curState = v1beta1.StateError
		} else if !state.doesFilestoreMatch() {
			//If the filestore exists, but does not match, update it.
			state.operation = client.MODIFY
			state.curState = client.Updating
		} else if state.fsInstance.State == string(client.READY) {
			state.curState = v1beta1.StateReady
		} else {
			//If the filestore exists but is not READY or in ERROR, it is in a transient state.
			return composed.StopWithRequeueDelay(config.GcpConfig.GcpRetryWaitTime), ctx
		}
	}

	//Update State Info
	prevState := nfsInstance.Status.State
	nfsInstance.Status.State = state.curState
	logger.Info("State Info", "curState", state.curState, "Operation", state.operation)
	if state.curState == v1beta1.StateReady {
		nfsInstance.Status.Hosts = state.fsInstance.Networks[0].IpAddresses
		nfsInstance.Status.Host = pie.First(state.fsInstance.Networks[0].IpAddresses)
		nfsInstance.Status.Path = state.ObjAsNfsInstance().Spec.Instance.Gcp.FileShareName
		nfsInstance.Status.CapacityGb = int(state.fsInstance.FileShares[0].CapacityGb)
		if qty, err := resource.ParseQuantity(fmt.Sprintf("%dGi", state.fsInstance.FileShares[0].CapacityGb)); err == nil {
			nfsInstance.Status.Capacity = qty
		} else {
			logger.Error(err, "Error parsing capacity quantity")
		}
		nfsInstance.SetStateData(client.GcpNfsStateDataProtocol, state.fsInstance.Protocol)
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
		if state.curState == v1beta1.StateError {
			return composed.UpdateStatus(nfsInstance).
				SetExclusiveConditions(metav1.Condition{
					Type:    v1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  v1beta1.ReasonGcpError,
					Message: state.fsInstance.StatusMessage,
				}).
				SuccessError(composed.StopWithRequeueDelay(config.GcpConfig.GcpRetryWaitTime)).
				Run(ctx, state)
		}
		return composed.UpdateStatus(nfsInstance).
			RemoveConditions(v1beta1.ConditionTypeReady).
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	return nil, nil
}
