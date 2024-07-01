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

func loadRemoteVpc(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	remoteVpcId := state.ObjAsVpcPeering().Spec.VpcPeering.Aws.RemoteVpcId
	remoteAccountId := state.ObjAsVpcPeering().Spec.VpcPeering.Aws.RemoteAccountId
	state.remoteRegion = state.ObjAsVpcPeering().Spec.VpcPeering.Aws.RemoteRegion

	roleArn := fmt.Sprintf("arn:aws:iam::%s:role/%s", remoteAccountId, state.roleName)

	logger.WithValues(
		"awsRegion", state.remoteRegion,
		"awsRole", roleArn,
	).Info("Assuming AWS role")

	client, err := state.provider(
		ctx,
		state.remoteRegion,
		state.awsAccessKeyid,
		state.awsSecretAccessKey,
		roleArn,
	)

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error initializing remote AWS client", composed.StopWithRequeue, ctx)
	}

	vpcList, err := client.DescribeVpcs(ctx)

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error loading AWS VPC Networks", composed.StopWithRequeue, ctx)
	}

	var vpc *ec2Types.Vpc
	var remoteVpcName string
	var allLoadedVpcs []string
	// TODO refactor as it is more or less the same as in loadVpc

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

		if pointer.StringDeref(v.VpcId, "xxx") == remoteVpcId {
			remoteVpcName = util.GetEc2TagValue(v.Tags, "Name")
			vpc = &v
		}
	}

	if vpc == nil {
		logger.
			WithValues(
				"remoteVpcId", remoteVpcId,
				"allLoadedVpcs", fmt.Sprintf("%v", allLoadedVpcs),
			).
			Info("VPC not found")

		return composed.UpdateStatus(state.ObjAsVpcPeering()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonVpcNotFound,
				Message: fmt.Sprintf("AWS VPC ID %s not found", remoteVpcId),
			}).
			ErrorLogMessage("Error updating VpcPeering status when loading vpc").
			SuccessError(composed.StopAndForget).
			Run(ctx, st)
	}

	state.remoteVpc = vpc

	logger = logger.WithValues(
		"remoteVpcId", pointer.StringDeref(state.remoteVpc.VpcId, ""),
		"remoteVpcName", remoteVpcName,
	)
	ctx = composed.LoggerIntoCtx(ctx, logger)

	return nil, ctx
}
