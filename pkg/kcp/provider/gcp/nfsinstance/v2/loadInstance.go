package v2

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/config"
	v2client "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsinstance/v2/client"
)

// loadInstance loads the GCP Filestore instance from the API.
func loadInstance(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	nfsInstance := state.ObjAsNfsInstance()
	logger.Info("Loading GCP Filestore Instance")

	// Get GCP details
	project := state.GetGcpProjectId()
	location := state.GetGcpLocation()
	name := v2client.GetFilestoreInstanceId(nfsInstance.Name)

	// Call Filestore API to get instance
	instance, err := state.GetFilestoreClient().GetInstance(ctx, project, location, name)
	if err != nil {
		// Handle "not found" - instance doesn't exist yet
		if status.Code(err) == codes.NotFound {
			state.SetInstance(nil)
			return nil, ctx
		}

		// API error - update status and requeue
		logger.Error(err, "Error getting Filestore Instance from GCP")
		return composed.UpdateStatus(nfsInstance).
			SetExclusiveConditions(metav1.Condition{
				Type:    v1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  v1beta1.ReasonGcpError,
				Message: "Error getting Filestore Instance from GCP",
			}).
			SuccessError(composed.StopWithRequeueDelay(config.GcpConfig.GcpRetryWaitTime)).
			Run(ctx, state)
	}

	// Cache the instance in state
	state.SetInstance(instance)

	return nil, ctx
}
