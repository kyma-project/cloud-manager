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

func loadRemoteVpc(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	obj := state.ObjAsVpcPeering()

	remoteVpcId := obj.Spec.VpcPeering.Aws.RemoteVpcId

	vpcList, err := state.remoteClient.DescribeVpcs(ctx)

	if err != nil {
		if awsmeta.IsErrorRetryable(err) {
			return awsmeta.LogErrorAndReturn(err, "Error loading remote AWS VPC Networks", ctx)
		}

		logger.Error(err, "Error loading remote AWS VPC Networks")

		condition := metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  metav1.ConditionTrue,
			Reason:  cloudcontrolv1beta1.ReasonVpcNotFound,
			Message: err.Error(),
		}

		if !composed.AnyConditionChanged(obj, condition) {
			return composed.StopAndForget, nil
		}

		return composed.UpdateStatus(obj).
			SetExclusiveConditions(condition).
			ErrorLogMessage("Error updating VpcPeering status when loading vpc").
			SuccessError(composed.StopAndForget).
			Run(ctx, st)
	}

	var vpc *ec2Types.Vpc
	var remoteVpcName string

	for _, vv := range vpcList {
		v := vv
		if ptr.Deref(v.VpcId, "xxx") == remoteVpcId {
			remoteVpcName = util.GetEc2TagValue(v.Tags, "Name")
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

		if !composed.AnyConditionChanged(obj, condition) {
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
