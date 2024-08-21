package backupschedule

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type backupImpl interface {
	emptySourceObject() composed.ObjWithConditions
	emptyBackupList() client.ObjectList
	toObjectSlice(list client.ObjectList) []client.Object
	getBackupObject(state *State, objectMeta *metav1.ObjectMeta) (client.Object, error)
}

func loadBackupImpl(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	schedule := state.ObjAsBackupSchedule()
	logger := composed.LoggerFromCtx(ctx)

	logger.WithValues("BackupSchedule :", schedule.GetName()).Info("Load Provider Specific Implementation")

	if _, ok := schedule.(*cloudresourcesv1beta1.GcpNfsBackupSchedule); ok {
		state.backupImpl = &backupImplGcpNfs{}
	} else {
		return composed.UpdateStatus(schedule).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ReasonUnknownSchedule,
				Message: "Error identifying Schedule Provider",
			}).
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	return nil, nil
}
