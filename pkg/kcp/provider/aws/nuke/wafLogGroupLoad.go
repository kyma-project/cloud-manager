package nuke

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	cloudwatchlogstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	nuketypes "github.com/kyma-project/cloud-manager/pkg/kcp/nuke/types"
	awswafpolicyutil "github.com/kyma-project/cloud-manager/pkg/skr/provider/aws/wafpolicy/util"
	"k8s.io/utils/ptr"
)

type WafLogGroupResource struct {
	logGroup cloudwatchlogstypes.LogGroup
}

func (r WafLogGroupResource) GetId() string {
	return ptr.Deref(r.logGroup.LogGroupName, "")
}

func (r WafLogGroupResource) GetObject() any {
	return r
}

func wafLogGroupLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	expectedLogGroupName := awswafpolicyutil.WafLogGroupName(state.Scope().Spec.ShootName)

	paginator := cloudwatchlogs.NewDescribeLogGroupsPaginator(state.awsClient, &cloudwatchlogs.DescribeLogGroupsInput{
		LogGroupNamePrefix: new(expectedLogGroupName),
	})

	rks := &nuketypes.ProviderResourceKindState{
		Kind:     "WafLogGroup",
		Provider: cloudcontrolv1beta1.ProviderAws,
		Objects:  []nuketypes.ProviderResourceObject{},
	}

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			logger.Error(err, "Error listing LogGroups for WAF")
			state.hadErrors = true
			return nil, ctx
		}
		for _, lg := range page.LogGroups {
			if ptr.Deref(lg.LogGroupName, "") == expectedLogGroupName {
				rks.Objects = append(rks.Objects, WafLogGroupResource{logGroup: lg})
				break
			}
		}
	}

	state.ProviderResources = append(state.ProviderResources, rks)

	return nil, ctx
}
