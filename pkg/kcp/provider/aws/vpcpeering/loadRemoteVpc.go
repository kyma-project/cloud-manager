package vpcpeering

import (
	"context"
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	awsutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func loadRemoteVpc(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	// remote client not created
	if state.remoteClient == nil {
		return nil, nil
	}

	remoteVpcId := state.RemoteNetwork().Status.Network.Aws.VpcId

	vpc, err := state.remoteClient.DescribeVpc(ctx, remoteVpcId)

	result := composed.StopAndForget

	if err != nil {
		if composed.IsMarkedForDeletion(state.Obj()) {
			return composed.LogErrorAndReturn(err,
				"Error loading remote AWS VPC Network but skipping as marked for deletion",
				nil,
				ctx)
		}

		if awsmeta.IsErrorRetryable(err) {
			return awsmeta.LogErrorAndReturn(err, "Error loading remote AWS VPC Network", ctx)
		}

		// User can recover by setting permissions
		if awsmeta.IsUnauthorized(err) || awsmeta.IsAccessDenied(err) {
			result = composed.StopWithRequeueDelay(util.Timing.T60000ms())
		}

		logger.WithValues("remoteVpcId", remoteVpcId).
			Error(err, "Error loading remote AWS VPC Network")

		msg, isWarning := awsmeta.GetErrorMessage(err, "Error loading remote VPC Network")

		if awsmeta.IsNotFound(err) {
			msg = fmt.Sprintf("Remote VPC ID %s not found", remoteVpcId)
		}

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

		if meta.SetStatusCondition(state.ObjAsVpcPeering().Conditions(),
			metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonVpcNotFound,
				Message: msg,
			}) {
			changed = true
		}

		if !changed {
			return result, nil
		}

		return composed.PatchStatus(state.ObjAsVpcPeering()).
			ErrorLogMessage("Error updating VpcPeering status when loading vpc").
			SuccessError(result).
			Run(ctx, state)
	}

	state.remoteVpc = vpc

	return nil, composed.LoggerIntoCtx(ctx, logger.WithValues(
		"remoteVpcId", remoteVpcId,
		"remoteVpcName", awsutil.GetEc2TagValue(vpc.Tags, "Name"),
	))
}
