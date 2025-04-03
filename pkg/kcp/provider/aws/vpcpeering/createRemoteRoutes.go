package vpcpeering

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/elliotchance/pie/v2"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	awsutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/util"
	peeringconfig "github.com/kyma-project/cloud-manager/pkg/kcp/vpcpeering/config"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func createRemoteRoutes(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if awsutil.IsRouteTableUpdateStrategyNone(state.ObjAsVpcPeering().Spec.Details.RemoteRouteTableUpdateStrategy) {
		return nil, nil
	}

	for _, t := range state.remoteRouteTables {

		shouldUpdateRouteTable := awsutil.ShouldUpdateRouteTable(t.Tags,
			state.ObjAsVpcPeering().Spec.Details.RemoteRouteTableUpdateStrategy,
			state.Scope().Spec.ShootName)

		cidrBlockAssociations := pie.Filter(state.vpc.CidrBlockAssociationSet, func(cidrBlockAssociation types.VpcCidrBlockAssociation) bool {
			return peeringconfig.VpcPeeringConfig.RouteAsociatedCidrBlocks || ptr.Equal(cidrBlockAssociation.CidrBlock, state.vpc.CidrBlock)
		})

		for _, cidrBlockAssociation := range cidrBlockAssociations {

			cidrBlock := cidrBlockAssociation.CidrBlock

			routeExists := pie.Any(t.Routes, func(r types.Route) bool {
				return ptr.Equal(r.VpcPeeringConnectionId, state.vpcPeering.VpcPeeringConnectionId) &&
					ptr.Equal(r.DestinationCidrBlock, cidrBlock)
			})

			var err error

			lll := logger.WithValues(
				"remoteRouteTableId", ptr.Deref(t.RouteTableId, "xxx"),
				"destinationCidrBlock", ptr.Deref(cidrBlock, "xxx"))

			// Create route if it should exist but it doesn't
			if shouldUpdateRouteTable && !routeExists {
				err = state.remoteClient.CreateRoute(ctx, t.RouteTableId, cidrBlock, state.vpcPeering.VpcPeeringConnectionId)
				if err != nil {
					lll.Error(err, "Error creating remote route")
				} else {
					lll.Info("Remote route created")
				}
			}

			if err != nil {

				if awsmeta.IsErrorRetryable(err) {
					return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
				}

				successError := composed.StopWithRequeueDelay(util.Timing.T60000ms())

				if awsmeta.IsRouteNotSupported(err) {
					successError = composed.StopAndForget
				}

				changed := false

				msg, _ := awsmeta.GetErrorMessage(err, "")
				if meta.SetStatusCondition(state.ObjAsVpcPeering().Conditions(), metav1.Condition{
					Type:    cloudcontrolv1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  cloudcontrolv1beta1.ReasonFailedCreatingRoutes,
					Message: fmt.Sprintf("Failed updating routes for remote route table %s. %s", ptr.Deref(t.RouteTableId, ""), msg),
				}) {
					changed = true
				}

				if state.ObjAsVpcPeering().Status.State != string(cloudcontrolv1beta1.StateWarning) {
					state.ObjAsVpcPeering().Status.State = string(cloudcontrolv1beta1.StateWarning)
					changed = true
				}

				// Do not update status if nothing is changed
				if !changed {
					return successError, nil
				}

				// User can recover by modifying routes
				return composed.PatchStatus(state.ObjAsVpcPeering()).
					ErrorLogMessage("Error updating VpcPeering status when updating routes").
					SuccessError(successError).
					Run(ctx, state)
			}
		}
	}

	return nil, nil
}
