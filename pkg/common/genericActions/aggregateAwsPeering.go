package genericActions

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-resources-manager/api/cloud-resources/v1beta1"
	composed "github.com/kyma-project/cloud-resources-manager/pkg/common/composedAction"
)

type caseAwsPeerings struct{}

func (c *caseAwsPeerings) Predicate(_ context.Context, state composed.State) bool {
	return genericPredicate(state, "AwsVpcPeering")
}

func (c *caseAwsPeerings) Action(_ context.Context, state composed.State) error {
	obj := state.Obj().(*cloudresourcesv1beta1.AwsVpcPeering)
	genericAggregateAction(obj, &AwsVpcPeeringInfoListWrap{
		AwsVpcPeeringInfoList: &state.(StateWithCloudResources).ServedCloudResources().Spec.Aggregations.AwsVpcPeerings,
	})
	return nil
}
