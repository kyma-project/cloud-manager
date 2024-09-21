package vpcpeering

import (
	"context"
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func acceptVpcPeeringConnection(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	obj := state.ObjAsVpcPeering()

	if state.remoteVpcPeering != nil {
		return nil, nil
	}

	remoteAccountId := state.remoteNetwork.Spec.Network.Reference.Aws.AwsAccountId
	remoteRegion := state.remoteNetwork.Spec.Network.Reference.Aws.Region

	roleArn := fmt.Sprintf("arn:aws:iam::%s:role/%s", remoteAccountId, state.roleName)

	logger = logger.WithValues("awsRegion", remoteRegion, "awsRole", roleArn)

	logger.Info("Assuming AWS role")

	client, err := state.provider(
		ctx,
		remoteRegion,
		state.awsAccessKeyid,
		state.awsSecretAccessKey,
		roleArn,
	)
	if err != nil {
		logger.Error(err, "Failed to create aws acceptVpcPeeringConnection client")
		return composed.StopWithRequeueDelay(util.Timing.T300000ms()), nil
	}

	peering, err := client.AcceptVpcPeeringConnection(ctx,
		state.vpcPeering.VpcPeeringConnectionId)

	if err != nil {
		logger.Error(err, "Error accepting VPC Peering")

		obj.Status.State = string(cloudcontrolv1beta1.ErrorState)

		return composed.PatchStatus(obj).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  "True",
				Reason:  cloudcontrolv1beta1.ReasonFailedAcceptingVpcPeeringConnection,
				Message: fmt.Sprintf("Failed accepting VpcPeerings %s", err),
			}).
			ErrorLogMessage("Error updating VpcPeering status due to failed accepting vpc peering connection").
			FailedError(composed.StopWithRequeue).
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T60000ms())).
			Run(ctx, state)
	}

	logger = logger.WithValues("remotePeeringId", *peering.VpcPeeringConnectionId)

	ctx = composed.LoggerIntoCtx(ctx, logger)

	logger.Info("AWS VPC Peering Connection accepted")

	state.remoteVpcPeering = peering

	obj.Status.RemoteId = *peering.VpcPeeringConnectionId

	return composed.UpdateStatus(obj).
		ErrorLogMessage("Error updating VpcPeering status with remote connection id").
		FailedError(composed.StopWithRequeue).
		SuccessErrorNil().
		Run(ctx, state)
}
