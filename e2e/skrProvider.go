package e2e

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/e2e/sim"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
)

type defaultSkrProvider struct {
	kcp        client.Client
	skrCreator SkrCreator
}

var _ sim.SkrProvider = (*defaultSkrProvider)(nil)

func NewSkrProvider(kcp client.Client, skr SkrCreator) sim.SkrProvider {
	return &defaultSkrProvider{
		kcp:        kcp,
		skrCreator: skr,
	}
}

func (p *defaultSkrProvider) GetSKR(ctx context.Context, runtimeID string) (cluster.Cluster, error) {
	skr := p.skrCreator.GetByRuntimeId(runtimeID)
	if skr == nil {
		return nil, fmt.Errorf("could not find skr with runtimeID %q", runtimeID)
	}

	return skr, nil
}
