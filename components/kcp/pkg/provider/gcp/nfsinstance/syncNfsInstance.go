package nfsinstance

import (
	"context"
	"strings"

	"github.com/kyma-project/cloud-manager/components/kcp/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/provider/gcp/client"
	"github.com/kyma-project/cloud-manager/components/lib/composed"
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
	case focal.ADD:
		operation, err = state.filestoreClient.CreateFilestoreInstance(ctx, project, location, name, state.toInstance())
	case focal.MODIFY:
		mask := strings.Join(state.updateMask, ",")
		operation, err = state.filestoreClient.PatchFilestoreInstance(ctx, project, location, name, mask, state.toInstance())
	case focal.DELETE:
		operation, err = state.filestoreClient.DeleteFilestoreInstance(ctx, project, location, name)
	}

	if err != nil {
		state.AddErrorCondition(ctx, v1beta1.ReasonGcpError, err)
		return composed.LogErrorAndReturn(err, "Error synchronizing Filestore object in GCP", composed.StopWithRequeueDelay(client.GcpRetryWaitTime), nil)
	}

	if operation != nil {
		nfsInstance.Status.OpIdentifier = operation.Name
		state.UpdateObjStatus(ctx)
	}
	return nil, nil
}
