package subnet

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	gcpmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/meta"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/subnet/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func loadSubnet(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.subnet != nil {
		return nil, ctx
	}

	gcpScope := state.Scope().Spec.Scope.Gcp
	region := state.Scope().Spec.Region

	logger.Info("loading GCP Subnet")
	subnet, err := state.computeClient.GetSubnet(ctx, client.GetSubnetRequest{
		ProjectId: gcpScope.Project,
		Region:    region,
		Name:      GetSubnetShortName(state.Obj().GetName()),
	})

	if err != nil {
		if gcpmeta.IsNotFound(err) {
			logger.Info("target Subnet not found, continuing")
			return nil, ctx
		}

		logger.Error(err, "Error loading GCP Private Subnet")

		subnet := state.ObjAsGcpSubnet()
		meta.SetStatusCondition(subnet.Conditions(), metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  cloudcontrolv1beta1.ReasonCloudProviderError,
			Message: "Failed to load Subnet",
		})
		subnet.Status.State = cloudcontrolv1beta1.StateError

		err = state.UpdateObjStatus(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err,
				"Error updating Subnet status due failed GCP Subnet loading",
				composed.StopWithRequeueDelay((util.Timing.T10000ms())),
				ctx,
			)
		}

		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}

	if subnet != nil {
		logger.Info("GCP Private Subnet found and loaded")
		state.subnet = subnet
	}

	return nil, ctx
}
