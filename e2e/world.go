package e2e

import (
	"context"
	"sync"

	"github.com/kyma-project/cloud-manager/e2e/sim"
)

type WorldIntf interface {
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

func (w *defaultWorld) Ctx() context.Context {
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
