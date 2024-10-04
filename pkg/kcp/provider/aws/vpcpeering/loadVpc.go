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

func loadVpc(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	obj := state.ObjAsVpcPeering()

	vpcNetworkName := state.Scope().Spec.Scope.Aws.VpcNetwork

	vpcList, err := state.client.DescribeVpcs(ctx, vpcNetworkName)

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error loading AWS VPC Networks", composed.StopWithRequeue, ctx)
	}

	var vpc *ec2Types.Vpc

	for _, vv := range vpcList {
		v := vv
		if util.NameEc2TagEquals(v.Tags, vpcNetworkName) {
			vpc = &v
			break
		}
	}

	if vpc == nil {

		allLoadedVpcs := pie.StringsUsing(vpcList, func(x ec2Types.Vpc) string {
			return fmt.Sprintf(
				"%s{%s}",
				ptr.Deref(x.VpcId, ""),
				util.TagsToString(x.Tags))
		})

		logger.
			WithValues(
				"vpcName", vpcNetworkName,
				"allLoadedVpcs", fmt.Sprintf("%v", allLoadedVpcs),
			).
			Info("VPC not found")

		return composed.PatchStatus(obj).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonVpcNotFound,
				Message: fmt.Sprintf("AWS VPC %s not found", vpcNetworkName),
			}).
			ErrorLogMessage("Error updating VpcPeering status when loading vpc").
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	}

	state.vpc = vpc

	vpcId := ptr.Deref(vpc.VpcId, "")

	state.ObjAsVpcPeering().Status.VpcId = vpcId

	logger = logger.WithValues("vpcId", vpcId)

	ctx = composed.LoggerIntoCtx(ctx, logger)

	return nil, ctx
}
