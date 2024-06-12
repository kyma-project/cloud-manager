package vpcpeering

import (
	"context"
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azureconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/config"
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

	clientId := azureconfig.AzureConfig.ClientId
	clientSecret := azureconfig.AzureConfig.ClientSecret
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

		return composed.UpdateStatus(obj).
			SetCondition(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  "True",
				Reason:  cloudcontrolv1beta1.ReasonFailedLoadingRemoteVpcNetwork,
				Message: fmt.Sprintf("Failed loading VpcNetwork %s", err),
			}).
			ErrorLogMessage("Error updating VpcPeering status due to failed loading of remote vpc network").
			FailedError(composed.StopWithRequeue).
			SuccessError(composed.StopWithRequeueDelay(time.Minute)).
			Run(ctx, state)
	}

	// If VpcNetwork is not found user can not recover from this error without updating the resource so, we are doing
	// stop and forget.
	if network == nil {
		logger.Info("Remote VPC Network not found")

		return composed.UpdateStatus(obj).
			SetCondition(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  "True",
				Reason:  cloudcontrolv1beta1.ReasonFailedLoadingRemoteVpcNetwork,
				Message: fmt.Sprintf("Failed loading VpcNetwork %s", err),
			}).
			ErrorLogMessage("Error updating VpcPeering status due to failed loading of remote vpc network").
			FailedError(composed.StopWithRequeue).
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	}

	state.remoteVpc = network

	return nil, nil
}
