package nfsinstance

import (
	"context"
	"fmt"
	"strings"

	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"google.golang.org/api/file/v1"
)

func syncNfsInstance(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	nfsInstance := state.ObjAsNfsInstance()
	logger.WithValues("NfsInstance", nfsInstance.Name).Info("Saving GCP Filestore")

	//Get GCP details.
	gcpScope := state.Scope().Spec.Scope.Gcp
	project := gcpScope.Project
	location := state.getGcpLocation()
	name := fmt.Sprintf("cm-%.60s", nfsInstance.Name)

	var operation *file.Operation
	var err error

	switch state.operation {
	case gcpclient.ADD:
		operation, err = state.filestoreClient.CreateFilestoreInstance(ctx, project, location, name, state.toInstance())
	case gcpclient.MODIFY:
		mask := strings.Join(state.updateMask, ",")
		operation, err = state.filestoreClient.PatchFilestoreInstance(ctx, project, location, name, mask, state.toInstance())
	case gcpclient.DELETE:
		operation, err = state.filestoreClient.DeleteFilestoreInstance(ctx, project, location, name)
	case gcpclient.NONE:
		//If the operation is not ADD, MODIFY or DELETE, continue.
		return nil, nil
	}

	if err != nil {
		return composed.UpdateStatus(nfsInstance).
			SetExclusiveConditions(metav1.Condition{
				Type:    v1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  v1beta1.ReasonGcpError,
				Message: err.Error(),
			}).
			SuccessError(composed.StopWithRequeueDelay(gcpclient.GcpConfig.GcpRetryWaitTime)).
			SuccessLogMsg(fmt.Sprintf("Error updating Filestore object in GCP :%s", err)).
			Run(ctx, state)
	}

	if operation != nil {
		if state.operation == gcpclient.ADD {
			nfsInstance.Status.Id = gcpclient.GetFilestoreInstancePath(project, location, name)
		}
		nfsInstance.Status.OpIdentifier = operation.Name
		return composed.UpdateStatus(nfsInstance).
			SuccessError(composed.StopWithRequeueDelay(gcpclient.GcpConfig.GcpOperationWaitTime)).
			Run(ctx, state)
	}
	return nil, nil
}
