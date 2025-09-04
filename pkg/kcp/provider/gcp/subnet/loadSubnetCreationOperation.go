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

func loadSubnetCreationOperation(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	subnet := state.ObjAsGcpSubnet()

	if subnet.Status.SubnetCreationOperationName == "" {
		return nil, ctx
	}

	gcpScope := state.Scope().Spec.Scope.Gcp
	region := state.Scope().Spec.Region

	logger.Info("loading GCP Subnet Creation Operation")

	op, err := state.regionOperationsClient.GetRegionOperation(ctx, client.GetRegionOperationRequest{
		ProjectId: gcpScope.Project,
		Region:    region,
		Name:      subnet.Status.SubnetCreationOperationName,
	})

	if err != nil {
		if gcpmeta.IsNotFound(err) {
			logger.Info("target GCP Subnet Creation Operation not found, continuing")
			return nil, ctx
		}

		logger.Error(err, "Error loading GCP GCP Subnet Creation Operation")

		subnet := state.ObjAsGcpSubnet()
		meta.SetStatusCondition(subnet.Conditions(), metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  cloudcontrolv1beta1.ReasonCloudProviderError,
			Message: "Failed to load GCP Subnet Creation Operation",
		})
		subnet.Status.State = cloudcontrolv1beta1.StateError

		err = state.UpdateObjStatus(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err,
				"Error updating Subnet status due failed GCP GCP Subnet Creation Operation loading",
				composed.StopWithRequeueDelay((util.Timing.T10000ms())),
				ctx,
			)
		}

		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}

	if op != nil {
		logger.Info("GCP GCP Subnet Creation Operation found and loaded")
		state.subnetCreationOperation = op
	}

	return nil, ctx
}
