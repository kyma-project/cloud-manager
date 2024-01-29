package focal

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func AwsProviderPredicate(_ context.Context, st composed.State) bool {
	state := st.(State)
	return state.Scope().Spec.Provider == cloudresourcesv1beta1.ProviderAws
}

func AzureProviderPredicate(_ context.Context, st composed.State) bool {
	state := st.(State)
	return state.Scope().Spec.Provider == cloudresourcesv1beta1.ProviderAzure
}

func GcpProviderPredicate(_ context.Context, st composed.State) bool {
	state := st.(State)
	return state.Scope().Spec.Provider == cloudresourcesv1beta1.ProviderGCP
}
