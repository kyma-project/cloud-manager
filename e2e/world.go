package e2e

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/kyma-project/cloud-manager/e2e/sim"
)

type World interface {
	ClusterProvider() ClusterProvider
	Sim() sim.Sim

	// Stop all clusters - kcp, garden and all skr clusters in the registry but does not
	// delete any skr or shoot
	Stop(ctx context.Context) error
}

type defaultWorld struct {
	clusterProvider ClusterProvider
	simu            sim.Sim
}

func NewWorld(clusterProvider ClusterProvider, simu sim.Sim) World {
	return &defaultWorld{
		clusterProvider: clusterProvider,
		simu:            simu,
	}
}

func (w *defaultWorld) ClusterProvider() ClusterProvider {
	return w.clusterProvider
}

func (w *defaultWorld) Sim() sim.Sim {
	return w.simu
}

func (w *defaultWorld) Stop(ctx context.Context) error {
	var result error

	if err := w.ClusterProvider().Stop(); err != nil {
		result = multierror.Append(result, fmt.Errorf("could not stop kcp or garden cluster: %w", err))
	}
	return result
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
