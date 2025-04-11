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

func deleteSubnet(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.subnet == nil {
		return nil, nil
	}

	logger.Info("Deleting GCP Private Subnet")

	gcpScope := state.Scope().Spec.Scope.Gcp
	region := state.Scope().Spec.Region

	err := state.computeClient.DeleteSubnet(ctx, gcpsubnetclient.DeleteSubnetRequest{
		ProjectId:     gcpScope.Project,
		Region:        region,
		Name:          GetSubnetShortName(state.Obj().GetName()),
		IdempotenceId: uuid.NewString(),
	})
	if err != nil {
		if gcpmeta.IsNotFound(err) {
			logger.Info("target GCP Private Subnet for delete not found, continuing to next loop")
			return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
		}

		logger.Error(err, "Error deleting GCP Private Subnet")
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
				"Error updating Subnet status due failed GCP Private Subnet deleting",
				composed.StopWithRequeueDelay((util.Timing.T10000ms())),
				ctx,
			)
		}

		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}

	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
}
