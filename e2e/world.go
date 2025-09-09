package e2e

import (
	"context"
	"fmt"

	gardenertypes "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	gardenerconstants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	"github.com/hashicorp/go-multierror"
	e2econfig "github.com/kyma-project/cloud-manager/e2e/config"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/external/infrastructuremanagerv1"
	"github.com/kyma-project/cloud-manager/pkg/external/operatorv1beta2"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type World interface {
	ClusterProvider() ClusterProvider
	SKR() SkrCreator

	// DeleteSKR stops the skr cluster and deletes all related resources (like
	// runtime, gardenercluster, kubeconfig secret, kyma, shoot) from the kcp and garden
	DeleteSKR(ctx context.Context, skr SkrCluster) error

	// Stop all clusters - kcp, garden and all skr clusters in the registry but does not
	// delete any skr or shoot
	Stop(ctx context.Context) error
}

type defaultWorld struct {
	clusterProvider ClusterProvider
	skr             SkrCreator
}

func NewWorld(clusterProvider ClusterProvider, skr SkrCreator) World {
	return &defaultWorld{
		clusterProvider: clusterProvider,
		skr:             skr,
	}
}

func (w *defaultWorld) ClusterProvider() ClusterProvider {
	return w.clusterProvider
}

func (w *defaultWorld) SKR() SkrCreator {
	return w.skr
}

func (w *defaultWorld) Stop(ctx context.Context) error {
	var result error
	for _, skr := range w.SKR().AllClusters() {
		if err := w.SKR().Remove(skr.Alias()); err != nil {
			result = multierror.Append(result, fmt.Errorf("failed to remove skr cluster %s: %w", skr.Alias, err))
		}
	}

	if err := w.ClusterProvider().Stop(); err != nil {
		result = multierror.Append(result, fmt.Errorf("could not stop kcp or garden cluster: %w", err))
	}
	return result
}

func (w *defaultWorld) DeleteSKR(ctx context.Context, skr SkrCluster) error {
	err := w.skr.Remove(skr.Alias())
	if err != nil {
		return fmt.Errorf("could not remove skr %q: %w", skr.Alias, err)
	}

	garden, err := w.ClusterProvider().Garden(ctx)
	if err != nil {
		return fmt.Errorf("could not get Garden cluster: %w", err)
	}

	kcp, err := w.ClusterProvider().KCP(ctx)
	if err != nil {
		return fmt.Errorf("could not get KCP cluster: %w", err)
	}

	err = cleanSkrNoWait(ctx, skr.GetClient())
	if err != nil {
		return fmt.Errorf("could not clean skr: %w", err)
	}

	if err := skr.Stop(); err != nil {
		return fmt.Errorf("could not stop skr cluster: %w", err)
	}

	// Kyma

	kyma := &operatorv1beta2.Kyma{}
	err = kcp.GetClient().Get(ctx, types.NamespacedName{
		Namespace: e2econfig.Config.KcpNamespace,
		Name:      skr.RuntimeID(),
	}, kyma)
	if client.IgnoreNotFound(err) != nil {
		return fmt.Errorf("error getting kyma for skr: %w", err)
	}
	if err == nil {
		err = kcp.GetClient().Delete(ctx, kyma)
		if err != nil {
			return fmt.Errorf("error deleting kyma for skr: %w", err)
		}
	}

	// Runtime

	rt := &infrastructuremanagerv1.Runtime{}
	err = kcp.GetClient().Get(ctx, types.NamespacedName{
		Namespace: e2econfig.Config.KcpNamespace,
		Name:      skr.RuntimeID(),
	}, rt)
	if client.IgnoreNotFound(err) != nil {
		return fmt.Errorf("error getting runtime for skr: %w", err)
	}
	if err == nil {
		err = kcp.GetClient().Delete(ctx, rt)
		if err != nil {
			return fmt.Errorf("error deleting runtime for skr: %w", err)
		}
	}

	// GardenerCluster

	gc := &infrastructuremanagerv1.GardenerCluster{}
	err = kcp.GetClient().Get(ctx, types.NamespacedName{
		Namespace: e2econfig.Config.KcpNamespace,
		Name:      skr.RuntimeID(),
	}, rt)
	if client.IgnoreNotFound(err) != nil {
		return fmt.Errorf("error getting gardenercluster for skr: %w", err)
	}
	kubeSecretName := "kubeconfig-" + skr.RuntimeID()
	kubeSecretNamespace := e2econfig.Config.KcpNamespace
	if err == nil {
		kubeSecretName = gc.Spec.Kubeconfig.Secret.Name
		kubeSecretNamespace = gc.Spec.Kubeconfig.Secret.Namespace
		err = kcp.GetClient().Delete(ctx, gc)
		if err != nil {
			return fmt.Errorf("error deleting gardenercluster for skr: %w", err)
		}
	}

	// Kube secret

	kubeSecret := &corev1.Secret{}
	err = kcp.GetClient().Get(ctx, types.NamespacedName{
		Namespace: kubeSecretNamespace,
		Name:      kubeSecretName,
	}, kubeSecret)
	if client.IgnoreNotFound(err) != nil {
		return fmt.Errorf("error getting kube secret for skr: %w", err)
	}
	if err == nil {
		err = kcp.GetClient().Delete(ctx, kubeSecret)
		if err != nil {
			return fmt.Errorf("error deleting kube secret for skr: %w", err)
		}
	}

	// Shoot

	shoot := &gardenertypes.Shoot{}
	err = garden.GetClient().Get(ctx, types.NamespacedName{
		Namespace: e2econfig.Config.GardenNamespace,
		Name:      skr.ShootName(),
	}, shoot)
	if apierrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("could not load shoot %q: %w", skr.ShootName, err)
	}

	_, err = composed.PatchObjMergeAnnotation(ctx, gardenerconstants.ConfirmationDeletion, "true", shoot, kcp.GetClient())
	if err != nil {
		return fmt.Errorf("error adding deletion confirmation annotation to the shoot: %w", err)
	}

	err = garden.GetClient().Delete(ctx, shoot)
	if err != nil {
		return fmt.Errorf("failed deleting shoot %q: %w", skr.ShootName, err)
	}

	return nil
}

//func (w *defaultWorld) EvaluationContext(ctx context.Context) (map[string]interface{}, error) {
//	result := make(map[string]interface{})
//
//	merge := func(c Cluster, err error) error {
//		if err != nil {
//			return nil
//		}
//		data, err := c.EvaluationContext(ctx)
//		if err != nil {
//			return err
//		}
//		maps.Copy(result, data)
//		return nil
//	}
//
//	if err := merge(w.clusterProvider.KCP(ctx)); err != nil {
//		return nil, fmt.Errorf("failed to evaluate KCP cluster: %w", err)
//	}
//	for id, skr := range w.clusterProvider.KnownSkrClusters() {
//		if err := merge(skr, nil); err != nil {
//			return nil, fmt.Errorf("failed to evaluate SKR cluster %q: %w", id, err)
//		}
//	}
//	if err := merge(w.clusterProvider.Garden(ctx)); err != nil {
//		return nil, fmt.Errorf("failed to evaluate Garden cluster: %w", err)
//	}
//
//	return result, nil
//}
