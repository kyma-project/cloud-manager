package e2e

import (
	"context"
	"fmt"
	"os"

	"github.com/kyma-project/cloud-manager/config/crd"
	e2econfig "github.com/kyma-project/cloud-manager/e2e/config"
	"github.com/kyma-project/cloud-manager/pkg/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func InitializeKcp(ctx context.Context, kcpClient client.Client, config *e2econfig.ConfigType) error {
	// install crds
	arr, err := crd.KCP_All()
	if err != nil {
		return fmt.Errorf("error reading CRDs: %w", err)
	}
	err = util.Apply(ctx, kcpClient, arr)
	if err != nil {
		return fmt.Errorf("error installing CRDs: %w", err)
	}

	// kcp-system namespace
	ns := &corev1.Namespace{}
	err = kcpClient.Get(ctx, types.NamespacedName{Name: "kcp-system"}, ns)
	if client.IgnoreNotFound(err) != nil {
		return fmt.Errorf("error reading kcp-system namespace: %w", err)
	}
	if err != nil {
		// namespace does not exist
		ns.Name = "kcp-system"
		err = kcpClient.Create(ctx, ns)
		if err != nil {
			return fmt.Errorf("error creating kcp-system namespace: %w", err)
		}
	}

	// gardener credentials
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: config.KcpNamespace,
			Name:      "gardener-credentials",
		},
	}

	err = kcpClient.Get(ctx, client.ObjectKeyFromObject(secret), secret)
	if err == nil {
		if config.OverwriteGardenerCredentials {
			err = kcpClient.Delete(ctx, secret)
			if err != nil {
				return fmt.Errorf("error deleting existing garden secret: %w", err)
			}
		} else {
			return fmt.Errorf("secret %s already exists", secret.Name)
		}
	}

	if config.GardenKubeconfig == "" {
		return fmt.Errorf("garden kubeconfig is not set in config")
	}
	kubeBytes, err := os.ReadFile(config.GardenKubeconfig)
	if err != nil {
		return fmt.Errorf("failed to read garden kubeconfig from %q: %w", config.GardenKubeconfig, err)
	}

	err = config.SetGardenNamespaceFromKubeconfigBytes(kubeBytes)
	if err != nil {
		return fmt.Errorf("failed to set garden kubeconfig: %w", err)
	}

	secret = &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: config.KcpNamespace,
			Name:      "gardener-credentials",
		},
		Data: map[string][]byte{
			"kubeconfig": kubeBytes,
		},
	}
	err = kcpClient.Create(ctx, secret)
	if err != nil {
		return fmt.Errorf("error creating gardener credentials: %w", err)
	}

	return nil
}
