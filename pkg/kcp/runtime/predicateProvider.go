package runtime

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/feature"
)

func awsProviderPredicate(ctx context.Context, st composed.State) bool {
	if !feature.RuntimeSecurityAws.Value(ctx) {
		return false
	}
	state := st.(*State)
	return state.ObjAsRuntime().Spec.Shoot.Provider.Type == string(cloudcontrolv1beta1.ProviderAws)
}

func azureProviderPredicate(ctx context.Context, st composed.State) bool {
	if !feature.RuntimeSecurityAzure.Value(ctx) {
		return false
	}
	state := st.(*State)
	return state.ObjAsRuntime().Spec.Shoot.Provider.Type == string(cloudcontrolv1beta1.ProviderAzure)
}

func gcpProviderPredicate(ctx context.Context, st composed.State) bool {
	if !feature.RuntimeSecurityGcp.Value(ctx) {
		return false
	}
	state := st.(*State)
	return state.ObjAsRuntime().Spec.Shoot.Provider.Type == string(cloudcontrolv1beta1.ProviderGCP)
}
