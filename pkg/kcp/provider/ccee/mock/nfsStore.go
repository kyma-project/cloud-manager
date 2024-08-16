package mock

import (
	"context"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/networks"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/subnets"
	"github.com/gophercloud/gophercloud/v2/openstack/sharedfilesystems/v2/sharenetworks"
	"github.com/gophercloud/gophercloud/v2/openstack/sharedfilesystems/v2/shares"
)

type nfsStore struct {
}

func (n *nfsStore) ListInternalNetworks(ctx context.Context, name string) ([]networks.Network, error) {
	//TODO implement me
	panic("implement me")
}

func (n *nfsStore) GetNetwork(ctx context.Context, id string) (*networks.Network, error) {
	//TODO implement me
	panic("implement me")
}

func (n *nfsStore) ListSubnets(ctx context.Context, networkId string) ([]subnets.Subnet, error) {
	//TODO implement me
	panic("implement me")
}

func (n *nfsStore) GetSubnet(ctx context.Context, id string) (*subnets.Subnet, error) {
	//TODO implement me
	panic("implement me")
}

func (n *nfsStore) ListShareNetworks(ctx context.Context, networkId string) ([]sharenetworks.ShareNetwork, error) {
	//TODO implement me
	panic("implement me")
}

func (n *nfsStore) GetShareNetwork(ctx context.Context, id string) (*sharenetworks.ShareNetwork, error) {
	//TODO implement me
	panic("implement me")
}

func (n *nfsStore) CreateShareNetwork(ctx context.Context, networkId, subnetId, name string) (*sharenetworks.ShareNetwork, error) {
	//TODO implement me
	panic("implement me")
}

func (n *nfsStore) DeleteShareNetwork(ctx context.Context, id string) error {
	//TODO implement me
	panic("implement me")
}

func (n *nfsStore) ListShares(ctx context.Context, shareNetworkId string) ([]shares.Share, error) {
	//TODO implement me
	panic("implement me")
}

func (n *nfsStore) GetShare(ctx context.Context, id string) (*shares.Share, error) {
	//TODO implement me
	panic("implement me")
}

func (n *nfsStore) CreateShare(ctx context.Context, shareNetworkId, name string, size int, snapshotID string, metadata map[string]string) (*shares.Share, error) {
	//TODO implement me
	panic("implement me")
}

func (n *nfsStore) DeleteShare(ctx context.Context, id string) error {
	//TODO implement me
	panic("implement me")
}

func (n *nfsStore) ListShareExportLocations(ctx context.Context, id string) ([]shares.ExportLocation, error) {
	//TODO implement me
	panic("implement me")
}

func (n *nfsStore) ListShareAccessRights(ctx context.Context, id string) ([]shares.AccessRight, error) {
	//TODO implement me
	panic("implement me")
}

func (n *nfsStore) GrantShareAccess(ctx context.Context, id string, cidr string) (*shares.AccessRight, error) {
	//TODO implement me
	panic("implement me")
}

func (n *nfsStore) RevokeShareAccess(ctx context.Context, shareId, accessId string) error {
	//TODO implement me
	panic("implement me")
}
