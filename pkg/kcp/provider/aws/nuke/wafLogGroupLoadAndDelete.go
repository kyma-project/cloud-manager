package nuke

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	cloudwatchlogstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/kyma-project/cloud-manager/pkg/common/rate"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awswafpolicyutil "github.com/kyma-project/cloud-manager/pkg/skr/provider/aws/wafpolicy/util"
	"k8s.io/utils/ptr"
)

func wafLogGroupLoadAndDelete(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	expectedLogGroupName := awswafpolicyutil.WafLogGroupName(state.Scope().Spec.ShootName)

	paginator := cloudwatchlogs.NewDescribeLogGroupsPaginator(state.awsClient, &cloudwatchlogs.DescribeLogGroupsInput{
		LogGroupNamePrefix: new(expectedLogGroupName),
	})

	var theLogGroup *cloudwatchlogstypes.LogGroup
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err, "Error listing LogGroups for WAF", composed.StopWithRequeueDelay(rate.Default.Slow1s.When(state.Obj())), ctx)
		}
		for _, lg := range page.LogGroups {
			if ptr.Deref(lg.LogGroupName, "") == expectedLogGroupName {
				theLogGroup = &lg
				break
			}
		}
	}

	if theLogGroup == nil {
		return nil, ctx
	}

	_, err := state.awsClient.DeleteLogGroup(ctx, &cloudwatchlogs.DeleteLogGroupInput{
		LogGroupName: new(expectedLogGroupName),
	})
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error deleting WAF LogGroup", composed.StopWithRequeueDelay(rate.Default.Slow1s.When(state.Obj())), ctx)
	}

	return nil, ctx
}
