package testinfra

import (
	"fmt"
	"github.com/onsi/ginkgo/v2"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

var _ Infra = &infra{}

type infra struct {
	InfraEnv
	InfraDSL

	clusters map[ClusterType]*clusterInfo
}

func (i *infra) KCP() ClusterInfo {
	return i.clusters[ClusterTypeKcp]
}

func (i *infra) SKR() ClusterInfo {
	return i.clusters[ClusterTypeSkr]
}

func (i *infra) Garden() ClusterInfo {
	return i.clusters[ClusterTypeGarden]
}

func (i *infra) Stop() error {
	i.stopControllers()

	var lastErr error
	for name, cluster := range i.clusters {
		ginkgo.By(fmt.Sprintf("Stopping cluster %s", name))
		if err := cluster.env.Stop(); err != nil {
			lastErr = err
		}
	}

	return lastErr
}

// =======================

var _ ClusterInfo = &clusterInfo{}

type clusterInfo struct {
	ClusterEnv
	ClusterDSL

	crdDirs []string
	env     *envtest.Environment
	cfg     *rest.Config
	scheme  *runtime.Scheme
	client  client.Client
}

func (c *clusterInfo) Scheme() *runtime.Scheme {
	return c.scheme
}

func (c *clusterInfo) Client() client.Client {
	return c.client
}

func (c *clusterInfo) Cfg() *rest.Config {
	return c.cfg
}
