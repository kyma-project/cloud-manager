package v2

import (
	"context"
	"errors"
	"fmt"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awserrorhandling "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/errorhandling"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"strings"
)

func vpcFind(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.vpc != nil {
		return nil, nil
	}

	vpcNetworkName := state.Scope().Spec.Scope.Aws.VpcNetwork

	vpcList, err := state.awsClient.DescribeVpcs(ctx, vpcNetworkName)
	if x := awserrorhandling.HandleError(ctx, err, state, "KCP IpRange on list AWS VPC networks",
		cloudcontrolv1beta1.ReasonUnknown, "Error getting AWS VPC network"); x != nil {
		return x, nil
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
		// loop var will change its value, and we're taking a pointer to it below
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
			Error(errors.New("VPC not found"), "VPC not found for KCP IpRange")

		return composed.PatchStatus(state.ObjAsIpRange()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  "True",
				Reason:  cloudcontrolv1beta1.ReasonVpcNotFound,
				Message: fmt.Sprintf("AWS VPC %s not found", vpcNetworkName),
			}).
			ErrorLogMessage("Error patching KCP IpRange status when vpc not found").
			SuccessLogMsg("Forgetting KCP IpRange with AWS VPC not found").
			SuccessError(composed.StopAndForget).
			Run(ctx, st)
	}

	state.vpc = vpc
	state.ObjAsIpRange().Status.VpcId = ptr.Deref(vpc.VpcId, "")

	logger = logger.WithValues("vpcId", ptr.Deref(state.vpc.VpcId, ""))
	ctx = composed.LoggerIntoCtx(ctx, logger)

	return composed.PatchStatus(state.ObjAsIpRange()).
		SuccessErrorNil().
		Run(ctx, st)
}
