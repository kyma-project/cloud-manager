package nuke

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"k8s.io/utils/ptr"
)

func wafLogGroupDelete(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	for _, rks := range state.ProviderResources {
		if rks.Kind == "WafLogGroup" && rks.Provider == cloudcontrolv1beta1.ProviderAws {
			for _, obj := range rks.Objects {
				resource := obj.(WafLogGroupResource)
				logGroupName := ptr.Deref(resource.logGroup.LogGroupName, "")

				_, err := state.awsClient.DeleteLogGroup(ctx, &cloudwatchlogs.DeleteLogGroupInput{
					LogGroupName: new(logGroupName),
				})
				if err != nil {
					if awsmeta.IsNotFound(err) {
						logger.Info("WAF LogGroup already deleted", "logGroupName", logGroupName)
						continue
					}
					logger.Error(err, "Error deleting WAF LogGroup", "logGroupName", logGroupName)
					state.hadErrors = true
				} else {
					logger.Info("Deleted WAF LogGroup", "logGroupName", logGroupName)
				}
			}
		}
	}

	return nil, ctx
}
