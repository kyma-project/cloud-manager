package vpcpeering

import (
	"context"
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/util"
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

	remoteVpcId := state.remoteNetwork.Spec.Network.Reference.Aws.VpcId

	vpc, err := state.remoteClient.DescribeVpc(ctx, remoteVpcId)

	if awsmeta.IsErrorRetryable(err) {
		return awsmeta.LogErrorAndReturn(err, "Error loading remote AWS VPC Network", ctx)
	}
	if err != nil {
		logger.Error(err, "Error loading remote AWS VPC Networks")

		if !composed.SetExclusiveConditions(obj.Conditions(), metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  metav1.ConditionTrue,
			Reason:  cloudcontrolv1beta1.ReasonVpcNotFound,
			Message: err.Error(),
		}) {
			return composed.StopAndForget, nil
		}

		return composed.PatchStatus(obj).
			ErrorLogMessage("Error updating VpcPeering status when loading vpc").
			SuccessError(composed.StopAndForget).
			Run(ctx, st)
	}

	if vpc == nil {
		logger.
			WithValues(
				"remoteVpcId", remoteVpcId,
			).
			Info("VPC not found")

		if !composed.SetExclusiveConditions(obj.Conditions(), metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  metav1.ConditionTrue,
			Reason:  cloudcontrolv1beta1.ReasonVpcNotFound,
			Message: fmt.Sprintf("AWS VPC ID %s not found", remoteVpcId),
		}) {
			return composed.StopAndForget, nil
		}

		return composed.PatchStatus(obj).
			ErrorLogMessage("Error updating VpcPeering status when loading vpc").
			SuccessError(composed.StopAndForget).
			Run(ctx, st)
	}

	state.remoteVpc = vpc

	ctx = composed.LoggerIntoCtx(ctx, logger.WithValues(
		"remoteVpcId", remoteVpcId,
		"remoteVpcName", util.GetEc2TagValue(vpc.Tags, "Name"),
	))

	return nil, ctx
}
