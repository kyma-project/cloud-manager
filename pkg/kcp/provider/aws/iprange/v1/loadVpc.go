package v1

import (
	"context"
	"fmt"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"strings"
)

func loadVpc(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	vpcNetworkName := state.Scope().Spec.Scope.Aws.VpcNetwork
	vpcList, err := state.client.DescribeVpcs(ctx)
	if err != nil {
		return awsmeta.LogErrorAndReturn(err, "Error loading AWS VPC Networks", ctx)
	}

	var allLoadedVpcs []string
	for _, vv := range vpcList {
		v := vv
		var sb strings.Builder
		for _, t := range v.Tags {
			sb.WriteString(ptr.Deref(t.Key, ""))
			sb.WriteString("=")
			sb.WriteString(ptr.Deref(t.Value, ""))
			sb.WriteString(",")
		}
		allLoadedVpcs = append(allLoadedVpcs, fmt.Sprintf(
			"%s{%s}",
			ptr.Deref(v.VpcId, ""),
			sb.String(),
		))
	}

	var vpc *ec2Types.Vpc
	for _, vv := range vpcList {
		// loop var will change it's value, and we're taking a pointer to it below
		// MUST make a copy to another var that will not change the value
		v := vv
		if util.NameEc2TagEquals(v.Tags, vpcNetworkName) {
			vpc = &v
			break
		}
	}

	if vpc == nil {
		logger.
			WithValues(
				"vpcName", vpcNetworkName,
				"allLoadedVpcs", fmt.Sprintf("%v", allLoadedVpcs),
			).
			Info("VPC not found")

		return composed.UpdateStatus(state.ObjAsIpRange()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonVpcNotFound,
				Message: fmt.Sprintf("AWS VPC %s not found", vpcNetworkName),
			}).
			ErrorLogMessage("Error updating IpRange status when loading vpc").
			SuccessError(composed.StopAndForget).
			Run(ctx, st)
	}

	state.vpc = vpc
	state.ObjAsIpRange().Status.VpcId = ptr.Deref(vpc.VpcId, "") // will be saved when subnets are created

	logger = logger.WithValues("vpcId", ptr.Deref(state.vpc.VpcId, ""))
	ctx = composed.LoggerIntoCtx(ctx, logger)

	return nil, ctx
}
