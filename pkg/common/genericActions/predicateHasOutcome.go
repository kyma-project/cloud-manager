package genericActions

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-resources-manager/apis/cloud-resources/v1beta1"
	composed "github.com/kyma-project/cloud-resources-manager/pkg/common/composedAction"
)

func HasOutcome(outcomeType cloudresourcesv1beta1.OutcomeType) composed.Predicate {
	return func(ctx context.Context, state composed.State) bool {
		outcome := state.Obj().(Aggregable).GetOutcome()
		return outcome != nil && outcome.Type == outcomeType
	}
}
