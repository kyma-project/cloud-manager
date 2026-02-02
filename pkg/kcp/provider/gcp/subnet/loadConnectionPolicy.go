package subnet

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	gcpmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func loadConnectionPolicy(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.serviceConnectionPolicy != nil {
		return nil, ctx
	}

	gcpScope := state.Scope().Spec.Scope.Gcp
	region := state.Scope().Spec.Region

	logger.Info("loading GCP Service Connection Policy")
	connectionPolicy, err := state.networkConnectivityClient.GetServiceConnectionPolicy(
		ctx,
		GetServiceConnectionPolicyFullName(gcpScope.Project, region, gcpScope.VpcNetwork),
	)

	if err != nil {
		if gcpmeta.IsNotFound(err) {
			logger.Info("target Service Connection Policy not found, continuing")
			return nil, ctx
		}

		logger.Error(err, "Error loading GCP Service Connection Policy")

		subnet := state.ObjAsGcpSubnet()
		meta.SetStatusCondition(subnet.Conditions(), metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  cloudcontrolv1beta1.ReasonCloudProviderError,
			Message: "Failed to load Service Connection Policy",
		})
		subnet.Status.State = cloudcontrolv1beta1.StateError

		err = state.UpdateObjStatus(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err,
				"Error updating Subnet status due failed Service Connection Policy loading",
				composed.StopWithRequeueDelay((util.Timing.T10000ms())),
				ctx,
			)
		}

		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}

	if connectionPolicy != nil {
		logger.Info("GCP Service Connection Policy found and loaded")
		state.serviceConnectionPolicy = connectionPolicy
	}

	return nil, ctx
}
