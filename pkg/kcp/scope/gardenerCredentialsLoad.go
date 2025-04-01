package scope

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func gardenerCredentialsLoad(ctx context.Context, st composed.State) (error, context.Context) {
	logger := composed.LoggerFromCtx(ctx)
	state := st.(*State)

	bindingName := *state.shoot.Spec.SecretBindingName

	secretBinding, err := state.gardenerClient.SecretBindings(state.shootNamespace).Get(ctx, bindingName, metav1.GetOptions{})
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error getting shoot secret binding", composed.StopWithRequeue, ctx)
	}

	state.provider = cloudcontrolv1beta1.ProviderType(secretBinding.Provider.Type)

	secret, err := state.gardenK8sClient.CoreV1().Secrets(secretBinding.SecretRef.Namespace).
		Get(ctx, secretBinding.SecretRef.Name, metav1.GetOptions{})
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error getting shoot secret", composed.StopWithRequeue, ctx)
	}

	for k, v := range secret.Data {
		state.credentialData[k] = string(v)
	}

	logger.Info("Garden credential loaded")

	return nil, ctx
}
