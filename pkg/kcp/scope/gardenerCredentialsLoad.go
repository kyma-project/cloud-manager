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

	in := commongardener.LoadGardenerCloudProviderCredentialsInput{
		Client:    state.gardenerClient,
		Namespace: state.shootNamespace,
	}
	if x := ptr.Deref(state.shoot.Spec.CredentialsBindingName, ""); x != "" {
		in.BindingName = x
	}
	if in.BindingName == "" {
		//lint:ignore SA1019 we keep support for secretBinding until all landscapes are migrated
		x := ptr.Deref(state.shoot.Spec.SecretBindingName, "")
		in.BindingName = x
	}
	out, err := commongardener.LoadGardenerCloudProviderCredentials(ctx, in)
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
