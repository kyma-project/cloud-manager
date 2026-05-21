package gardener

import (
	"context"
	"time"

	scopeconfig "github.com/kyma-project/cloud-manager/pkg/kcp/scope/config"
	"k8s.io/utils/clock"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var defaultGardenerNamespaceProvider GardenerNamespaceProvider

func init() {
	SetDefaultGardenerNamespaceProvider()
}

func DefaultGardenerNamespaceProvider() GardenerNamespaceProvider {
	return defaultGardenerNamespaceProvider
}

func SetDefaultGardenerNamespaceProvider() {
	defaultGardenerNamespaceProvider = &cachedGardenerNamespaceProvider{
		clock:      &clock.RealClock{},
		expiration: 5 * time.Minute,
		inner:      &gardenerNamespaceProvider{},
	}
}

type GardenerNamespaceProvider interface {
	GetGardenerNamespace(ctx context.Context, kcpApiReader client.Reader) (string, error)
}

// gardenerNamespaceProvider =====================================================================

type gardenerNamespaceProvider struct {
}

func (p *gardenerNamespaceProvider) GetGardenerNamespace(ctx context.Context, kcpApiReader client.Reader) (string, error) {
	out, err := CreateGardenerClient(ctx, CreateGardenerClientInput{
		KcpClient:                 kcpApiReader,
		GardenerFallbackNamespace: scopeconfig.ScopeConfig.GardenerNamespace,
	})
	if err != nil {
		return "", err
	}

	return out.Namespace, nil
}

// cachedGardenerNamespaceProvider =====================================================================

type cachedGardenerNamespaceProvider struct {
	clock      clock.Clock
	expiration time.Duration
	inner      GardenerNamespaceProvider

	value    *string
	lastRead time.Time
}

func (c *cachedGardenerNamespaceProvider) GetGardenerNamespace(ctx context.Context, kcpApiReader client.Reader) (string, error) {
	if c.isValid() {
		return *c.value, nil
	}
	val, err := c.inner.GetGardenerNamespace(ctx, kcpApiReader)
	if err != nil {
		return "", err
	}
	c.lastRead = c.clock.Now()
	c.value = &val
	return val, nil
}

func (c *cachedGardenerNamespaceProvider) isValid() bool {
	if c.value == nil {
		return false
	}
	if c.clock.Since(c.lastRead) > c.expiration {
		return false
	}
	return true
}

// fixedValueGardenerNamespaceProvider =====================================================================

type fixedValueGardenerNamespaceProvider struct {
	value string
}

func (f *fixedValueGardenerNamespaceProvider) GetGardenerNamespace(ctx context.Context, kcpApiReader client.Reader) (string, error) {
	return f.value, nil
}
