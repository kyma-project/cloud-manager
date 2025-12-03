package nfsinstance

import (
	"context"
	"errors"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"google.golang.org/api/googleapi"
)

func loadNfsInstance(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	nfsInstance := state.ObjAsNfsInstance()
	logger.WithValues("NfsInstance", nfsInstance.Name).Info("Loading GCP Filestore Instance")

	//Get GCP details.
	gcpScope := state.Scope().Spec.Scope.Gcp
	project := gcpScope.Project
	location := state.getGcpLocation()
	name := fmt.Sprintf("cm-%.60s", nfsInstance.Name)

	inst, err := state.filestoreClient.GetFilestoreInstance(ctx, project, location, name)
	if err != nil {

		var e *googleapi.Error
		if ok := errors.As(err, &e); ok {
			// Handle 404 (not found) and 403 (invalid location or unauthorized) as resource doesn't exist
			if e.Code == 404 || e.Code == 403 {
				state.fsInstance = nil
				return nil, nil
			}
		}
		logger.Error(err, "Error getting Filestore Instance from GCP.")
		return composed.UpdateStatus(nfsInstance).
			SetExclusiveConditions(metav1.Condition{
				Type:    v1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  v1beta1.ReasonGcpError,
				Message: "Error getting Filestore Instance from GCP",
			}).
			SuccessError(composed.StopWithRequeueDelay(gcpclient.GcpConfig.GcpRetryWaitTime)).
			Run(ctx, state)
	}

	//Store the fsInstance in state
	state.fsInstance = inst

	return nil, nil
}
