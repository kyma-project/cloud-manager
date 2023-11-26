package genericActions

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-resources-manager/apis/cloud-resources/v1beta1"
	composed "github.com/kyma-project/cloud-resources-manager/pkg/common/composedAction"
)

type caseNfsVolume struct{}

func (c *caseNfsVolume) Predicate(_ context.Context, state composed.State) bool {
	return genericPredicate(state, "NfsVolume")
}

func (c *caseNfsVolume) Action(_ context.Context, state composed.State) error {
	obj := state.Obj().(*cloudresourcesv1beta1.NfsVolume)
	genericAggregateAction(obj, &NfsVolumeInfoListWrap{
		NfsVolumeInfoList: &state.(StateWithCloudResources).ServedCloudResources().Spec.Aggregations.NfsVolumes,
	})
	return nil
}
