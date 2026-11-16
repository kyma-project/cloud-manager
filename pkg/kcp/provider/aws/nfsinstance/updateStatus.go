package nfsinstance

import (
	"context"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func updateStatus(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if composed.MarkedForDeletionPredicate(ctx, state) {
		return nil, ctx
	}

	return composed.NewStatusPatcherComposed(state.ObjAsNfsInstance()).
		MutateStatus(func(nfsInstance *cloudcontrolv1beta1.NfsInstance) {
			host := fmt.Sprintf(
				"%s.efs.%s.amazonaws.com",
				*state.efs.FileSystemId,
				state.Scope().Spec.Region,
			)
			nfsInstance.Status.Hosts = []string{host}
			nfsInstance.Status.Host = host
			nfsInstance.Status.Path = "/"
			nfsInstance.Status.Id = *state.efs.FileSystemId
		}).
		RemoveConditions(cloudcontrolv1beta1.ConditionTypeError).
		SetConditions(metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeReady,
			Status:  metav1.ConditionTrue,
			Reason:  cloudcontrolv1beta1.ReasonReady,
			Message: "NFS instance is ready",
		}).
		OnStatusChanged(composed.Log("KCP NfsInstance is ready")).
		Run(ctx, state.Cluster().K8sClient())
}
