package vpcpeering

import (
	"context"
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createRemoteClient(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	// remote client can't be created if remote network is not found
	if state.remoteNetwork == nil {
		return nil, nil
	}

	remoteAccountId := state.remoteNetwork.Status.Network.Aws.AwsAccountId
	remoteRegion := state.remoteNetwork.Status.Network.Aws.Region

	roleArn := fmt.Sprintf("arn:aws:iam::%s:role/%s", remoteAccountId, state.roleName)

	composed.LoggerIntoCtx(ctx, logger.WithValues(
		"remoteAwsRegion", remoteRegion,
		"remoteAwsRole", roleArn,
	))

	logger.Info("Assuming remote AWS role")

	client, err := state.provider(
		ctx,
		remoteAccountId,
		remoteRegion,
		state.awsAccessKeyid,
		state.awsSecretAccessKey,
		roleArn,
	)

	if err != nil {
		logger.Error(err, "Error initializing remote AWS client")

		changed := false
		if meta.SetStatusCondition(state.ObjAsVpcPeering().Conditions(), metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  metav1.ConditionTrue,
			Reason:  cloudcontrolv1beta1.ConditionTypeError,
			Message: fmt.Sprintf("Failed creating AWS client for account %s", remoteAccountId),
		}) {
			changed = true
		}

		if state.ObjAsVpcPeering().Status.State != string(cloudcontrolv1beta1.ErrorState) {
			state.ObjAsVpcPeering().Status.State = string(cloudcontrolv1beta1.ErrorState)
			changed = true
		}

		if !changed {
			return composed.StopWithRequeueDelay(util.Timing.T300000ms()), nil
		}

		return composed.PatchStatus(state.ObjAsVpcPeering()).
			ErrorLogMessage("Error patching KCP VpcPeering with error state after remote client creation failed").
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T300000ms())). // try again in 5mins
			Run(ctx, state)
	}

	state.remoteClient = client

	return nil, ctx
}
