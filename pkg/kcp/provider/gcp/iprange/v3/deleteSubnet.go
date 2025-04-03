package v3

import (
	"context"

	"github.com/google/uuid"
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	v3 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange/v3/client"
	gcpmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/meta"
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

	err := state.computeClient.DeleteSubnet(ctx, v3.DeleteSubnetRequest{
		ProjectId:     gcpScope.Project,
		Region:        region,
		Name:          GetPrivateSubnetShortName(state.Obj().GetName()),
		IdempotenceId: uuid.NewString(),
	})
	if err != nil {
		if gcpmeta.IsNotFound(err) {
			logger.Info("target GCP Private Subnet for delete not found, continuing to next loop")
			return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
		}

		logger.Error(err, "Error deleting GCP Private Subnet")
		ipRange := state.ObjAsIpRange()
		meta.SetStatusCondition(ipRange.Conditions(), metav1.Condition{
			Type:    v1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  v1beta1.ReasonCloudProviderError,
			Message: "Failed to delete IpRange",
		})
		ipRange.Status.State = cloudcontrolv1beta1.StateError

		err = state.UpdateObjStatus(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err,
				"Error updating IpRange status due failed GCP Private Subnet deleting",
				composed.StopWithRequeueDelay((util.Timing.T10000ms())),
				ctx,
			)
		}

		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}

	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
}
