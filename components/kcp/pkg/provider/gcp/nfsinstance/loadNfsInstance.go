package nfsinstance

import (
	"context"
	"errors"

	"github.com/kyma-project/cloud-manager/components/kcp/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/provider/gcp/client"
	"github.com/kyma-project/cloud-manager/components/lib/composed"
	"google.golang.org/api/googleapi"
)

func loadNfsInstance(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	nfsInstance := state.ObjAsNfsInstance()
	logger.WithValues("NfsInstance :", nfsInstance.Name).Info("Loading GCP Filestore Instance")

	//Get GCP details.
	gcpScope := state.Scope().Spec.Scope.Gcp
	project := gcpScope.Project
	location := state.getGcpLocation()
	name := nfsInstance.Spec.RemoteRef.Name

	inst, err := state.filestoreClient.GetFilestoreInstance(ctx, project, location, name)
	if err != nil {

		var e *googleapi.Error
		if ok := errors.As(err, &e); ok {
			if e.Code == 404 {
				state.fsInstance = nil
				return nil, nil
			}
		}
		state.AddErrorCondition(ctx, v1beta1.ReasonGcpError, err)
		return composed.LogErrorAndReturn(err, "Error getting Filestore Instance from GCP", composed.StopWithRequeueDelay(client.GcpRetryWaitTime), nil)
	}

	//Store the fsInstance in state
	state.fsInstance = inst

	return nil, nil
}
