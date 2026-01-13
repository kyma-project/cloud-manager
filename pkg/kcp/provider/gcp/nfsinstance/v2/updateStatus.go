package v2

import (
	"context"
	"fmt"

	"cloud.google.com/go/filestore/apiv1/filestorepb"
	"github.com/elliotchance/pie/v2"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/config"
)

// updateStatus updates the NfsInstance status based on the current state.
func updateStatus(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	nfsInstance := state.ObjAsNfsInstance()
	instance := state.GetInstance()

	// Determine desired state from instance
	var desiredState v1beta1.StatusState
	if instance == nil {
		desiredState = v1beta1.StatusState("Creating")
	} else {
		switch instance.State {
		case filestorepb.Instance_READY:
			desiredState = v1beta1.StateReady
		case filestorepb.Instance_CREATING:
			desiredState = v1beta1.StatusState("Creating")
		case filestorepb.Instance_DELETING:
			desiredState = v1beta1.StatusState("Deleting")
		case filestorepb.Instance_ERROR:
			desiredState = v1beta1.StateError
		default:
			desiredState = v1beta1.StatusState(instance.State.String())
		}
	}

	currentState := nfsInstance.Status.State
	changed := false

	// Update state if different
	if currentState != desiredState {
		nfsInstance.Status.State = desiredState
		changed = true
		logger.Info("State changed", "previousState", currentState, "newState", desiredState)
	}

	// READY state: update fields
	if desiredState == v1beta1.StateReady && instance != nil {
		// Set hosts and path
		if len(instance.Networks) > 0 && len(instance.Networks[0].IpAddresses) > 0 {
			newHosts := instance.Networks[0].IpAddresses
			if !pie.Equals(nfsInstance.Status.Hosts, newHosts) {
				nfsInstance.Status.Hosts = newHosts
				changed = true
			}
			newHost := pie.First(instance.Networks[0].IpAddresses)
			if nfsInstance.Status.Host != newHost {
				nfsInstance.Status.Host = newHost
				changed = true
			}
		}
		newPath := nfsInstance.Spec.Instance.Gcp.FileShareName
		if nfsInstance.Status.Path != newPath {
			nfsInstance.Status.Path = newPath
			changed = true
		}

		// Set capacity
		if len(instance.FileShares) > 0 {
			if nfsInstance.Status.CapacityGb != int(instance.FileShares[0].CapacityGb) {
				nfsInstance.Status.CapacityGb = int(instance.FileShares[0].CapacityGb)
				changed = true
			}
			if qty, err := resource.ParseQuantity(fmt.Sprintf("%dGi", instance.FileShares[0].CapacityGb)); err == nil {
				if nfsInstance.Status.Capacity.Cmp(qty) != 0 {
					nfsInstance.Status.Capacity = qty
					changed = true
				}
			} else {
				logger.Error(err, "Error parsing capacity quantity")
			}
		}

		// Set protocol state data
		prevProtocol, _ := nfsInstance.GetStateData(gcpclient.GcpNfsStateDataProtocol)
		newProtocol := ""
		if instance.Protocol != 0 {
			newProtocol = instance.Protocol.String()
		}
		if prevProtocol != newProtocol {
			nfsInstance.SetStateData(gcpclient.GcpNfsStateDataProtocol, newProtocol)
			changed = true
		}

		if changed {
			return composed.UpdateStatus(nfsInstance).
				SetExclusiveConditions(metav1.Condition{
					Type:    v1beta1.ConditionTypeReady,
					Status:  metav1.ConditionTrue,
					Reason:  v1beta1.ReasonReady,
					Message: "Filestore instance provisioned in GCP.",
				}).
				SuccessError(composed.StopAndForget).
				Run(ctx, state)
		}
		return nil, ctx
	}

	// ERROR state: update error condition
	if desiredState == v1beta1.StateError && instance != nil {
		errorMessage := "Filestore instance in error state"
		if instance.StatusMessage != "" {
			errorMessage = instance.StatusMessage
		}
		return composed.UpdateStatus(nfsInstance).
			SetExclusiveConditions(metav1.Condition{
				Type:    v1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  v1beta1.ReasonGcpError,
				Message: errorMessage,
			}).
			SuccessError(composed.StopWithRequeueDelay(config.GcpConfig.GcpRetryWaitTime)).
			Run(ctx, state)
	}

	// If state changed but not READY or ERROR, update status
	if changed {
		return composed.UpdateStatus(nfsInstance).
			RemoveConditions(v1beta1.ConditionTypeReady).
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	return nil, ctx
}
