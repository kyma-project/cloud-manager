package v3

import (
	"context"

	"github.com/google/uuid"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	v3 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange/v3/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func createConnectionPolicy(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.serviceConnectionPolicy != nil {
		return nil, nil
	}

	if state.subnet == nil {
		logger.Info("Subnet not loaded, requeueing")
		return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
	}

	ipRange := state.ObjAsIpRange()
	gcpScope := state.Scope().Spec.Scope.Gcp
	region := state.Scope().Spec.Region

	err := state.networkComnnectivityClient.CreateServiceConnectionPolicy(ctx, v3.CreateServiceConnectionPolicyRequest{
		ProjectId:     gcpScope.Project,
		Region:        region,
		Network:       gcpScope.VpcNetwork,
		Name:          GetServiceConnectionPolicyShortName(gcpScope.VpcNetwork, region),
		Subnets:       []string{GetSubnetFullName(gcpScope.Project, region, ptr.Deref(state.subnet.Name, ""))},
		IdempotenceId: uuid.NewString(),
	})

	if err != nil {
		logger.Error(err, "Error creating GCP Connection Policy")
		meta.SetStatusCondition(ipRange.Conditions(), metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  cloudcontrolv1beta1.ReasonCloudProviderError,
			Message: "Failed to create GCP Connection Policy",
		})
		ipRange.Status.State = cloudcontrolv1beta1.StateError

		err = state.UpdateObjStatus(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err,
				"Error updating IpRange status due failed GCP Connection Policy creation",
				composed.StopWithRequeueDelay((util.Timing.T10000ms())),
				ctx,
			)
		}

		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}

	return composed.StopWithRequeue, nil
}
