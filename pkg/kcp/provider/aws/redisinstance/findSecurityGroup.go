package redisinstance

import (
	"context"
	"fmt"

	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"k8s.io/utils/ptr"
)

func findSecurityGroup(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	if state.securityGroup != nil {
		return nil, nil
	}

	logger := composed.LoggerFromCtx(ctx)

	list, err := state.awsClient.DescribeElastiCacheSecurityGroups(ctx, []ec2Types.Filter{
		{
			Name:   ptr.To("vpc-id"),
			Values: []string{state.IpRange().Status.VpcId},
		},
		{
			Name:   ptr.To("tag:Name"),
			Values: []string{GetAwsElastiCacheSecurityGroupName(state.Obj().GetName())},
		},
		{
			Name:   ptr.To(fmt.Sprintf("tag:%s", common.TagCloudManagerName)),
			Values: []string{state.Name().String()},
		},
	}, nil)
	if err != nil {
		return awsmeta.LogErrorAndReturn(err, "Error listing security groups", ctx)
	}

	if len(list) > 0 {
		state.securityGroup = &list[0]
		state.securityGroupId = ptr.Deref(state.securityGroup.GroupId, "")
		logger = logger.WithValues("securityGroupId", state.securityGroupId)
		logger.Info("Redis security group found")
		return nil, composed.LoggerIntoCtx(ctx, logger)
	}

	logger.Info("Security group not found")

	return nil, nil
}
