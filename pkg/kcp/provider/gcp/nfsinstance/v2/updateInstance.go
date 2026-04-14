package v2

import (
	"context"
	"fmt"

	"cloud.google.com/go/filestore/apiv1/filestorepb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/config"
	v2client "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsinstance/v2/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// updateInstance updates an existing Filestore instance in GCP.
// Uses the updateMask pattern from RedisInstance for efficient updates.
func updateInstance(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	nfsInstance := state.ObjAsNfsInstance()
	instance := state.GetInstance()

	if instance == nil {
		return composed.StopWithRequeue, nil
	}

	if instance.State != filestorepb.Instance_READY {
		return nil, ctx
	}

	if !state.ShouldUpdateInstance() {
		return nil, ctx
	}

	logger.Info("Updating GCP Filestore Instance")

	project := state.GetGcpProjectId()
	location := state.GetGcpLocation()
	name := nfsInstance.Name

	updateInstance := state.GetInstance()
	updateInstance.Name = v2client.GetFilestoreName(project, location, name)

	// Use the modified instance from state (updated by modify actions)
	op, err := state.GetFilestoreClient().UpdateFilestoreInstance(ctx, &filestorepb.UpdateInstanceRequest{
		UpdateMask: &fieldmaskpb.FieldMask{
			Paths: state.updateMask,
		},
		Instance: updateInstance,
	})
	if err != nil {
		logger.Error(err, "Error updating Filestore Instance in GCP")
		return composed.UpdateStatus(nfsInstance).
			SetExclusiveConditions(metav1.Condition{
				Type:    v1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  v1beta1.ReasonGcpError,
				Message: fmt.Sprintf("Error updating Filestore Instance: %s", err.Error()),
			}).
			SuccessError(composed.StopWithRequeueDelay(config.GcpConfig.GcpRetryWaitTime)).
			SuccessLogMsg("Error updating Filestore Instance in GCP").
			Run(ctx, state)
	}

	operationName := op.Name()
	if operationName != "" {
		nfsInstance.Status.OpIdentifier = operationName

		return composed.UpdateStatus(nfsInstance).
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T60000ms())).
			Run(ctx, state)
	}

	return nil, ctx
}
