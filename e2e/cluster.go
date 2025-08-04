package e2e

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-multierror"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
)

type Cluster struct {
	cluster.Cluster

	runCtx     context.Context
	cancelRun  context.CancelFunc
	started    bool
	stoppingCh chan error

	resources map[string]*ResourceInfo
}

func (c *Cluster) IsStarted() bool {
	return c.started
}

func (c *Cluster) Start(ctx context.Context) error {
	if c.started {
		return fmt.Errorf("cluster already started")
	}

	c.runCtx, c.cancelRun = context.WithCancel(ctx)
	c.stoppingCh = make(chan error)

	var err error
	go func() {
		err = c.Cluster.Start(c.runCtx)
		c.stoppingCh <- err
		close(c.stoppingCh)
	}()

	// Wait for the cluster to be started... and as a side effect informers synced
	ok := c.Cluster.GetCache().WaitForCacheSync(c.runCtx)

	if err != nil {
		err = multierror.Append(err, fmt.Errorf("failed to start cluster: %w", err))
	}
	if !ok {
		err = multierror.Append(err, fmt.Errorf("failed to wait for cache sync"))
	}

	return err
}

func (c *Cluster) Stop() error {
	if !c.started {
		return fmt.Errorf("cluster not started")
	}
	c.cancelRun()
	err := <-c.stoppingCh
	c.started = false
	return err
}

func (c *Cluster) AddResources(ctx context.Context, arr []*ResourceDeclaration) error {
	if !c.started {
		return fmt.Errorf("cluster not started")
	}

	for _, decl := range arr {
		if _, ok := c.resources[decl.Alias]; ok {
			return fmt.Errorf("alias %s already declared", decl.Alias)
		}
		c.resources[decl.Alias] = &ResourceInfo{
			ResourceDeclaration: *decl,
			Evaluated:           false,
		}
	}

	ok := c.Cluster.GetCache().WaitForCacheSync(ctx)
	if !ok {
		return fmt.Errorf("failed to wait for cache sync after adding resources")
	}

	return nil
}
