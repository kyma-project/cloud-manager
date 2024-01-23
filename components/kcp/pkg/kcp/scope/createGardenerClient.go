package scope

import (
	"context"
	gardenerClient "github.com/gardener/gardener/pkg/client/core/clientset/versioned/typed/core/v1beta1"
	"github.com/kyma-project/cloud-manager/components/lib/composed"
	kubernetesClient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"os"
)

func createGardenerClient(ctx context.Context, st composed.State) (error, context.Context) {
	logger := composed.LoggerFromCtx(ctx)
	state := st.(*State)
	fn := os.Getenv("GARDENER_CREDENTIALS")
	if len(fn) == 0 {
		fn = "/opt/cloud-manager/gardener-credentials/kubeconfig"
	}

	logger = logger.WithValues("credentialsPath", fn)
	logger.Info("Loading gardener credentials")
	kubeBytes, err := state.fileReader.ReadFile(fn)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error loading gardener client", composed.StopAndForget, nil)
	}

	config, err := clientcmd.NewClientConfigFromBytes(kubeBytes)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error creating gardener client config", composed.StopAndForget, nil)
	}

	rawConfig, err := config.RawConfig()
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error getting gardener raw client config", composed.StopAndForget, nil)
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
		state.shootNamespace = configContext.Namespace
	} else {
		state.shootNamespace = os.Getenv("GARDENER_NAMESPACE")
	}

	logger = logger.WithValues("shootProject", state.shootNamespace)
	logger.Info("Detected shoot namespace")

	restConfig, err := clientcmd.RESTConfigFromKubeConfig(kubeBytes)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error creating gardener rest config", composed.StopAndForget, nil)
	}

	gClient, err := gardenerClient.NewForConfig(restConfig)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error creating gardener client", composed.StopAndForget, nil)
	}

	state.gardenerClient = gClient

	k8sClient, err := kubernetesClient.NewForConfig(restConfig)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error creating garden k8s client", composed.StopAndForget, nil)
	}

	state.gardenK8sClient = k8sClient

	logger.Info("Gardener clients created")

	return nil, composed.LoggerIntoCtx(ctx, logger)
}
