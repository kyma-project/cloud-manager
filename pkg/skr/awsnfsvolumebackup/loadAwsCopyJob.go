package awsnfsvolumebackup

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func loadAwsCopyJob(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	backup := state.ObjAsAwsNfsVolumeBackup()

	//If the object is being deleted continue...
	if composed.IsMarkedForDeletion(backup) {
		return nil, ctx
	}

	//If the state.copyJob is not empty continue
	if state.copyJob != nil {
		return nil, ctx
	}

	//If Status.CopyJobId is empty, Continue.
	if len(backup.Status.CopyJobId) == 0 {
		return nil, ctx
	}

	// Load the copyJob from AWS
	logger.Info("Loading AWS Copy Job")
	copyJob, err := state.awsClient.DescribeCopyJob(ctx, backup.Status.CopyJobId)
	if err != nil && !state.awsClient.IsNotFound(err) {
		return composed.LogErrorAndReturn(err, "Error loading AWS Copy Job", err, ctx)
	}

	//store the copyJob in the state object
	state.copyJob = copyJob
	if len(backup.Status.RemoteId) >= 0 || copyJob == nil {
		return nil, ctx
	}

	//Update the remoteId with the value from CopyJob.
	remoteId := state.awsClient.ParseRecoveryPointId(*copyJob.CopyJob.DestinationRecoveryPointArn)
	backup.Status.RemoteId = remoteId
	return composed.PatchStatus(backup).
		SetExclusiveConditions().
		SuccessError(composed.StopWithRequeueDelay(util.Timing.T1000ms())).
		Run(ctx, state)

}
