package awsredisinstance

import (
	"context"
	"fmt"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func deleteAuthSecret(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.AuthSecret == nil {
		return nil, nil
	}

	if !state.AuthSecret.DeletionTimestamp.IsZero() {
		return nil, nil
	}

	err, _ := composed.PatchStatus(state.ObjAsAwsRedisInstance()).
		SetExclusiveConditions(metav1.Condition{
			Type:    cloudresourcesv1beta1.ConditionTypeDeleting,
			Status:  metav1.ConditionTrue,
			Reason:  cloudresourcesv1beta1.ConditionReasonDeletingAuthSecret,
			Message: fmt.Sprintf("Deleting Auth Secret %s", state.AuthSecret.Name),
		}).
		ErrorLogMessage("Error setting ConditionReasonDeletingAuthSecret condition on AwsRedisInstance").
		SuccessErrorNil().
		FailedError(composed.StopWithRequeue).
		Run(ctx, state)
	if err != nil {
		return err, nil
	}

	logger.Info("Deleting AuthSecret for AwsRedisInstance")

	err = state.Cluster().K8sClient().Delete(ctx, state.AuthSecret)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error deleting AuthSecret for AwsRedisInstance", composed.StopWithRequeue, ctx)
	}

	return composed.StopWithRequeue, nil
}
