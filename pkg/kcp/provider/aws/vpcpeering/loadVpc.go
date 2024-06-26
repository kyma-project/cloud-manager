package vpcpeering

import (
	"context"
	"fmt"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"strings"
)

func loadVpc(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	obj := state.ObjAsVpcPeering()

	if len(obj.Status.Id) != 0 {
		return nil, nil
	}

	vpcNetworkName := state.Scope().Spec.Scope.Aws.VpcNetwork

	vpcList, err := state.client.DescribeVpcs(ctx)

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error loading AWS VPC Networks", composed.StopWithRequeue, ctx)
	}

	var vpc *ec2Types.Vpc
	var allLoadedVpcs []string
	for _, vv := range vpcList {
		v := vv
		var sb strings.Builder
		for _, t := range v.Tags {
			sb.WriteString(pointer.StringDeref(t.Key, ""))
			sb.WriteString("=")
			sb.WriteString(pointer.StringDeref(t.Value, ""))
			sb.WriteString(",")
		}
		allLoadedVpcs = append(allLoadedVpcs, fmt.Sprintf(
			"%s{%s}",
			pointer.StringDeref(v.VpcId, ""),
			sb.String(),
		))
		if util.NameEc2TagEquals(v.Tags, vpcNetworkName) {
			vpc = &v
		}
	}

	if vpc == nil {
		logger.
			WithValues(
				"vpcName", vpcNetworkName,
				"allLoadedVpcs", fmt.Sprintf("%v", allLoadedVpcs),
			).
			Info("VPC not found")

		return composed.UpdateStatus(state.ObjAsVpcPeering()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonVpcNotFound,
				Message: fmt.Sprintf("AWS VPC %s not found", vpcNetworkName),
			}).
			ErrorLogMessage("Error updating VpcPeering status when loading vpc").
			SuccessError(composed.StopAndForget).
			Run(ctx, st)
	}

	state.vpc = vpc

	state.ObjAsVpcPeering().Status.VpcId = pointer.StringDeref(vpc.VpcId, "")

	logger = logger.WithValues("vpcId", pointer.StringDeref(state.vpc.VpcId, ""))
	ctx = composed.LoggerIntoCtx(ctx, logger)

	return nil, ctx
}
