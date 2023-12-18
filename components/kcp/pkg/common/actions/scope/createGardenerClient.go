package scope

import (
	"context"
	"fmt"
	gardenerClient "github.com/gardener/gardener/pkg/client/core/clientset/versioned/typed/core/v1beta1"
	"github.com/kyma-project/cloud-resources/components/lib/composed"
	kubernetesClient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"os"
)

func createGardenerClient(ctx context.Context, st composed.State) (error, context.Context) {
	logger := composed.LoggerFromCtx(ctx)
	state := st.(State)
	fn := os.Getenv("GARDENER_CREDENTIALS")
	if len(fn) == 0 {
		fn = "/opt/cloud-resources/gardener-credentials/kubeconfig"
	}

	logger = logger.WithValues("credentialsPath", fn)
	logger.Info("Loading gardener credentials")
	kubeBytes, err := state.FileReader().ReadFile(fn)
	if err != nil {
		err = fmt.Errorf("error loading gardener credentials: %w", err)
		logger.Error(err, "error creating gardener client")
		return composed.StopAndForget, nil // no requeue
	}

	config, err := clientcmd.NewClientConfigFromBytes(kubeBytes)
	if err != nil {
		return fmt.Errorf("error creating gardener client config: %w", err), nil
	}

	rawConfig, err := config.RawConfig()
	if err != nil {
		return fmt.Errorf("error getting gardener raw client config: %w", err), nil
	}
	var configContext *clientcmdapi.Context
	if len(rawConfig.CurrentContext) > 0 {
		configContext = rawConfig.Contexts[rawConfig.CurrentContext]
	} else {
		for _, c := range rawConfig.Contexts {
			configContext = c
			break
		}
	}
	if configContext != nil && len(configContext.Namespace) > 0 {
		state.SetShootNamespace(configContext.Namespace)
	} else {
		state.SetShootNamespace(os.Getenv("GARDENER_NAMESPACE"))
	}

	logger = logger.WithValues("shootProject", state.ShootNamespace())
	logger.Info("Detected shoot namespace")

	restConfig, err := clientcmd.RESTConfigFromKubeConfig(kubeBytes)
	if err != nil {
		err = fmt.Errorf("error creating gardener rest config: %w", err)
		logger.Error(err, "error creating gardener client")
		return composed.StopAndForget, nil // no requeue
	}

	gClient, err := gardenerClient.NewForConfig(restConfig)
	if err != nil {
		err = fmt.Errorf("error creating gardener client: %w", err)
		logger.Error(err, "error creating gardener client")
		return composed.StopAndForget, nil // no requeue
	}

	state.SetGardenerClient(gClient)

	k8sClient, err := kubernetesClient.NewForConfig(restConfig)
	if err != nil {
		err = fmt.Errorf("error creating gardene k8s client: %w", err)
		logger.Error(err, "error creating gardene k8s client")
		return composed.StopAndForget, nil // no requeue
	}

	state.SetGardenK8sClient(k8sClient)

	logger.Info("Gardener clients created")

	return nil, composed.LoggerIntoCtx(ctx, logger)
}
