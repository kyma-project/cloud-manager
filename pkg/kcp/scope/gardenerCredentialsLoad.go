package scope

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	commongardener "github.com/kyma-project/cloud-manager/pkg/common/gardener"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/utils/ptr"
)

func gardenerCredentialsLoad(ctx context.Context, st composed.State) (error, context.Context) {
	logger := composed.LoggerFromCtx(ctx)
	state := st.(*State)

	out, err := commongardener.LoadGardenerCloudProviderCredentials(ctx, commongardener.LoadGardenerCloudProviderCredentialsInput{
		GardenerClient:  state.gardenerClient,
		GardenK8sClient: state.gardenK8sClient,
		Namespace:       state.shootNamespace,
		BindingName:     ptr.Deref(state.shoot.Spec.SecretBindingName, ""),
	})
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error loading gardener cloud credentials", composed.StopWithRequeue, ctx)
	}

	state.provider = cloudcontrolv1beta1.ProviderType(out.Provider)

	for k, v := range out.CredentialsData {
		state.credentialData[k] = string(v)
	}

	logger.Info("Garden credential loaded")

	return nil, ctx
}
