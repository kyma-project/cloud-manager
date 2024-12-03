package vpcpeering

import (
	"context"
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	awsutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func loadRemoteVpc(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	obj := state.ObjAsVpcPeering()

	// remote client not created
	if state.remoteClient == nil {
		return nil, nil
	}

	remoteVpcId := state.remoteNetwork.Status.Network.Aws.VpcId

	vpc, err := state.remoteClient.DescribeVpc(ctx, remoteVpcId)

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

		logger.Error(err, "Error loading remote AWS VPC Networks")

		condition := metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  metav1.ConditionTrue,
			Reason:  cloudcontrolv1beta1.ReasonVpcNotFound,
			Message: awsmeta.GetErrorMessage(err),
		}

		successError := composed.StopAndForget

		// User can recover by setting permissions
		if awsmeta.IsUnauthorized(err) {
			successError = composed.StopWithRequeueDelay(util.Timing.T60000ms())
		}

		if !composed.AnyConditionChanged(obj, condition) {
			return successError, nil
		}

		return composed.PatchStatus(obj).
			SetExclusiveConditions(condition).
			ErrorLogMessage("Error updating VpcPeering status when loading vpc").
			SuccessError(successError).
			Run(ctx, st)
	}

	if vpc == nil {
		logger.
			WithValues(
				"remoteVpcId", remoteVpcId,
			).
			Info("VPC not found")

		condition := metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  metav1.ConditionTrue,
			Reason:  cloudcontrolv1beta1.ReasonVpcNotFound,
			Message: fmt.Sprintf("AWS VPC ID %s not found", remoteVpcId),
		}

		if !composed.AnyConditionChanged(obj, condition) {
			return composed.StopAndForget, nil
		}

		return composed.PatchStatus(obj).
			SetExclusiveConditions(condition).
			ErrorLogMessage("Error updating VpcPeering status when loading vpc").
			SuccessError(composed.StopAndForget).
			Run(ctx, st)
	}

	state.remoteVpc = vpc

	ctx = composed.LoggerIntoCtx(ctx, logger.WithValues(
		"remoteVpcId", remoteVpcId,
		"remoteVpcName", awsutil.GetEc2TagValue(vpc.Tags, "Name"),
	))

	return nil, ctx
}
