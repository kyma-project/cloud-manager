package reconcile

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-resources-manager/apis/cloud-resources/v1beta1"
	composed "github.com/kyma-project/cloud-resources-manager/pkg/common/composedAction"
	ctrl "sigs.k8s.io/controller-runtime"
)

type caseAwsPeerings struct{}

func (c *caseAwsPeerings) Predicate(_ context.Context, state composed.LoggableState) bool {
	return genericPredicate(state, "AwsVpcPeering")
}

func (c *caseAwsPeerings) Action(_ context.Context, state1 composed.LoggableState) (*ctrl.Result, error) {
	state := state1.(*ReconcilingState)
	obj := state.Obj.(*cloudresourcesv1beta1.AwsVpcPeering)
	genericAggregateAction(obj, &AwsVpcPeeringInfoListWrap{
		AwsVpcPeeringInfoList: &state.CloudResources.Spec.Aggregations.AwsVpcPeerings,
	})
	return nil, nil
}
