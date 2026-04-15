package subnet

import (
	"context"

	"cloud.google.com/go/networkconnectivity/apiv1/networkconnectivitypb"
	"github.com/google/uuid"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func updateConnectionPolicy(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.serviceConnectionPolicy == nil {
		return nil, ctx
	}

	if state.subnet == nil {
		return nil, ctx
	}

	subnet := state.ObjAsGcpSubnet()

	if !state.ShouldUpdateConnectionPolicy() {
		return nil, ctx
	}

	_, err := state.networkConnectivityClient.UpdateServiceConnectionPolicy(ctx, &networkconnectivitypb.UpdateServiceConnectionPolicyRequest{
		ServiceConnectionPolicy: state.serviceConnectionPolicy,
		RequestId:               uuid.NewString(),
		UpdateMask: &fieldmaskpb.FieldMask{
			Paths: state.updateMask,
		},
	})

	if err != nil {
		logger.Error(err, "Error updating GCP Connection Policy")
		meta.SetStatusCondition(subnet.Conditions(), metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  cloudcontrolv1beta1.ReasonCloudProviderError,
			Message: "Failed to update GCP Connection Policy",
		})
		subnet.Status.State = cloudcontrolv1beta1.StateError

		err = state.UpdateObjStatus(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err,
				"Error updating Subnet status due failed GCP Connection Policy update",
				composed.StopWithRequeueDelay((util.Timing.T10000ms())),
				ctx,
			)
		}

		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}

	return composed.StopWithRequeue, nil
}
