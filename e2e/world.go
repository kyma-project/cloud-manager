package e2e

import (
	"context"
	"sync"

	"github.com/kyma-project/cloud-manager/e2e/cloud"
	e2econfig "github.com/kyma-project/cloud-manager/e2e/config"
	e2ekeb "github.com/kyma-project/cloud-manager/e2e/keb"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type WorldIntf interface {
	Config() *e2econfig.ConfigType
	Ctx() context.Context
	Cancel()
	StopWaitGroup() *sync.WaitGroup
	RunError() error

	KcpManager() manager.Manager
	GardenManager() manager.Manager

	Kcp() Cluster
	Garden() Cluster
	Keb() e2ekeb.Keb
	Cloud() cloud.Cloud
}

type defaultWorld struct {
	config   *e2econfig.ConfigType
	mCtx     context.Context
	wg       *sync.WaitGroup
	cancel   context.CancelFunc
	runError error

	kcpManager    manager.Manager
	gardenManager manager.Manager

	kcp    Cluster
	garden Cluster

	kebi e2ekeb.Keb

	cloud cloud.Cloud
}

func (w *defaultWorld) Config() *e2econfig.ConfigType {
	return w.config
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

func (w *defaultWorld) KcpManager() manager.Manager {
	return w.kcpManager
}

func (w *defaultWorld) GardenManager() manager.Manager {
	return w.gardenManager
}

func (w *defaultWorld) Kcp() Cluster {
	return w.kcp
}

func (w *defaultWorld) Garden() Cluster {
	return w.garden
}

func (w *defaultWorld) Keb() e2ekeb.Keb {
	return w.kebi
}

func (w *defaultWorld) Cloud() cloud.Cloud {
	return w.cloud
}
