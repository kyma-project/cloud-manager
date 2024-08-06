package vpcpeering

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azureconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/config"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

func loadRemoteVpc(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	obj := state.ObjAsVpcPeering()

	if state.remotePeering != nil {
		return nil, nil
	}

	clientId := azureconfig.AzureConfig.VpcPeeringClientId
	clientSecret := azureconfig.AzureConfig.VpcPeeringClientSecret
	tenantId := state.tenantId

	remote, err := util.ParseResourceID(obj.Spec.VpcPeering.Azure.RemoteVnet)

	if err != nil {
		logger.Error(err, "Error parsing remoteVnet")
		return err, ctx
	}

	subscriptionId := remote.Subscription

	c, err := state.provider(ctx, clientId, clientSecret, subscriptionId, tenantId)

	if err != nil {
		return err, ctx
	}

	network, err := c.GetNetwork(ctx, remote.ResourceGroup, remote.ResourceName)

	if err != nil {
		logger.Error(err, "Error loading remote network")

		message := azuremeta.GetErrorMessage(err)

		successError := composed.StopWithRequeueDelay(time.Minute)

		// If VpcNetwork is not found user can not recover from this error without updating the resource so, we are doing
		// stop and forget.
		if azuremeta.IsNotFound(err) {
			successError = composed.StopAndForget
			message = "Remote VPC Network not found"
			logger.Info(message)
		}

		return composed.UpdateStatus(obj).
			SetCondition(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  "True",
				Reason:  cloudcontrolv1beta1.ReasonFailedLoadingRemoteVpcNetwork,
				Message: message,
			}).
			ErrorLogMessage("Error updating VpcPeering status due to failed loading of remote vpc network").
			FailedError(composed.StopWithRequeue).
			SuccessError(successError).
			Run(ctx, state)
	}

	state.remoteVpc = network

	return nil, nil
}
