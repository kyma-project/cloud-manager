package scope

import (
	"context"
	gardenerclient "github.com/gardener/gardener/pkg/client/core/clientset/versioned/typed/core/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	kubernetesclient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func gardenerClientCreate(ctx context.Context, st composed.State) (error, context.Context) {
	logger := composed.LoggerFromCtx(ctx)
	state := st.(*State)

	logger.Info("Loading gardener credentials")

	secret := &corev1.Secret{}
	err := state.Cluster().ApiReader().Get(ctx, types.NamespacedName{
		Namespace: "kcp-system",
		Name:      "gardener-credentials",
	}, secret)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error getting gardener credentials", composed.StopWithRequeue, ctx)
	}

	kubeBytes, ok := secret.Data["kubeconfig"]
	if !ok {
		return composed.LogErrorAndReturn(err, "Gardener credentials missing kubeconfig", composed.StopAndForget, ctx)
	}

	config, err := clientcmd.NewClientConfigFromBytes(kubeBytes)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error creating gardener client config", composed.StopAndForget, ctx)
	}

	rawConfig, err := config.RawConfig()
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error getting gardener raw client config", composed.StopAndForget, ctx)
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
		state.shootNamespace = ScopeConfig.GardenerNamespace
	}

	logger = logger.WithValues("shootNamespace", state.shootNamespace)
	logger.Info("Detected shoot namespace")

	restConfig, err := clientcmd.RESTConfigFromKubeConfig(kubeBytes)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error creating gardener rest config", composed.StopAndForget, ctx)
	}

	gClient, err := gardenerclient.NewForConfig(restConfig)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error creating gardener client", composed.StopAndForget, ctx)
	}

	state.gardenerClient = gClient

	k8sClient, err := kubernetesclient.NewForConfig(restConfig)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error creating garden k8s client", composed.StopAndForget, ctx)
	}

	state.gardenK8sClient = k8sClient

	logger.Info("Gardener clients created")

	return nil, composed.LoggerIntoCtx(ctx, logger)
}
