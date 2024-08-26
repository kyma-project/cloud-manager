package vpcpeering

import (
	"context"
	"fmt"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/elliotchance/pie/v2"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func loadRemoteRouteTables(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	obj := state.ObjAsVpcPeering()
	logger := composed.LoggerFromCtx(ctx)

	remoteAccountId := obj.Spec.VpcPeering.Aws.RemoteAccountId
	remoteRegion := obj.Spec.VpcPeering.Aws.RemoteRegion

	roleArn := fmt.Sprintf("arn:aws:iam::%s:role/%s", remoteAccountId, state.roleName)

	logger.WithValues(
		"awsRegion", remoteRegion,
		"awsRole", roleArn,
	).Info("Assuming AWS role")

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

	routeTables, err := client.DescribeRouteTables(ctx, *state.remoteVpc.VpcId)

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error loading AWS VPC Networks", composed.StopWithRequeue, ctx)
	}

	allLoadedRouteTables := pie.StringsUsing(routeTables, func(t ec2Types.RouteTable) string {
		return fmt.Sprintf(
			"%s{%s}",
			ptr.Deref(t.RouteTableId, ""),
			util.TagsToString(t.Tags))
	})

	state.remoteRouteTables = routeTables

	if len(state.remoteRouteTables) == 0 {
		logger.
			WithValues(
				"vpcId", *state.remoteVpc.VpcId,
				"allLoadedRouteTables", fmt.Sprintf("%v", allLoadedRouteTables),
			).
			Info("Route tables not found")

		return composed.UpdateStatus(obj).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonFailedLoadingRouteTables,
				Message: fmt.Sprintf("AWS VPC %s route tables not found", *state.remoteVpc.VpcId),
			}).
			ErrorLogMessage("Error updating VpcPeering status when loading remote route tables").
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	}

	return nil, nil
}
