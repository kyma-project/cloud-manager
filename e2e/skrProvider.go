package e2e

import (
	"context"
	"fmt"

	e2econfig "github.com/kyma-project/cloud-manager/e2e/config"
	"github.com/kyma-project/cloud-manager/pkg/external/infrastructuremanagerv1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
)

type defaultSkrProvider struct {
	kcp        client.Client
	skrCreator SkrCreator
}

func (p *defaultSkrProvider) GetSKR(ctx context.Context, runtimeID string) (cluster.Cluster, error) {
	rt := &infrastructuremanagerv1.Runtime{}
	err := p.kcp.Get(ctx, types.NamespacedName{
		Namespace: e2econfig.Config.KcpNamespace,
		Name:      runtimeID,
	}, rt)
	if err != nil {
		return nil, fmt.Errorf("could not get runtime %s: %w", runtimeID, err)
	}

	alias := rt.Labels[aliasLabel]
	if alias == "" {
		return nil, fmt.Errorf("runtime %s has no alias label", runtimeID)
	}

	skr := p.skrCreator.Get(alias)
	if skr == nil {
		return nil, fmt.Errorf("could not find skr %s", alias)
	}

	return skr, nil
}
