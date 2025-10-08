package e2e

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/kyma-project/cloud-manager/config/crd"
	e2econfig "github.com/kyma-project/cloud-manager/e2e/config"
	"github.com/kyma-project/cloud-manager/e2e/sim"
	"github.com/kyma-project/cloud-manager/pkg/util"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type WorldFactory struct {
}

func NewWorldFactory() *WorldFactory {
	return &WorldFactory{}
}

func (f *WorldFactory) Create(ctx context.Context) (World, error) {
	clusterProvider := NewClusterProvider()
	kcp, err := clusterProvider.KCP(ctx)
	if err != nil {
		return nil, fmt.Errorf("error creating kcp cluster: %w", err)
	}

	err = f.installCrds(ctx, kcp)
	if err != nil {
		return nil, fmt.Errorf("error installing CRDs in kcp: %w", err)
	}

	err = f.initGardenerCredentials(ctx, kcp)
	if err != nil {
		return nil, fmt.Errorf("error initializing gardener credentials in kcp: %w", err)
	}

	garden, err := clusterProvider.Garden(ctx)
	if err != nil {
		return nil, fmt.Errorf("error creating garden cluster: %w", err)
	}

	simu, err := sim.New(ctx, sim.CreateOptions{
		KCP:    kcp,
		Garden: garden,
		Logger: ctrl.Log.WithName("sim"),
		KubeconfigProvider: sim.NewKubeconfigProvider(garden.GetClient(), 10*time.Hour),
	})
	if err != nil {
		return nil, fmt.Errorf("error creating sim: %w", err)
	}

	go func() {

	}()

	return &defaultWorld{
		clusterProvider: clusterProvider,
	}, nil}
}

func (f *WorldFactory) setGardenNamespaceInConfig(gardenKubeBytes []byte) error {
	config, err := clientcmd.NewClientConfigFromBytes(gardenKubeBytes)
	if err != nil {
		return fmt.Errorf("error creating gardener client config: %w", err)
	}

	rawConfig, err := config.RawConfig()
	if err != nil {
		return fmt.Errorf("error getting gardener raw client config: %w", err)
	}

	if len(rawConfig.CurrentContext) > 0 {
		e2econfig.Config.GardenNamespace = rawConfig.Contexts[rawConfig.CurrentContext].Namespace
	}

	return nil
}

func (f *WorldFactory) initGardenerCredentials(ctx context.Context, kcp Cluster) error {
	if e2econfig.Config.GardenKubeconfig == "" {
		return fmt.Errorf("garden kubeconfig is not set in config")
	}

	kubeBytes, err := os.ReadFile(e2econfig.Config.GardenKubeconfig)
	if err != nil {
		return fmt.Errorf("failed to read garden kubeconfig from %q: %w", e2econfig.Config.GardenKubeconfig, err)
	}

	err = f.setGardenNamespaceInConfig(kubeBytes)
	if err != nil {
		return fmt.Errorf("failed to set garden namespace to config from gardener credentials: %w", err)
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: e2econfig.Config.KcpNamespace,
			Name:      "gardener-credentials",
		},
	}

	err = kcp.GetAPIReader().Get(ctx, client.ObjectKeyFromObject(secret), secret)
	if err == nil {
		// already exists
		err = kcp.GetClient().Delete(ctx, secret)
		if err != nil {
			return fmt.Errorf("failed to delete existing gardener credentials: %w", err)
		}
	}

	secret = &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: e2econfig.Config.KcpNamespace,
			Name:      "gardener-credentials",
		},
		Data: map[string][]byte{
			"kubeconfig": kubeBytes,
		},
	}
	err = kcp.GetClient().Create(ctx, secret)
	if apierrors.IsAlreadyExists(err) {
		// some race condition, let's assume the secret correctly created
		return nil
	}
	if err != nil {
		return fmt.Errorf("error creating gardener credentials: %w", err)
	}

	return nil
}

func (f *WorldFactory) installCrds(ctx context.Context, kcp Cluster) error {
	arr, err := crd.KCP_All()
	if err != nil {
		return fmt.Errorf("error reading CRDs: %w", err)
	}
	err = util.Apply(ctx, kcp.GetClient(), arr)
	if err != nil {
		return fmt.Errorf("error installing CRDs: %w", err)
	}
	return nil
}
