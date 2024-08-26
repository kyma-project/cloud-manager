package vpcpeering

import (
	"context"
	"fmt"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/elliotchance/pie/v2"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func loadRouteTables(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	obj := state.ObjAsVpcPeering()

	vpcNetworkName := state.Scope().Spec.Scope.Aws.VpcNetwork

	routeTables, err := state.client.DescribeRouteTables(ctx, *state.vpc.VpcId)

	if err != nil {
		return awsmeta.LogErrorAndReturn(err, "Error loading AWS route tables", ctx)
	}

	allLoadedRouteTables := pie.StringsUsing(routeTables, func(t ec2Types.RouteTable) string {
		return fmt.Sprintf(
			"%s{%s}",
			ptr.Deref(t.RouteTableId, ""),
			util.TagsToString(t.Tags))
	})

	state.routeTables = pie.Filter(routeTables, func(t ec2Types.RouteTable) bool {
		return util.HasEc2Tag(t.Tags, fmt.Sprintf("kubernetes.io/cluster/%s", vpcNetworkName))
	})

	if len(state.routeTables) == 0 {
		logger.
			WithValues(
				"vpcId", *state.vpc.VpcId,
				"allLoadedRouteTables", fmt.Sprintf("%v", allLoadedRouteTables),
			).
			Info("Route tables not found")

		return composed.UpdateStatus(obj).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonFailedLoadingRouteTables,
				Message: fmt.Sprintf("AWS VPC %s route tables not found", *state.vpc.VpcId),
			}).
			ErrorLogMessage("Error updating VpcPeering status when loading route tables").
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	}

	return nil, nil
}
