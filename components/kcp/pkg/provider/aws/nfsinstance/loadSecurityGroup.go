package nfsinstance

import (
	"context"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-resources/components/kcp/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-resources/components/lib/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
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
			Type:    cloudresourcesv1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  cloudresourcesv1beta1.ReasonUnknown,
			Message: "Unable to load security group",
		})
		state.ObjAsNfsInstance().Status.State = cloudresourcesv1beta1.ErrorState
		err := state.UpdateObjStatus(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err, "Error updating NfsInstance status after missing security group id", composed.StopWithRequeue, nil)
		}
		return composed.StopAndForget, nil
	}

	sg, err := state.awsClient.DescribeSecurityGroups(ctx, []ec2Types.Filter{
		{
			Name:   pointer.String("vpc-id"),
			Values: []string{state.IpRange().Status.VpcId},
		},
	}, []string{state.securityGroupId})
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error loading security group", composed.StopWithRequeue, nil)
	}
	if len(sg) < 1 {
		logger.Info("Security group with id not found!!!")
		meta.SetStatusCondition(state.ObjAsNfsInstance().Conditions(), metav1.Condition{
			Type:    cloudresourcesv1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  cloudresourcesv1beta1.ReasonUnknown,
			Message: "Unable to load security group",
		})
		state.ObjAsNfsInstance().Status.State = cloudresourcesv1beta1.ErrorState
		err := state.UpdateObjStatus(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err, "Error updating NfsInstance status after missing created security group", composed.StopWithRequeue, nil)
		}
		return composed.StopAndForget, nil
	}

	state.securityGroup = &sg[0]

	logger.Info("Created security group is loaded")

	return nil, nil
}
