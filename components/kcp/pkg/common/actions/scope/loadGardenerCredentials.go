package scope

import (
	"context"
	"fmt"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-resources-manager/components/kcp/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-resources-manager/components/lib/composed"
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

	err = state.SetProvider(cloudresourcesv1beta1.ProviderType(secretBinding.Provider.Type))
	if err != nil {
		panic(err)
	}

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
