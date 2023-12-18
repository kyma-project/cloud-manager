package scope

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-resources/components/lib/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func loadGardenerCredentials(ctx context.Context, st composed.State) (error, context.Context) {
	logger := composed.LoggerFromCtx(ctx)
	state := st.(State)

	bindingName := *state.Shoot().Spec.SecretBindingName

	secretBinding, err := state.GardenerClient().SecretBindings(state.ShootNamespace()).Get(ctx, bindingName, metav1.GetOptions{})
	if err != nil {
		err = fmt.Errorf("error getting shoot secret binding %s: %w", bindingName, err)
		logger.Error(err, "Error")
		return err, nil
	}

	state.SetProvider(ProviderType(secretBinding.Provider.Type))

	secret, err := state.GardenK8sClient().CoreV1().Secrets(secretBinding.SecretRef.Namespace).
		Get(ctx, secretBinding.SecretRef.Name, metav1.GetOptions{})
	if err != nil {
		err = fmt.Errorf("error getting shoot related secret: %w", err)
		logger.Error(err, "Error")
		return err, nil
	}

	for k, v := range secret.Data {
		state.CredentialData()[k] = string(v)
	}

	logger.Info("Garden credential loaded")

	return nil, nil
}
