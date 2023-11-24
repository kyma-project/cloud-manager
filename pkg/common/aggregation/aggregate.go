package aggregation

import (
	"context"
	"errors"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-resources-manager/apis/cloud-resources/v1beta1"
	composed "github.com/kyma-project/cloud-resources-manager/pkg/common/composedAction"
	ctrl "sigs.k8s.io/controller-runtime"
)

func Aggregate(ctx context.Context, state1 composed.LoggableState) (*ctrl.Result, error) {
	state := state1.(*ReconcilingState)
	if state.CloudResources == nil {
		return nil, errors.New("unexpected empty CloudResources")
	}

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

func genericPredicate(state1 composed.LoggableState, kind string) bool {
	state := state1.(*ReconcilingState)
	return state.Obj.GetObjectKind().GroupVersionKind().GroupVersion().String() == cloudresourcesv1beta1.GroupVersion.String() &&
		state.Obj.GetObjectKind().GroupVersionKind().Kind == kind
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

	// it has not been aggregated, maybe log something ???
	return
}
