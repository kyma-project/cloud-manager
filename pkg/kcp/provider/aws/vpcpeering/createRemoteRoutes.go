package vpcpeering

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/elliotchance/pie/v2"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func createRemoteRoutes(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	obj := state.ObjAsVpcPeering()

	for _, t := range state.remoteRouteTables {
		routeExists := pie.Any(t.Routes, func(r types.Route) bool {
			return ptr.Equal(r.VpcPeeringConnectionId, state.vpcPeering.VpcPeeringConnectionId) &&
				ptr.Equal(r.DestinationCidrBlock, state.vpc.CidrBlock)
		})

		if !routeExists {
			err := state.remoteClient.CreateRoute(ctx, t.RouteTableId, state.vpc.CidrBlock, state.vpcPeering.VpcPeeringConnectionId)

			if err != nil {
				routeTableId := ptr.Deref(t.RouteTableId, "")

				logger.
					WithValues("remoteRouteTableId", routeTableId).
					Error(err, "Failed to create remote route")

				if awsmeta.IsErrorRetryable(err) {
					return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
				}

				changed := false

				if meta.SetStatusCondition(obj.Conditions(), metav1.Condition{
					Type:    cloudcontrolv1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  cloudcontrolv1beta1.ReasonFailedCreatingRoutes,
					Message: fmt.Sprintf("Failed creating route for remote route table %s. %s", routeTableId, awsmeta.GetErrorMessage(err)),
				}) {
					changed = true
				}

				if obj.Status.State != string(cloudcontrolv1beta1.StateWarning) {
					obj.Status.State = string(cloudcontrolv1beta1.StateWarning)
					changed = true
				}

				// Do not update status if nothing is changed
				if !changed {
					return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
				}

				// User can recover by modifying routes
				return composed.PatchStatus(obj).
					ErrorLogMessage("Error updating VpcPeering status when creating routes").
					SuccessError(composed.StopWithRequeueDelay(util.Timing.T60000ms())).
					Run(ctx, state)
			}
		}
	}

	return nil, nil
}
