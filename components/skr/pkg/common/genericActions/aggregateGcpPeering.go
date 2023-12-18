package genericActions

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-resources/components/skr/api/cloud-resources/v1beta1"
	composed "github.com/kyma-project/cloud-resources/components/skr/pkg/common/composedAction"
)

type caseGcpPeerings struct{}

func (c *caseGcpPeerings) Predicate(_ context.Context, state composed.State) bool {
	return genericPredicate(state, "GcpVpcPeering")
}

func (c *caseGcpPeerings) Action(_ context.Context, state composed.State) error {
	obj := state.Obj().(*cloudresourcesv1beta1.GcpVpcPeering)
	genericAggregateAction(obj, &GcpVpcPeeringInfoListWrap{
		GcpVpcPeeringInfoList: &state.(StateWithCloudResources).ServedCloudResources().Spec.Aggregations.GcpVpcPeerings,
	})
	return nil
}
