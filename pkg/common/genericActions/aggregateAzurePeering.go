package genericActions

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-resources-manager/apis/cloud-resources/v1beta1"
	composed "github.com/kyma-project/cloud-resources-manager/pkg/common/composedAction"
)

type caseAzurePeerings struct{}

func (c *caseAzurePeerings) Predicate(_ context.Context, state composed.State) bool {
	return genericPredicate(state, "AzureVpcPeering")
}

func (c *caseAzurePeerings) Action(_ context.Context, state composed.State) error {
	obj := state.Obj().(*cloudresourcesv1beta1.AzureVpcPeering)
	genericAggregateAction(obj, &AzureVpcPeeringInfoListWrap{
		AzureVpcPeeringInfoList: &state.(StateWithCloudResources).ServedCloudResources().Spec.Aggregations.AzureVpcPeerings,
	})
	return nil
}
