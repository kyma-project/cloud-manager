package e2e

import (
	"context"
	"fmt"
	"os"

	"github.com/kyma-project/cloud-manager/config/crd"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type WorldFactory struct {
}

func NewWorldFactory() *WorldFactory {
	return &WorldFactory{}
}

func (f *WorldFactory) Create(ctx context.Context) (World, error) {
	clusterProvider := newClusterProvider()
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

	skr := NewSkrCreator(kcp, garden)

	return NewWorld(clusterProvider, skr), nil
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
		Config.GardenNamespace = rawConfig.Contexts[rawConfig.CurrentContext].Namespace
	}

	return nil
}

func (f *WorldFactory) initGardenerCredentials(ctx context.Context, kcp Cluster) error {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: Config.KcpNamespace,
			Name:      "gardener-credentials",
		},
	}

	err := kcp.Cluster.GetAPIReader().Get(ctx, client.ObjectKeyFromObject(secret), secret)
	if err == nil {
		// already exists
		err = f.setGardenNamespaceInConfig(secret.Data["kubeconfig"])
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

	err = f.setGardenNamespaceInConfig(kubeBytes)
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

func (f *WorldFactory) installCrds(ctx context.Context, kcp Cluster) error {
	arr, err := crd.KCP_All()
	if err != nil {
		return fmt.Errorf("error reading CRDs: %w", err)
	}
	err = Install(ctx, kcp.GetClient(), arr)
	if err != nil {
		return fmt.Errorf("error installing CRDs: %w", err)
	}
	return nil
}

//
//func (f *WorldFactory) _installCrds(ctx context.Context, kcp Cluster) error {
//	files, err := crd.FS.ReadDir("bases")
//	if err != nil {
//		return fmt.Errorf("error reading crd bases: %w", err)
//	}
//	for _, file := range files {
//		if !strings.HasPrefix(file.Name(), "cloud-control.kyma-project.io_") {
//			continue
//		}
//		content, err := crd.FS.ReadFile(path.Join("bases", file.Name()))
//		if err != nil {
//			return fmt.Errorf("error reading crd file bases/%s: %w", file.Name(), err)
//		}
//		err = f.applyCrdFile(ctx, kcp, content)
//		if err != nil {
//			return fmt.Errorf("error applying crd file bases/%s: %w", file.Name(), err)
//		}
//	}
//
//	files, err = crd.FS.ReadDir("kim")
//	if err != nil {
//		return fmt.Errorf("error reading crd kim: %w", err)
//	}
//	for _, file := range files {
//		content, err := crd.FS.ReadFile(path.Join("kim", file.Name()))
//		if err != nil {
//			return fmt.Errorf("error reading crd file kim/%s: %w", file.Name(), err)
//		}
//		err = f.applyCrdFile(ctx, kcp, content)
//		if err != nil {
//			return fmt.Errorf("error applying crd file kim/%s: %w", file.Name(), err)
//		}
//	}
//
//	files, err = crd.FS.ReadDir("operator")
//	if err != nil {
//		return fmt.Errorf("error reading crd operator: %w", err)
//	}
//	for _, file := range files {
//		content, err := crd.FS.ReadFile(path.Join("operator", file.Name()))
//		if err != nil {
//			return fmt.Errorf("error reading crd file operator/%s: %w", file.Name(), err)
//		}
//		err = f.applyCrdFile(ctx, kcp, content)
//		if err != nil {
//			return fmt.Errorf("error applying crd file operator/%s: %w", file.Name(), err)
//		}
//	}
//
//	return nil
//}
//
//func (f *WorldFactory) _applyCrdFile(ctx context.Context, kcp Cluster, content []byte) error {
//	objArr, err := util.YamlMultiDecodeToUnstructured(content)
//	if err != nil {
//		return fmt.Errorf("error unmarshalling crd: %w", err)
//	}
//	for _, desiredObj := range objArr {
//		key := client.ObjectKeyFromObject(desiredObj)
//		existingObj := desiredObj.DeepCopyObject().(*unstructured.Unstructured)
//		err = kcp.GetClient().Get(ctx, key, existingObj)
//		if client.IgnoreNotFound(err) != nil {
//			return fmt.Errorf("error getting existing crd %s: %w", key, err)
//		}
//		if err != nil {
//			// does not exist, should be created
//			err = kcp.GetClient().Create(ctx, desiredObj)
//			if err != nil {
//				return fmt.Errorf("error creating crd %s: %w", key, err)
//			}
//		} else {
//			// exists, should be updated
//			// TODO: maybe this should be reversed, since not all kinds have the `spec` filed: copy existing metadata to the desired object
//			spec, _, err := unstructured.NestedMap(existingObj.Object, "spec")
//			if err != nil {
//				return fmt.Errorf("error getting existing crd %s spec: %w", key, err)
//			}
//			err = unstructured.SetNestedField(existingObj.Object, spec, "spec")
//			if err != nil {
//				return fmt.Errorf("error setting existing crd %s spec: %w", key, err)
//			}
//			err = kcp.GetClient().Update(ctx, existingObj)
//			if err != nil {
//				return fmt.Errorf("error updating existing crd %s: %w", key, err)
//			}
//		}
//	}
//	return nil
//}
