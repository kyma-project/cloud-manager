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

	// Determine current state from instance
	var currentState v1beta1.StatusState
	if instance == nil {
		currentState = v1beta1.StatusState("Creating")
	} else {
		switch instance.State {
		case filestorepb.Instance_READY:
			currentState = v1beta1.StateReady
		case filestorepb.Instance_CREATING:
			currentState = v1beta1.StatusState("Creating")
		case filestorepb.Instance_DELETING:
			currentState = v1beta1.StatusState("Deleting")
		case filestorepb.Instance_ERROR:
			currentState = v1beta1.StateError
		default:
			currentState = v1beta1.StatusState(instance.State.String())
		}
	}

	previousState := nfsInstance.Status.State

	logger.Info("Updating status", "currentState", currentState, "previousState", previousState)

	// Update state
	nfsInstance.Status.State = currentState

	// Handle READY state
	if currentState == v1beta1.StateReady && instance != nil {
		// Set hosts and path
		if len(instance.Networks) > 0 && len(instance.Networks[0].IpAddresses) > 0 {
			nfsInstance.Status.Hosts = instance.Networks[0].IpAddresses
			nfsInstance.Status.Host = pie.First(instance.Networks[0].IpAddresses)
		}
		nfsInstance.Status.Path = nfsInstance.Spec.Instance.Gcp.FileShareName

		// Set capacity
		if len(instance.FileShares) > 0 {
			nfsInstance.Status.CapacityGb = int(instance.FileShares[0].CapacityGb)
			if qty, err := resource.ParseQuantity(fmt.Sprintf("%dGi", instance.FileShares[0].CapacityGb)); err == nil {
				nfsInstance.Status.Capacity = qty
			} else {
				logger.Error(err, "Error parsing capacity quantity")
			}
		}

		// Set protocol state data
		if instance.Protocol != 0 {
			nfsInstance.SetStateData(gcpclient.GcpNfsStateDataProtocol, instance.Protocol.String())
		}

		// Set ready condition and stop
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

	// Handle ERROR state
	if currentState == v1beta1.StateError && instance != nil {
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

	// Handle state changes (not ready, not error)
	if previousState != currentState {
		return composed.UpdateStatus(nfsInstance).
			RemoveConditions(v1beta1.ConditionTypeReady).
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	return nil, nil
}
