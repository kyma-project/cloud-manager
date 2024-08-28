package vpcpeering

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/elliotchance/pie/v2"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func createRemoteRoutes(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	obj := state.ObjAsVpcPeering()

	remoteAccountId := obj.Spec.VpcPeering.Aws.RemoteAccountId
	remoteRegion := obj.Spec.VpcPeering.Aws.RemoteRegion

	roleArn := fmt.Sprintf("arn:aws:iam::%s:role/%s", remoteAccountId, state.roleName)

	logger.WithValues(
		"remoteAwsRegion", remoteRegion,
		"remoteAwsRole", roleArn,
	).Info("Assuming remote AWS role")

	client, err := state.provider(
		ctx,
		remoteRegion,
		state.awsAccessKeyid,
		state.awsSecretAccessKey,
		roleArn,
	)

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error initializing remote AWS client", composed.StopWithRequeue, ctx)
	}

	for _, t := range state.remoteRouteTables {
		routeExists := pie.Any(t.Routes, func(r types.Route) bool {
			return ptr.Equal(r.VpcPeeringConnectionId, state.vpcPeering.VpcPeeringConnectionId) &&
				ptr.Equal(r.DestinationCidrBlock, state.vpc.CidrBlock)
		})

		if !routeExists {
			err = client.CreateRoute(ctx, t.RouteTableId, state.vpc.CidrBlock, state.vpcPeering.VpcPeeringConnectionId)

			if err != nil {
				routeTableId := ptr.Deref(t.RouteTableId, "")

				logger.
					WithValues(
						"remoteRouteTableId", routeTableId,
					).
					Error(err, "Failed to create route")

				condition := metav1.Condition{
					Type:    cloudcontrolv1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  cloudcontrolv1beta1.ReasonFailedCreatingRoutes,
					Message: fmt.Sprintf("AWS Failed to create route for remote route table %s", routeTableId),
				}

				if !awsmeta.AnyConditionChanged(obj, condition) {
					return composed.StopAndForget, nil
				}

				return composed.UpdateStatus(obj).
					SetExclusiveConditions(condition).
					ErrorLogMessage("Error updating VpcPeering status when creating routes").
					SuccessError(composed.StopAndForget).
					Run(ctx, state)
			}
		}
	}

	return nil, nil
}
