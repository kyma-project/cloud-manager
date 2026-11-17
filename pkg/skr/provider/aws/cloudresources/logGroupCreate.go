package cloudresources

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/common/rate"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awswafpolicyutil "github.com/kyma-project/cloud-manager/pkg/skr/provider/aws/wafpolicy/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func logGroupCreate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.logGroup != nil {
		return nil, ctx
	}

	_, err := state.awsClient.CreateLogGroup(ctx, &cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: new(awswafpolicyutil.WafLogGroupName(state.Scope().Spec.ShootName)),
		Tags: map[string]string{
			common.TagCloudManagerName:       state.Name().String(),
			common.TagCloudManagerRemoteName: state.Name().String(),
			common.TagScope:                  state.Scope().Name,
			common.TagShoot:                  state.Scope().Spec.ShootName,
		},
	})
	if err != nil {
		return composed.NewStatusPatcherComposed(state.ObjAsCloudResources()).
			MutateStatus(func(cr *cloudresourcesv1beta1.CloudResources) {
				meta.SetStatusCondition(&cr.Status.Conditions, metav1.Condition{
					Type:               cloudresourcesv1beta1.ConditionTypeReady,
					Status:             metav1.ConditionFalse,
					ObservedGeneration: cr.GetGeneration(),
					Reason:             cloudresourcesv1beta1.ConditionReasonError,
					Message:            fmt.Sprintf("Failed to create log group for WafPolicy: %s", err.Error()),
				})
			}).
			OnSuccess(composed.RequeueAfter(rate.Default.Slow1s.When(st.Obj()))).
			OnFailure(
				composed.Log("Failed to patch CloudResources status with provider error while creating log group for WafPolicy subscription setup"),
				composed.RequeueAfter(rate.Default.Slow1s.When(st.Obj())),
			).
			OnStatusChanged(composed.Log("Error creating log group for waf policy")).
			Run(ctx, state.Cluster().K8sClient())
	}

	composed.LoggerFromCtx(ctx).Info("Successfully created logGroup for WafPolicy")

	return composed.StopWithRequeueDelay(time.Millisecond), ctx
}
