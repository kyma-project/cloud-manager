package e2e

import (
	"context"
	"sync"

	"github.com/kyma-project/cloud-manager/e2e/sim"
)

type World interface {
	Ctx() context.Context
	Cancel()
	StopWaitGroup() *sync.WaitGroup
	RunError() error

	Kcp() Cluster
	Garden() Cluster
	Sim() sim.Sim
}

type defaultWorld struct {
	mCtx     context.Context
	wg       *sync.WaitGroup
	cancel   context.CancelFunc
	runError error

	kcp    Cluster
	garden Cluster
	simu   sim.Sim
}

func (w *defaultWorld)  Ctx() context.Context {
	return w.mCtx
}

func (w *defaultWorld) Cancel() {
	w.cancel()
}

func (w *defaultWorld) StopWaitGroup() *sync.WaitGroup {
	return w.wg
}

func (w *defaultWorld) RunError() error {
	return w.runError
}

func (w *defaultWorld) Kcp() Cluster {
	return w.kcp
}

func (w *defaultWorld) Garden() Cluster {
	return w.garden
}

func (w *defaultWorld) Sim() sim.Sim {
	return w.simu
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
