package reconcile

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-resources-manager/apis/cloud-resources/v1beta1"
	composed "github.com/kyma-project/cloud-resources-manager/pkg/common/composedAction"
	ctrl "sigs.k8s.io/controller-runtime"
)

type caseAzurePeerings struct{}

func (c *caseAzurePeerings) Predicate(_ context.Context, state composed.LoggableState) bool {
	return genericPredicate(state, "AzureVpcPeering")
}

func (c *caseAzurePeerings) Action(ctx context.Context, state1 composed.LoggableState) (*ctrl.Result, error) {
	state := state1.(*ReconcilingState)
	obj := state.Obj.(*cloudresourcesv1beta1.AzureVpcPeering)
	genericAggregateAction(obj, &AzureVpcPeeringInfoListWrap{
		AzureVpcPeeringInfoList: &state.CloudResources.Spec.Aggregations.AzureVpcPeerings,
	})
	return nil, nil
}
