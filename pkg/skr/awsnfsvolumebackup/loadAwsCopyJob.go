package awsnfsvolumebackup

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/backup/types"
	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
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

	//If the Remote RestorePoint already exists, continue
	if len(backup.Status.RemoteId) > 0 {
		return nil, ctx
	}

	//If the CopyJob is nil or not able to get the recoveryPoint ARN, continue
	jobState := copyJob.CopyJob.State
	if jobState == types.CopyJobStateFailed || copyJob.CopyJob.State == types.CopyJobStatePartial {
		msg := *copyJob.CopyJob.StatusMessage
		backup.Status.State = v1beta1.StateError
		return composed.PatchStatus(backup).
			SetExclusiveConditions(metav1.Condition{
				Type:    v1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  v1beta1.ReasonBackupFailed,
				Message: msg,
			}).
			SuccessLogMsg(fmt.Sprint("AwsNfsVolumeBackup copy Job Failed:", msg)).
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	} else if copyJob.CopyJob.DestinationRecoveryPointArn == nil {
		logger.Info(fmt.Sprint("Waiting for the CopyJob to get RestorePoint details: ", string(jobState)))
		return composed.LogErrorAndReturn(err, "AWS Copy Job does not have RestorePoint ARN, will wait", composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
	}

	//Update the remoteId with the value from CopyJob.
	logger.Info("Updating the Status with remote RestorePoint details")
	remoteId := awsutil.ParseArnResourceId(ptr.Deref(copyJob.CopyJob.DestinationRecoveryPointArn, ""))
	backup.Status.RemoteId = remoteId
	backup.Status.Locations = append(backup.Status.Locations, backup.Spec.Location)
	return composed.PatchStatus(backup).
		SetExclusiveConditions().
		SuccessError(composed.StopWithRequeueDelay(util.Timing.T1000ms())).
		Run(ctx, state)

}
