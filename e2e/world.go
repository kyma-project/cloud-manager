package e2e

import (
	"context"
	"fmt"
	"maps"
	"os"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type World struct {
	ClusterProvider ClusterProvider
}

type worldKey struct{}

func setWorld(ctx context.Context, fc *World) context.Context {
	return context.WithValue(ctx, worldKey{}, fc)
}

func getWorld(ctx context.Context) *World {
	g, _ := ctx.Value(worldKey{}).(*World)

	return g
}

func NewWorld() *World {
	return &World{
		ClusterProvider: &defaultClusterProvider{},
	}
}

func (w *World) Init(ctx context.Context) error {
	if err := w.initGardenerCredentials(ctx); err != nil {
		return fmt.Errorf("failed to initialize gardener credentials: %w", err)
	}
	return nil
}

func (w *World) setGardenConfig(gardenKubeBytes []byte) error {
	config, err := clientcmd.NewClientConfigFromBytes(gardenKubeBytes)
	if err != nil {
		return fmt.Errorf("error creating gardener client config: %w", err)
	}

	rawConfig, err := config.RawConfig()
	if err != nil {
		return fmt.Errorf("error getting gardener raw client config: %w", err)
	}

	if len(rawConfig.CurrentContext) > 0 {
		Config.GardenNamespace = rawConfig.Contexts[rawConfig.CurrentContext].Namespace
	}

	return nil
}

func (w *World) initGardenerCredentials(ctx context.Context) error {
	kcp, err := w.ClusterProvider.KCP(ctx)
	if err != nil {
		return fmt.Errorf("failed to get KCP cluster: %w", err)
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: Config.KcpNamespace,
			Name:      "gardener-credentials",
		},
	}

	err = kcp.Cluster.GetAPIReader().Get(ctx, client.ObjectKeyFromObject(secret), secret)
	if err == nil {
		// already exists
		err = w.setGardenConfig(secret.Data["kubeconfig"])
		if err != nil {
			return fmt.Errorf("failed to set garden kubeconfig: %w", err)
		}
		return nil
	}

	if Config.GardenKubeconfig == "" {
		return fmt.Errorf("garden kubeconfig is not set in config")
	}
	kubeBytes, err := os.ReadFile(Config.GardenKubeconfig)
	if err != nil {
		return fmt.Errorf("failed to read garden kubeconfig from %q: %w", Config.GardenKubeconfig, err)
	}

	err = w.setGardenConfig(kubeBytes)
	if err != nil {
		return fmt.Errorf("failed to set garden kubeconfig: %w", err)
	}

	secret = &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: Config.KcpNamespace,
			Name:      "gardener-credentials",
		},
		Data: map[string][]byte{
			"kubeconfig": kubeBytes,
		},
	}
	err = kcp.Cluster.GetClient().Create(ctx, secret)
	if apierrors.IsAlreadyExists(err) {
		// some race condition, let's assume the secret correctly created
		return nil
	}
	if err != nil {
		return fmt.Errorf("error creating gardener credentials: %w", err)
	}

	return nil
}

func (w *World) EvaluationContext(ctx context.Context) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	merge := func(c *Cluster, err error) error {
		if err != nil {
			return nil
		}
		data, err := c.EvaluationContext(ctx)
		if err != nil {
			return err
		}
		maps.Copy(result, data)
		return nil
	}

	if err := merge(w.ClusterProvider.KCP(ctx)); err != nil {
		return nil, fmt.Errorf("failed to evaluate KCP cluster: %w", err)
	}
	for id, skr := range w.ClusterProvider.KnownSkrClusters() {
		if err := merge(skr, nil); err != nil {
			return nil, fmt.Errorf("failed to evaluate SKR cluster %q: %w", id, err)
		}
	}
	if err := merge(w.ClusterProvider.Garden(ctx)); err != nil {
		return nil, fmt.Errorf("failed to evaluate Garden cluster: %w", err)
	}

	return result, nil
}
