package client

import (
	"context"
	"fmt"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack"
)

type ClientFactory struct {
	pp ProviderParams

	pi       *ProvidedInfo
	netSvc   *gophercloud.ServiceClient
	shareSvc *gophercloud.ServiceClient
}

func NewClientFactory(pp ProviderParams) *ClientFactory {
	return &ClientFactory{
		pp: pp,
	}
}

func (b *ClientFactory) ensureProviderInfo(ctx context.Context) error {
	if b.pi != nil {
		return nil
	}
	pi, err := NewProviderClient(ctx, b.pp)
	if err != nil {
		return err
	}
	b.pi = pi
	return nil
}

func (b *ClientFactory) ensureNetSvc(ctx context.Context) error {
	if b.netSvc != nil {
		return nil
	}
	if err := b.ensureProviderInfo(ctx); err != nil {
		return err
	}
	netSvc, err := openstack.NewNetworkV2(b.pi.ProviderClient, b.pi.EndpointOptions)
	if err != nil {
		return fmt.Errorf("failed to create network v2 client: %w", err)
	}
	b.netSvc = netSvc
	return nil
}

func (b *ClientFactory) ensureShareSvc(ctx context.Context) error {
	if b.shareSvc != nil {
		return nil
	}
	if err := b.ensureProviderInfo(ctx); err != nil {
		return err
	}
	// IMPORTANT to be able to choose manila api v2 - otherwise since CC is advertising both v1 and v2, it will pick first - v1
	gophercloud.ServiceTypeAliases["shared-file-system"] = []string{"sharev2"}
	shareSvc, err := openstack.NewSharedFileSystemV2(b.pi.ProviderClient, b.pi.EndpointOptions)
	if err != nil {
		return fmt.Errorf("failed to create shared file system v2 client: %w", err)
	}
	b.shareSvc = shareSvc
	return nil
}

func (b *ClientFactory) NetworkClient(ctx context.Context) (NetworkClient, error) {
	if err := b.ensureNetSvc(ctx); err != nil {
		return nil, err
	}
	return &networkClient{netSvc: b.netSvc}, nil
}

func (b *ClientFactory) SubnetClient(ctx context.Context) (SubnetClient, error) {
	if err := b.ensureNetSvc(ctx); err != nil {
		return nil, err
	}
	return &subnetClient{netSvc: b.netSvc}, nil
}

func (b *ClientFactory) RouterClient(ctx context.Context) (RouterClient, error) {
	if err := b.ensureNetSvc(ctx); err != nil {
		return nil, err
	}
	return &routerClient{netSvc: b.netSvc}, nil
}

func (b *ClientFactory) PortClient(ctx context.Context) (PortClient, error) {
	if err := b.ensureNetSvc(ctx); err != nil {
		return nil, err
	}
	return &portClient{netSvc: b.netSvc}, nil
}

func (b *ClientFactory) ShareClient(ctx context.Context) (ShareClient, error) {
	if err := b.ensureShareSvc(ctx); err != nil {
		return nil, err
	}
	return &shareClient{shareSvc: b.shareSvc}, nil
}
