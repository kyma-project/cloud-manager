package nfsinstance

import (
	"context"
	"fmt"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/common"
	"github.com/kyma-project/cloud-manager/components/lib/composed"
	"k8s.io/utils/pointer"
)

func createSecurityGroup(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	if state.securityGroup != nil {
		return nil, nil
	}

	logger := composed.LoggerFromCtx(ctx)

	name := fmt.Sprintf("cr--%s", state.ObjAsNfsInstance().Spec.RemoteRef.String())
	sgId, err := state.awsClient.CreateSecurityGroup(ctx, state.IpRange().Status.VpcId, name, []ec2Types.Tag{
		{
			Key:   pointer.String(common.TagCloudManagerName),
			Value: pointer.String(state.Name().String()),
		},
	})
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error creating security group", composed.StopWithRequeue, nil)
	}

	state.securityGroupId = sgId

	logger = logger.WithValues("securityGroupId", sgId)
	logger.Info("Security group created")

	return nil, composed.LoggerIntoCtx(ctx, logger)
}
