package vpcpeering

import (
	"context"

	pb "cloud.google.com/go/compute/apiv1/computepb"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	gcpmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func checkRemoteOperation(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if composed.IsMarkedForDeletion(state.ObjAsVpcPeering()) ||
		state.ObjAsVpcPeering().Status.RemoteId != "" ||
		state.remotePeeringOperation != nil ||
		state.ObjAsVpcPeering().Status.RemotePeeringOperation == "" {
		return nil, ctx
	}

	op, err := state.client.GetOperation(ctx, state.RemoteNetwork().Spec.Network.Reference.Gcp.GcpProject, state.ObjAsVpcPeering().Status.RemotePeeringOperation)
	if err != nil {
		if gcpmeta.IsNotAuthorized(err) {
			return composed.UpdateStatus(state.ObjAsVpcPeering()).
				SetExclusiveConditions(metav1.Condition{
					Type:    cloudcontrolv1beta1.ConditionTypeError,
					Status:  "True",
					Reason:  cloudcontrolv1beta1.ReasonUnauthorized,
					Message: "Error fetching GCP remote peering operation due to insufficient permissions, please check the documentation for more details",
				}).
				ErrorLogMessage("Error updating VPC Peering while fetching remote operation due to insufficient permissions").
				FailedError(composed.StopWithRequeue).
				SuccessError(composed.StopWithRequeueDelay(5*util.Timing.T60000ms())).
				Run(ctx, state)
		}
		logger.Error(err, "Error getting remote operation")
		meta.SetStatusCondition(state.ObjAsVpcPeering().Conditions(), metav1.Condition{
			Type:    v1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  v1beta1.ConditionReasonError,
			Message: "Error loading Remote Vpc Peering Operation: " + state.ObjAsVpcPeering().Status.RemotePeeringOperation,
		})
		err = state.PatchObjStatus(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err,
				"Error updating status since it was not possible to load the remote Vpc Peering operation.",
				composed.StopWithRequeueDelay(util.Timing.T10000ms()),
				ctx,
			)
		}
		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}
	state.remotePeeringOperation = op
	if op != nil && op.GetStatus() != pb.Operation_DONE {
		util.ExpiringSwitch().Key(op.GetStatus().String(), pb.Operation_RUNNING.String()).IfNotRecently(func() {
			logger.Info("Remote operation still in progress", "remotePeeringOperation", ptr.Deref(op.Name, "OperationUnknown"))
		})
		return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
	}
	return nil, ctx
}
