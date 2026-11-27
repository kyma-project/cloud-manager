package cloudresources

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	cloudwatchlogstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/rate"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awswafpolicyutil "github.com/kyma-project/cloud-manager/pkg/skr/provider/aws/wafpolicy/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func logGroupLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	expectedLogGroupName := awswafpolicyutil.WafLogGroupName(state.Scope().Spec.ShootName)

	paginator := cloudwatchlogs.NewDescribeLogGroupsPaginator(state.awsClient, &cloudwatchlogs.DescribeLogGroupsInput{
		LogGroupNamePrefix: new(expectedLogGroupName),
	})

	var theLogGroup *cloudwatchlogstypes.LogGroup
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return composed.NewStatusPatcherComposed(state.ObjAsCloudResources()).
				MutateStatus(func(cr *cloudresourcesv1beta1.CloudResources) {
					meta.SetStatusCondition(&cr.Status.Conditions, metav1.Condition{
						Type:               cloudresourcesv1beta1.ConditionTypeReady,
						Status:             metav1.ConditionFalse,
						ObservedGeneration: cr.GetGeneration(),
						Reason:             cloudresourcesv1beta1.ConditionReasonError,
						Message:            fmt.Sprintf("Failed to list log group: %s", err.Error()),
					})
				}).
				OnSuccess(composed.RequeueAfter(rate.Default.Slow1s.When(st.Obj()))).
				OnFailure(
					composed.Log("Failed to patch CloudResources status with provider error while listing log groups for WafPolicy subscription setup"),
					composed.RequeueAfter(rate.Default.Slow1s.When(st.Obj())),
				).
				OnStatusChanged(composed.Log("Error listing log groups for waf policy")).
				Run(ctx, state.Cluster().K8sClient())
		}
		for _, lg := range page.LogGroups {
			if ptr.Deref(lg.LogGroupName, "") == expectedLogGroupName {
				theLogGroup = &lg
				break
			}
		}
	}

	state.logGroup = theLogGroup

	return nil, ctx
}
