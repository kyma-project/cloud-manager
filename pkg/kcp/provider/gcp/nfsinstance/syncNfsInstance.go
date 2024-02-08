package nfsinstance

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"

	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"google.golang.org/api/file/v1"
)

func syncNfsInstance(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	nfsInstance := state.ObjAsNfsInstance()
	logger.WithValues("NfsInstance :", nfsInstance.Name).Info("Saving GCP Filestore")

	//Get GCP details.
	gcpScope := state.Scope().Spec.Scope.Gcp
	project := gcpScope.Project
	location := state.getGcpLocation()
	name := nfsInstance.Spec.RemoteRef.Name

	var operation *file.Operation
	var err error

	switch state.operation {
	case client.ADD:
		operation, err = state.filestoreClient.CreateFilestoreInstance(ctx, project, location, name, state.toInstance())
	case client.MODIFY:
		mask := strings.Join(state.updateMask, ",")
		operation, err = state.filestoreClient.PatchFilestoreInstance(ctx, project, location, name, mask, state.toInstance())
	case client.DELETE:
		operation, err = state.filestoreClient.DeleteFilestoreInstance(ctx, project, location, name)
	}

	if err != nil {
		return composed.UpdateStatus(nfsInstance).
			SetExclusiveConditions(metav1.Condition{
				Type:    v1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  v1beta1.ReasonGcpError,
				Message: err.Error(),
			}).
			SuccessError(composed.StopWithRequeueDelay(client.GcpRetryWaitTime)).
			SuccessLogMsg(fmt.Sprintf("Error updating Filestore object in GCP :%s", err)).
			Run(ctx, state)
	}

	if operation != nil {
		if state.operation == client.ADD {
			nfsInstance.Status.Id = client.GetFilestoreInstancePath(project, location, name)
		}
		nfsInstance.Status.OpIdentifier = operation.Name
		return composed.UpdateStatus(nfsInstance).
			SuccessError(composed.StopWithRequeueDelay(client.GcpOperationWaitTime)).
			Run(ctx, state)
	}
	return nil, nil
}
