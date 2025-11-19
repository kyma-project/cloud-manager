package vpcpeering

import (
	"context"

	pb "cloud.google.com/go/compute/apiv1/computepb"
	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func checkRemoteOperation(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if composed.MarkedForDeletionPredicate(ctx, state) || (state.remoteOperation != nil && state.remoteOperation.GetStatus() != pb.Operation_PENDING) {
		return nil, ctx
	}

	if state.ObjAsVpcPeering().Status.RemoteOperation != "" {
		op, err := state.client.GetOperation(ctx, state.RemoteNetwork().Spec.Network.Reference.Gcp.GcpProject, state.ObjAsVpcPeering().Status.RemoteOperation)
		if err != nil {
			logger.Error(err, "[KCP GCP VpcPeering checkRemoteOperation] Error getting remote operation")
			meta.SetStatusCondition(state.ObjAsVpcPeering().Conditions(), metav1.Condition{
				Type:    v1beta1.ConditionTypeError,
				Status:  "True",
				Reason:  v1beta1.ConditionReasonError,
				Message: "Error loading Remote Vpc Peering Operation: " + state.ObjAsVpcPeering().Status.RemoteOperation,
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
		state.remoteOperation = op
		if op != nil {
			if op.GetStatus() != pb.Operation_DONE {
				logger.Info("[KCP GCP VpcPeering checkRemoteOperation] Remote operation still in progress", "remoteOperation", ptr.Deref(op.Name, "OperationUnknown"))
				return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
			}
			return nil, ctx
		}
	}

	return nil, nil
}
