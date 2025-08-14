package nfsinstance

import (
	"context"
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func updateStatus(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if composed.MarkedForDeletionPredicate(ctx, state) {
		return nil, nil
	}

	if len(state.ObjAsNfsInstance().Status.Id) > 0 &&
		len(state.ObjAsNfsInstance().Status.Hosts) > 0 &&
		len(state.ObjAsNfsInstance().Status.Hosts[0]) > 0 &&
		meta.IsStatusConditionTrue(*state.ObjAsNfsInstance().Conditions(), cloudcontrolv1beta1.ConditionTypeReady) {
		// all already set and saved
		return nil, nil
	}

	host := fmt.Sprintf(
		"%s.efs.%s.amazonaws.com",
		*state.efs.FileSystemId,
		state.Scope().Spec.Region,
	)
	state.ObjAsNfsInstance().Status.Hosts = []string{host}
	state.ObjAsNfsInstance().Status.Host = host
	state.ObjAsNfsInstance().Status.Path = "/"

	state.ObjAsNfsInstance().Status.Id = *state.efs.FileSystemId

	return composed.UpdateStatus(state.ObjAsNfsInstance()).
		SetExclusiveConditions(metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeReady,
			Status:  metav1.ConditionTrue,
			Reason:  cloudcontrolv1beta1.ReasonReady,
			Message: "NFS instance is ready",
		}).
		ErrorLogMessage("Error updating KCP NfsInstance status after setting Ready condition").
		SuccessLogMsg("KCP NfsInstance is ready").
		SuccessErrorNil().
		Run(ctx, state)
}
