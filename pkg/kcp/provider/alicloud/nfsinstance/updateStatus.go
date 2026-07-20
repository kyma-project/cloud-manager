package nfsinstance

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// updateStatus sets the NfsInstance Host/Path/Id and the Ready condition once the NAS file
// system and its mount target are available.
func updateStatus(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if composed.MarkedForDeletionPredicate(ctx, state) {
		return nil, ctx
	}

	if len(state.mountTargets) == 0 {
		// nothing to expose yet; requeue via the wait actions
		return composed.StopWithRequeue, ctx
	}

	host := state.mountTargets[0].MountTargetDomain

	if state.ObjAsNfsInstance().Status.Id == state.fileSystemId &&
		state.ObjAsNfsInstance().Status.Host == host &&
		meta.IsStatusConditionTrue(*state.ObjAsNfsInstance().Conditions(), cloudcontrolv1beta1.ConditionTypeReady) {
		// already set and saved
		return nil, ctx
	}

	state.ObjAsNfsInstance().Status.Id = state.fileSystemId
	state.ObjAsNfsInstance().Status.Host = host
	state.ObjAsNfsInstance().Status.Path = "/"

	return composed.UpdateStatus(state.ObjAsNfsInstance()).
		SetExclusiveConditions(metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeReady,
			Status:  metav1.ConditionTrue,
			Reason:  cloudcontrolv1beta1.ReasonReady,
			Message: "NFS instance is ready",
		}).
		ErrorLogMessage("Error updating AliCloud KCP NfsInstance status after setting Ready condition").
		SuccessLogMsg("AliCloud KCP NfsInstance is ready").
		SuccessError(composed.StopAndForget).
		Run(ctx, state)
}
