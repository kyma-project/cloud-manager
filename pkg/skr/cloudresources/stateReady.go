package cloudresources

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func statusReady(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if composed.MarkedForDeletionPredicate(ctx, state) {
		return nil, nil
	}

	condReady := meta.FindStatusCondition(*state.ObjAsCloudResources().Conditions(), cloudresourcesv1beta1.ConditionTypeReady)
	if condReady != nil && state.ObjAsCloudResources().Status.State == cloudresourcesv1beta1.StateReady {
		return nil, nil
	}

	logger.Info("Setting Ready state and condition to CloudResources CR")
	state.ObjAsCloudResources().Status.State = cloudresourcesv1beta1.StateReady

	return composed.UpdateStatus(state.ObjAsCloudResources()).
		SetExclusiveConditions(metav1.Condition{
			Type:    cloudresourcesv1beta1.ConditionTypeReady,
			Status:  metav1.ConditionTrue,
			Reason:  cloudresourcesv1beta1.ConditionTypeReady,
			Message: "Ready",
		}).
		ErrorLogMessage("Error updating CloudResources CR with Ready condition").
		SuccessError(composed.StopAndForget). // do not continue further with the flow
		Run(ctx, state)
}
