package nfsinstance

import (
	"context"
	"fmt"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/utils/pointer"
)

func findSecurityGroup(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	if state.securityGroup != nil {
		return nil, nil
	}

	logger := composed.LoggerFromCtx(ctx)

	list, err := state.awsClient.DescribeSecurityGroups(ctx, []ec2Types.Filter{
		{
			Name:   pointer.String("vpc-id"),
			Values: []string{state.IpRange().Status.VpcId},
		},
		{
			Name:   pointer.String("tag:Name"),
			Values: []string{state.Obj().GetName()},
		},
		{
			Name:   pointer.String(fmt.Sprintf("tag:%s", common.TagCloudManagerName)),
			Values: []string{state.Name().String()},
		},
	}, nil)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error listing security groups", composed.StopWithRequeue, nil)
	}

	if len(list) > 0 {
		state.securityGroup = &list[0]
		state.securityGroupId = pointer.StringDeref(state.securityGroup.GroupId, "")
		logger = logger.WithValues("securityGroupId", state.securityGroupId)
		logger.Info("NFS security group found")
		return nil, composed.LoggerIntoCtx(ctx, logger)
	}

	logger.Info("Security group not found")

	return nil, nil
}
