package sapnfsvolume

import (
	"context"
	"fmt"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func preventDeleteWhenUsedBySnapshot(ctx context.Context, st composed.State) (error, context.Context) {
	return composed.PreventDeleteWhenUsed(
		&cloudresourcesv1beta1.SapNfsVolumeSnapshotList{},
		st.Name().Name,
		cloudresourcesv1beta1.SapNfsVolumeField,
		func(ctx context.Context, st composed.State, _ client.ObjectList, usedByNames []string) (error, context.Context) {
			state := st.(*State)
			sapNfsVolume := state.ObjAsSapNfsVolume()
			sapNfsVolume.Status.State = cloudresourcesv1beta1.StateError
			return composed.UpdateStatus(sapNfsVolume).
				SetExclusiveConditions(metav1.Condition{
					Type:    cloudresourcesv1beta1.ConditionTypeWarning,
					Status:  metav1.ConditionTrue,
					Reason:  cloudresourcesv1beta1.ConditionTypeDeleteWhileUsed,
					Message: fmt.Sprintf("Can not be deleted while used by: %s", usedByNames),
				}).
				ErrorLogMessage("Error updating SapNfsVolume status with Warning condition for delete while in use").
				SuccessLogMsg("Forgetting SKR SapNfsVolume marked for deleting that is in use").
				SuccessError(composed.StopWithRequeueDelay(util.Timing.T10000ms())).
				Run(ctx, state)
		},
	)(ctx, st)
}
