package vpcpeering

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func loadRemoteVpcPeeringConnection(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	obj := state.ObjAsVpcPeering()

	// remote client not created
	if state.remoteClient == nil {
		return nil, nil
	}

	// skip loading of vpc peering connections if remoteId is empty
	if len(obj.Status.RemoteId) == 0 {
		return nil, nil
	}

	peering, err := state.remoteClient.DescribeVpcPeeringConnection(ctx, obj.Status.RemoteId)

	if err != nil {
		if composed.IsMarkedForDeletion(state.Obj()) {
			return composed.LogErrorAndReturn(err,
				"Error listing AWS peering connections but skipping as marked for deletion",
				nil,
				ctx)
		}

		if awsmeta.IsNotFound(err) {
			return nil, nil
		}

		logger.Error(err, "Error listing AWS peering connections")

		msg, isWarning := awsmeta.GetErrorMessage(err, "Error listing AWS peering connections")

		statusState := string(cloudcontrolv1beta1.StateError)

		if isWarning {
			statusState = string(cloudcontrolv1beta1.StateWarning)
		}

		changed := false

		if state.ObjAsVpcPeering().Status.State != statusState {
			state.ObjAsVpcPeering().SetState(statusState)
			changed = true
		}

		if meta.RemoveStatusCondition(state.ObjAsVpcPeering().Conditions(), cloudcontrolv1beta1.ConditionTypeReady) {
			changed = true
		}

		if meta.SetStatusCondition(state.ObjAsVpcPeering().Conditions(), metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  metav1.ConditionTrue,
			Reason:  cloudcontrolv1beta1.ReasonFailedLoadingRemoteVpcPeeringConnection,
			Message: msg,
		}) {
			changed = true
		}

		if !changed {
			return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
		}

		return composed.PatchStatus(state.ObjAsVpcPeering()).
			ErrorLogMessage("Error updating VpcPeering status when updating remote routes").
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T60000ms())).
			Run(ctx, state)

	}

	state.remoteVpcPeering = peering

	if state.remoteVpcPeering == nil {
		logger.Info("Remote AWS VPC peering connection not found", "remoteId", obj.Status.RemoteId)
		return nil, nil
	}

	return nil, composed.LoggerIntoCtx(ctx, logger.WithValues("remoteId", obj.Status.RemoteId))
}
