package subnet

import (
	"context"

	"github.com/google/uuid"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	gcpmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/meta"
	gcpsubnetclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/subnet/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func deleteConnectionPolicy(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.serviceConnectionPolicy == nil {
		return nil, nil
	}

	if state.subnet == nil {
		return nil, nil
	}

	gcpScope := state.Scope().Spec.Scope.Gcp
	region := state.Scope().Spec.Region

	if !state.ShouldDeleteConnectionPolicy(gcpScope.Project, region) {
		return nil, nil
	}

	logger.Info("Deleting GCP Connection Policy")

	err := state.networkComnnectivityClient.DeleteServiceConnectionPolicy(ctx, gcpsubnetclient.DeleteServiceConnectionPolicyRequest{
		Name:          state.serviceConnectionPolicy.Name,
		IdempotenceId: uuid.NewString(),
	})
	if err != nil {
		if gcpmeta.IsNotFound(err) {
			logger.Info("target GCP Connection Policy for delete not found, continuing to next loop")
			return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
		}

		logger.Error(err, "Error deleting GCP Connection Policy")
		subnet := state.ObjAsGcpSubnet()
		meta.SetStatusCondition(subnet.Conditions(), metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  cloudcontrolv1beta1.ReasonCloudProviderError,
			Message: "Failed to delete Subnet",
		})
		subnet.Status.State = cloudcontrolv1beta1.StateError

		err = state.UpdateObjStatus(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err,
				"Error updating Subnet status due failed GCP Connection Policy deleting",
				composed.StopWithRequeueDelay((util.Timing.T10000ms())),
				ctx,
			)
		}

		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}

	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
}
