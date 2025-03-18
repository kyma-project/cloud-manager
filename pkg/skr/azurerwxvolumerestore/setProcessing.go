package azurerwxvolumerestore

import (
	"context"
	"github.com/google/uuid"
	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

// This action adds a start time to the status of the restore object prior to trigger the restore.
// This can be used as an indication that restore is going to be trigger shortly and also as filter to enhance the search for a restore job
func setProcessing(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	restore := state.ObjAsAzureRwxVolumeRestore()
	if restore.Status.RestoredDir != "" {
		return nil, nil
	}
	restore.Status.StartTime = &metav1.Time{Time: time.Now()}
	restore.Status.RestoredDir = uuid.NewString()
	restore.Status.State = v1beta1.JobStateProcessing
	return composed.PatchStatus(restore).
		SuccessErrorNil().
		Run(ctx, state)
}
