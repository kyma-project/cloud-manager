package runtime

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func awsProviderPredicate(_ context.Context, st composed.State) bool {
	state := st.(*State)
	return state.ObjAsRuntime().Spec.Shoot.Provider.Type == string(cloudcontrolv1beta1.ProviderAws)
}

func azureProviderPredicate(_ context.Context, st composed.State) bool {
	state := st.(*State)
	return state.ObjAsRuntime().Spec.Shoot.Provider.Type == string(cloudcontrolv1beta1.ProviderAzure)
}

func gcpProviderPredicate(_ context.Context, st composed.State) bool {
	state := st.(*State)
	return state.ObjAsRuntime().Spec.Shoot.Provider.Type == string(cloudcontrolv1beta1.ProviderGCP)
}
