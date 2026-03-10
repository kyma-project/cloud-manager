package gcpnfsbackupschedule

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func setStatusToActive(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	schedule := state.ObjAsBackupSchedule()

	// If schedule is already Active, nothing to update
	if schedule.State() == cloudresourcesv1beta1.JobStateActive {
		return nil, nil
	}

	schedule.SetState(cloudresourcesv1beta1.JobStateActive)
	return composed.PatchStatus(schedule).
		SetExclusiveConditions().
		SuccessError(composed.StopWithRequeue).
		Run(ctx, state)
}
