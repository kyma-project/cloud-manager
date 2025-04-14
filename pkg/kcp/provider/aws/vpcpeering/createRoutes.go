package vpcpeering

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/elliotchance/pie/v2"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	"k8s.io/apimachinery/pkg/api/meta"
)

func createRoutes(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	obj := state.ObjAsVpcPeering()

	for _, t := range state.routeTables {

		for _, cidrBlockAssociation := range state.remoteVpc.CidrBlockAssociationSet {

			cidrBlock := cidrBlockAssociation.CidrBlock

			routeExists := pie.Any(t.Routes, func(r types.Route) bool {
				return ptr.Equal(r.VpcPeeringConnectionId, state.vpcPeering.VpcPeeringConnectionId) &&
					ptr.Equal(r.DestinationCidrBlock, cidrBlock)
			})

			if routeExists {
				continue
			}

			lll := logger.WithValues(
				"routeTableId", ptr.Deref(t.RouteTableId, "xxx"),
				"destinationCidrBlock", ptr.Deref(cidrBlock, "xxx"))

			err := state.client.CreateRoute(ctx, t.RouteTableId, cidrBlock, state.remoteVpcPeering.VpcPeeringConnectionId)

			if err == nil {
				lll.Info("Route created")
				continue
			}

			lll.Error(err, "Error creating route")

			if awsmeta.IsErrorRetryable(err) {
				return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
			}

			changed := false
			if meta.RemoveStatusCondition(obj.Conditions(), cloudcontrolv1beta1.ConditionTypeReady) {
				changed = true
			}

			if meta.SetStatusCondition(obj.Conditions(), metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonFailedCreatingRoutes,
				Message: "Failed creating route for local route table",
			}) {
				changed = true
			}

			if obj.Status.State != string(cloudcontrolv1beta1.StateError) {
				obj.Status.State = string(cloudcontrolv1beta1.StateError)
				changed = true
			}

			if changed {
				// User can not recover from internal error
				return composed.PatchStatus(obj).
					ErrorLogMessage("Error updating VpcPeering status when creating local routes").
					SuccessError(composed.StopAndForget).
					Run(ctx, state)
			}

			// prevents alternating error messages
			return composed.StopAndForget, nil
		}
	}

	return nil, nil
}
