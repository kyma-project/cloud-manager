package v2

import (
	"context"
	"fmt"

	"cloud.google.com/go/filestore/apiv1/filestorepb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/config"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

// createInstance creates a new Filestore instance in GCP.
func createInstance(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.GetInstance() != nil {
		return nil, ctx
	}

	nfsInstance := state.ObjAsNfsInstance()
	logger.Info("Creating GCP Filestore Instance")

	project := state.GetGcpProjectId()
	location := state.GetGcpLocation()
	name := nfsInstance.Name

	instance := state.ToGcpInstance()

	gcpOptions := nfsInstance.Spec.Instance.Gcp
	if feature.Nfs41Gcp.Value(ctx) &&
		(gcpOptions.Tier == v1beta1.ZONAL || gcpOptions.Tier == v1beta1.REGIONAL) {
		instance.Protocol = filestorepb.Instance_NFS_V4_1
	}

	operationName, err := state.GetFilestoreClient().CreateInstance(ctx, project, location, name, instance)
	if err != nil {
		logger.Error(err, "Error creating Filestore Instance in GCP")
		return composed.UpdateStatus(nfsInstance).
			SetExclusiveConditions(metav1.Condition{
				Type:    v1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  v1beta1.ReasonGcpError,
				Message: fmt.Sprintf("Error creating Filestore Instance: %s", err.Error()),
			}).
			SuccessError(composed.StopWithRequeueDelay(config.GcpConfig.GcpRetryWaitTime)).
			SuccessLogMsg("Error creating Filestore Instance in GCP").
			Run(ctx, state)
	}

	changed := false

	newId := gcpclient.GetFilestoreInstancePath(project, location, name)
	if nfsInstance.Status.Id != newId {
		nfsInstance.Status.Id = newId
		changed = true
	}

	if operationName != "" && nfsInstance.Status.OpIdentifier != operationName {
		nfsInstance.Status.OpIdentifier = operationName
		changed = true
	}

	newState := v1beta1.StatusState("Creating")
	if nfsInstance.Status.State != newState {
		nfsInstance.Status.State = newState
		changed = true
	}

	if changed {
		return composed.UpdateStatus(nfsInstance).
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T60000ms())).
			Run(ctx, state)
	}

	return nil, ctx
}
