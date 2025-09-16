package e2e

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-multierror"
)

type World interface {
	ClusterProvider() ClusterProvider

	// Stop all clusters - kcp, garden and all skr clusters in the registry but does not
	// delete any skr or shoot
	Stop(ctx context.Context) error
}

type defaultWorld struct {
	clusterProvider ClusterProvider
	//skr             SkrCreator
}

func NewWorld(clusterProvider ClusterProvider) World {
	return &defaultWorld{
		clusterProvider: clusterProvider,
		//skr:             skr,
	}
}

func (w *defaultWorld) ClusterProvider() ClusterProvider {
	return w.clusterProvider
}

//func (w *defaultWorld) SKR() SkrCreator {
//	return w.skr
//}

func (w *defaultWorld) Stop(ctx context.Context) error {
	var result error
	//for _, skr := range w.SKR().AllClusters() {
	//	if err := w.SKR().Remove(skr.Alias()); err != nil {
	//		result = multierror.Append(result, fmt.Errorf("failed to remove skr cluster %s: %w", skr.Alias, err))
	//	}
	//}

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
