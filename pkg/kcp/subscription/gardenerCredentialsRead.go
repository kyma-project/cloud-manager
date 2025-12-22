package subscription

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	commongardener "github.com/kyma-project/cloud-manager/pkg/common/gardener"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func gardenerCredentialsRead(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	out, err := commongardener.LoadGardenerCloudProviderCredentials(ctx, commongardener.LoadGardenerCloudProviderCredentialsInput{
		Client:      state.gardenerClient,
		Namespace:   state.gardenNamespace,
		BindingName: state.ObjAsSubscription().Spec.Details.Garden.BindingName,
	})
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error reading garden credentials for Subscription", composed.StopWithRequeue, ctx)
	}

	state.provider = cloudcontrolv1beta1.ProviderType(out.Provider)
	state.credentialData = out.CredentialsData

	return nil, ctx
}
