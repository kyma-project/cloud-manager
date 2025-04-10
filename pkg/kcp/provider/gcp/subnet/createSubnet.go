package subnet

import (
	"context"

	"github.com/google/uuid"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/subnet/client"
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

	subnet := state.ObjAsGcpSubnet()
	gcpScope := state.Scope().Spec.Scope.Gcp
	region := state.Scope().Spec.Region

	err := state.computeClient.CreatePrivateSubnet(ctx, client.CreateSubnetRequest{
		ProjectId:     gcpScope.Project,
		Region:        region,
		Network:       gcpScope.VpcNetwork,
		Name:          GetSubnetShortName(state.Obj().GetName()),
		Cidr:          subnet.Spec.Cidr,
		IdempotenceId: uuid.NewString(),
	})

	if err != nil {
		logger.Error(err, "Error creating GCP Private Subnet")
		meta.SetStatusCondition(subnet.Conditions(), metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  cloudcontrolv1beta1.ReasonCloudProviderError,
			Message: "Failed to create GCP Private Subnet",
		})
		subnet.Status.State = cloudcontrolv1beta1.StateError

		err = state.UpdateObjStatus(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err,
				"Error updating Subnet status due failed GCP Private Subnet creation",
				composed.StopWithRequeueDelay((util.Timing.T10000ms())),
				ctx,
			)
		}

		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}

	return composed.StopWithRequeue, nil
}
