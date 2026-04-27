package runtime

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func awsProviderPredicate(_ context.Context, st composed.State) bool {
	state := st.(*State)
	return state.ObjAsRuntime().Spec.Shoot.Provider.Type == "aws"
}

func azureProviderPredicate(_ context.Context, st composed.State) bool {
	state := st.(*State)
	return state.ObjAsRuntime().Spec.Shoot.Provider.Type == "azure"
}

func gcpProviderPredicate(_ context.Context, st composed.State) bool {
	state := st.(*State)
	return state.ObjAsRuntime().Spec.Shoot.Provider.Type == "gcp"
}
