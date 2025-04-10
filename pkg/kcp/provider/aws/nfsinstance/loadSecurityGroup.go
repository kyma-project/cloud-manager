package nfsinstance

import (
	"context"

	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func loadSecurityGroup(ctx context.Context, st composed.State) (error, context.Context) {
	logger := composed.LoggerFromCtx(ctx)
	state := st.(*State)
	if state.securityGroup != nil {
		return nil, nil
	}

	if len(state.securityGroupId) == 0 {
		logger.Info("Missing security group id!!!")
		meta.SetStatusCondition(state.ObjAsNfsInstance().Conditions(), metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  metav1.ConditionTrue,
			Reason:  cloudcontrolv1beta1.ReasonUnknown,
			Message: "Unable to load security group",
		})
		state.ObjAsNfsInstance().Status.State = cloudcontrolv1beta1.StateError
		err := state.UpdateObjStatus(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err, "Error updating NfsInstance status after missing security group id", composed.StopWithRequeue, ctx)
		}
		return composed.StopAndForget, nil
	}

	sg, err := state.awsClient.DescribeSecurityGroups(
		ctx,
		[]ec2types.Filter{
			{
				Name:   ptr.To("vpc-id"),
				Values: []string{state.IpRange().Status.VpcId},
			},
		},
		[]string{state.securityGroupId},
	)
	if err != nil {
		return awsmeta.LogErrorAndReturn(err, "Error loading security group", ctx)
	}
	if len(sg) < 1 {
		logger.Info("Security group with id not found!!!")
		meta.SetStatusCondition(state.ObjAsNfsInstance().Conditions(), metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  metav1.ConditionTrue,
			Reason:  cloudcontrolv1beta1.ReasonUnknown,
			Message: "Unable to load security group",
		})
		state.ObjAsNfsInstance().Status.State = cloudcontrolv1beta1.StateError
		err := state.UpdateObjStatus(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err, "Error updating NfsInstance status after missing created security group", composed.StopWithRequeue, ctx)
		}
		return composed.StopAndForget, nil
	}

	state.securityGroup = &sg[0]

	logger.Info("Created security group is loaded")

	return nil, nil
}
