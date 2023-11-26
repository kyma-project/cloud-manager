package genericActions

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-resources-manager/apis/cloud-resources/v1beta1"
	composed "github.com/kyma-project/cloud-resources-manager/pkg/common/composedAction"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func OutcomeDeletedToCondition() composed.Action {
	return composed.BuildBranchingAction(
		"OutcomeDeletedToCondition",
		HasOutcome(cloudresourcesv1beta1.OutcomeTypeDeleted),
		UpdateCondition(
			cloudresourcesv1beta1.ProcessingState,
			cloudresourcesv1beta1.ConditionTypeDeleted,
			cloudresourcesv1beta1.ConditionReasonProcessing,
			metav1.ConditionTrue,
			"Peering was deleted",
		),
		nil,
	)
}

func OutcomeErrorToCondition() composed.Action {
	return composed.BuildBranchingAction(
		"OutcomeErrorToCondition",
		HasOutcome(cloudresourcesv1beta1.OutcomeTypeError),
		func(ctx context.Context, state composed.State) error {
			return UpdateCondition(
				cloudresourcesv1beta1.ErrorState,
				cloudresourcesv1beta1.ConditionTypeError,
				cloudresourcesv1beta1.ConditionReasonError,
				metav1.ConditionTrue,
				state.Obj().(Aggregable).GetOutcome().Message,
			)(ctx, state)
		},
		nil,
	)
}

func OutcomeCreatedToCondition() composed.Action {
	return composed.BuildBranchingAction(
		"",
		HasOutcome(cloudresourcesv1beta1.OutcomeTypeCreated),
		UpdateCondition(
			cloudresourcesv1beta1.ReadyState,
			cloudresourcesv1beta1.ConditionTypeReady,
			cloudresourcesv1beta1.ConditionReasonReady,
			metav1.ConditionTrue,
			"Resource provisioned",
		),
		nil,
	)
}
