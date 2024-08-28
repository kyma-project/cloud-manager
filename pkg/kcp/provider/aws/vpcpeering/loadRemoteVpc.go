package vpcpeering

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
)

func loadRemoteVpc(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	obj := state.ObjAsVpcPeering()

	logger := composed.LoggerFromCtx(ctx)
	remoteVpcId := obj.Spec.VpcPeering.Aws.RemoteVpcId
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
		return awsmeta.LogErrorAndReturn(err, "Error initializing remote AWS client", ctx)
	}

	vpcList, err := client.DescribeVpcs(ctx)

	if err != nil {
		return awsmeta.LogErrorAndReturn(err, "Error loading AWS VPC Networks", ctx)
	}

	var vpc *ec2Types.Vpc
	var remoteVpcName string
	var allLoadedVpcs []string
	// TODO refactor as it is more or less the same as in loadVpc

	for _, vv := range vpcList {
		v := vv

		allLoadedVpcs = append(allLoadedVpcs, fmt.Sprintf(
			"%s{%s}",
			ptr.Deref(v.VpcId, ""),
			util.TagsToString(v.Tags),
		))

		if ptr.Deref(v.VpcId, "xxx") == remoteVpcId {
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

		condition := metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  metav1.ConditionTrue,
			Reason:  cloudcontrolv1beta1.ReasonVpcNotFound,
			Message: fmt.Sprintf("AWS VPC ID %s not found", remoteVpcId),
		}

		if !awsmeta.AnyConditionChanged(obj, condition) {
			return composed.StopAndForget, nil
		}

		return composed.UpdateStatus(obj).
			SetExclusiveConditions(condition).
			ErrorLogMessage("Error updating VpcPeering status when loading vpc").
			SuccessError(composed.StopAndForget).
			Run(ctx, st)
	}

	state.remoteVpc = vpc

	ctx = composed.LoggerIntoCtx(ctx, logger.WithValues(
		"remoteVpcId", remoteVpcId,
		"remoteVpcName", remoteVpcName,
	))

	return nil, ctx
}
