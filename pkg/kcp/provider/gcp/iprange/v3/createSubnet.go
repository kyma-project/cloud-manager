package v3

import (
	"context"

	"github.com/google/uuid"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	gcpiprangev3client "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange/v3/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createSubnet(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.subnet != nil {
		return nil, nil
	}

	ipRange := state.ObjAsIpRange()
	gcpScope := state.Scope().Spec.Scope.Gcp
	region := state.Scope().Spec.Region

	err := state.computeClient.CreatePrivateSubnet(ctx, gcpiprangev3client.CreateSubnetRequest{
		ProjectId:     gcpScope.Project,
		Region:        region,
		Network:       gcpScope.VpcNetwork,
		Name:          GetSubnetShortName(state.Obj().GetName()),
		Cidr:          ipRange.Spec.Cidr,
		IdempotenceId: uuid.NewString(),
	})

	if err != nil {
		logger.Error(err, "Error creating GCP Private Subnet")
		meta.SetStatusCondition(ipRange.Conditions(), metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  cloudcontrolv1beta1.ReasonCloudProviderError,
			Message: "Failed to create GCP Private Subnet",
		})
		ipRange.Status.State = cloudcontrolv1beta1.StateError

		err = state.UpdateObjStatus(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err,
				"Error updating IpRange status due failed GCP Private Subnet creation",
				composed.StopWithRequeueDelay((util.Timing.T10000ms())),
				ctx,
			)
		}

		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}

	return composed.StopWithRequeue, nil
}
