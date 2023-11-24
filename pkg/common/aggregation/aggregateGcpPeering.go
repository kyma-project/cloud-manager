package aggregation

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-resources-manager/apis/cloud-resources/v1beta1"
	composed "github.com/kyma-project/cloud-resources-manager/pkg/common/composedAction"
	ctrl "sigs.k8s.io/controller-runtime"
)

type caseGcpPeerings struct{}

func (c *caseGcpPeerings) Predicate(_ context.Context, state composed.LoggableState) bool {
	return genericPredicate(state, "GcpVpcPeering")
}

func (c *caseGcpPeerings) Action(ctx context.Context, state1 composed.LoggableState) (*ctrl.Result, error) {
	state := state1.(*ReconcilingState)
	obj := state.Obj.(*cloudresourcesv1beta1.GcpVpcPeering)
	genericAggregateAction(obj, &GcpVpcPeeringInfoListWrap{
		GcpVpcPeeringInfoList: &state.CloudResources.Spec.Aggregations.GcpVpcPeerings,
	})
	return nil, nil
}
