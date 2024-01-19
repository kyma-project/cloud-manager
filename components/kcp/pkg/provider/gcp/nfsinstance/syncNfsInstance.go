package nfsinstance

import (
	"context"
	"errors"

	"github.com/kyma-project/cloud-manager/components/kcp/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-manager/components/lib/composed"
)

func syncNfsInstance(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	nfsInstance := state.ObjAsNfsInstance()
	logger.WithValues("NfsInstance :", nfsInstance.Name).Info("Saving GCP Filestore")

	gcpScope := state.Scope().Spec.Scope.Gcp
	project := gcpScope.Project
	region := state.Scope().Spec.Region
	name := nfsInstance.Spec.RemoteRef.Name

	switch state.operation {
	case focal.ADD:
		_, err := state.filestoreClient.CreateFilestoreInstance(ctx, project, region, name, state.toInstance())
		if err != nil {
			state.AddErrorCondition(ctx, v1beta1.ReasonGcpError, err)
			return composed.LogErrorAndReturn(err, "Error creating Filestore object in GCP", composed.StopWithRequeue, nil)
		}
	case focal.MODIFY:
		err := errors.New("Filestore update not supported.")
		state.AddErrorCondition(ctx, v1beta1.ReasonNotSupported, err)
		return composed.LogErrorAndReturn(err, "Filestore update not supported.", composed.StopAndForget, nil)
	case focal.DELETE:
		_, err := state.filestoreClient.DeleteFilestoreInstance(ctx, project, region, name)
		if err != nil {
			state.AddErrorCondition(ctx, v1beta1.ReasonGcpError, err)
			return composed.LogErrorAndReturn(err, "Error deleting Filestore object in GCP", composed.StopWithRequeue, nil)
		}
	}

	return composed.StopWithRequeue, nil
}
