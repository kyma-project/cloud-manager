package genericActions

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-resources/components/skr/api/cloud-resources/v1beta1"
	composed "github.com/kyma-project/cloud-resources/components/skr/pkg/common/composedAction"
)

func Aggregate(ctx context.Context, state1 composed.State) error {
	return composed.BuildSwitchAction(
		"Aggregate",
		nil,
		&caseGcpPeerings{},
		&caseAzurePeerings{},
		&caseAwsPeerings{},
		&caseNfsVolume{},
	)(ctx, state1)
}

// ========================================================================

func genericPredicate(state composed.State, kind string) bool {
	return state.Obj().GetObjectKind().GroupVersionKind().GroupVersion().String() == cloudresourcesv1beta1.GroupVersion.String() &&
		state.Obj().GetObjectKind().GroupVersionKind().Kind == kind
}

func genericAggregateAction(obj Aggregable, list AggregateInfoList) {
	indexFound := -1
	objSourceRef := obj.GetSourceRef()
	for i, item := range list.All() {
		if item.GetSourceRef().Name == objSourceRef.Name {
			indexFound = i
			break
		}
	}

	if obj.GetDeletionTimestamp() == nil {
		// object being created or updated
		if indexFound > -1 {
			// update existing aggregation
			list.Get(indexFound).SetSpec(obj.GetSpec())
			return
		}

		// add new aggregation
		list.Append(objSourceRef, obj.GetSpec())
		return
	}

	// object is marked for deletion, should be removed from aggregate list
	if indexFound > -1 {
		list.Remove(indexFound)
		return
	}

	// TODO: it has not been aggregated, maybe log something ???

	return
}
