package aggregation

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-resources-manager/apis/cloud-resources/v1beta1"
	composed "github.com/kyma-project/cloud-resources-manager/pkg/common/composedAction"
	ctrl "sigs.k8s.io/controller-runtime"
)

type caseNfsVolume struct{}

func (c *caseNfsVolume) Predicate(_ context.Context, state composed.LoggableState) bool {
	return genericPredicate(state, "NfsVolume")
}

func (c *caseNfsVolume) Action(ctx context.Context, state1 composed.LoggableState) (*ctrl.Result, error) {
	state := state1.(*ReconcilingState)
	obj := state.Obj.(*cloudresourcesv1beta1.NfsVolume)
	genericAggregateAction(obj, &NfsVolumeInfoListWrap{
		NfsVolumeInfoList: &state.CloudResources.Spec.Aggregations.NfsVolumes,
	})
	return nil, nil
}
