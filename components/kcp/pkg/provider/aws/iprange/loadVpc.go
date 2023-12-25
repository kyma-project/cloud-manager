package iprange

import (
	"context"
	"fmt"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-resources/components/kcp/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-resources/components/lib/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func loadVpc(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	vpcNetworkName := state.Scope().Spec.Scope.Aws.VpcNetwork
	vpcList, err := state.networkClient.DescribeVpcs(ctx)
	if err != nil {
		return err, nil
	}

	var vpc *ec2Types.Vpc
	for _, v := range vpcList {
		if nameEquals(v.Tags, vpcNetworkName) {
			vpc = &v
			break
		}
	}
	if vpc == nil {
		logger.WithValues("vpcName", vpcNetworkName).Info("VPC not found")

		meta.SetStatusCondition(state.IpRange().Conditions(), metav1.Condition{
			Type:    cloudresourcesv1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  cloudresourcesv1beta1.ReasonVpcNotFound,
			Message: fmt.Sprintf("AWS VPC %s not found", vpcNetworkName),
		})

		err := state.UpdateObjStatus(ctx)
		if err != nil {
			logger.Error(err, "Error updating IpRange status when loading vpc")
			return composed.StopWithRequeue, nil
		}

		return composed.StopAndForget, nil
	}

	state.vpc = vpc

	return nil, nil
}
