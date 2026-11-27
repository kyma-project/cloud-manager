package cloudresources

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/rate"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awswafpolicyutil "github.com/kyma-project/cloud-manager/pkg/skr/provider/aws/wafpolicy/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func logGroupPutRetention(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if ptr.Deref(state.logGroup.RetentionInDays, 0) == state.RetentionDays() {
		return nil, ctx
	}

	_, err := state.awsClient.PutRetentionPolicy(ctx, &cloudwatchlogs.PutRetentionPolicyInput{
		LogGroupName:    new(awswafpolicyutil.WafLogGroupName(state.Scope().Spec.ShootName)),
		RetentionInDays: new(state.RetentionDays()),
	})
	if err != nil {
		return composed.NewStatusPatcherComposed(state.ObjAsCloudResources()).
			MutateStatus(func(cr *cloudresourcesv1beta1.CloudResources) {
				meta.SetStatusCondition(&cr.Status.Conditions, metav1.Condition{
					Type:               cloudresourcesv1beta1.ConditionTypeReady,
					Status:             metav1.ConditionFalse,
					ObservedGeneration: cr.GetGeneration(),
					Reason:             cloudresourcesv1beta1.ConditionReasonError,
					Message:            fmt.Sprintf("Failed to put log group retention for WafPolicy: %s", err.Error()),
				})
			}).
			OnSuccess(composed.RequeueAfter(rate.Default.Slow1s.When(st.Obj()))).
			OnFailure(
				composed.Log("Failed to patch CloudResources status with provider error while putting log group retention for WafPolicy subscription setup"),
				composed.RequeueAfter(rate.Default.Slow1s.When(st.Obj())),
			).
			OnStatusChanged(composed.Log("Error putting log group retention for waf policy")).
			Run(ctx, state.Cluster().K8sClient())
	}

	return nil, ctx
}
