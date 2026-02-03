package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/elliotchance/pie/v2"
	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack/sharedfilesystems/v2/shareaccessrules"
	"github.com/gophercloud/gophercloud/v2/openstack/sharedfilesystems/v2/sharenetworks"
	"github.com/gophercloud/gophercloud/v2/openstack/sharedfilesystems/v2/shares"
)

type ShareClient interface {
	ListShareNetworks(ctx context.Context, opts sharenetworks.ListOpts) ([]sharenetworks.ShareNetwork, error)
	GetShareNetwork(ctx context.Context, shareNetworkId string) (*sharenetworks.ShareNetwork, error)
	CreateShareNetwork(ctx context.Context, opts sharenetworks.CreateOpts) (*sharenetworks.ShareNetwork, error)
	DeleteShareNetwork(ctx context.Context, shareNetworkId string) error

	// ListShares
	// share.status possible values https://docs.openstack.org/manila/latest/user/create-and-manage-shares.html
	// These “-ing” states end in an “available” state if everything goes well. They may end up in an “error” state in case there is an issue.
	// * available
	// * error
	// * creating
	// * extending
	// * shrinking
	// * migrating
	ListShares(ctx context.Context, opts shares.ListOpts) ([]shares.Share, error)
	GetShare(ctx context.Context, shareId string) (*shares.Share, error)
	CreateShare(ctx context.Context, opts shares.CreateOpts) (*shares.Share, error)
	DeleteShare(ctx context.Context, shareId string) error

	ShareShrink(ctx context.Context, shareId string, newSize int) error
	ShareExtend(ctx context.Context, shareId string, newSize int) error

	ListShareExportLocations(ctx context.Context, shareId string) ([]shares.ExportLocation, error)

	ListShareAccessRules(ctx context.Context, shareId string) ([]ShareAccess, error)
	GrantShareAccess(ctx context.Context, shareId string, cidr string) (*ShareAccess, error)
	RevokeShareAccess(ctx context.Context, shareId, accessId string) error

	// high level opinionated

	ListShareNetworksByNetworkId(ctx context.Context, networkId string) ([]sharenetworks.ShareNetwork, error)
	CreateShareNetworkOp(ctx context.Context, networkId, subnetId, name string) (*sharenetworks.ShareNetwork, error)
	ListSharesInShareNetwork(ctx context.Context, shareNetworkId string) ([]shares.Share, error)
	CreateShareOp(ctx context.Context, shareNetworkId, name string, size int, snapshotID string, metadata map[string]string) (*shares.Share, error)
}

type ShareAccess struct {
	ID          string
	ShareID     string
	AccessType  string
	AccessTo    string
	AccessKey   string
	State       string
	AccessLevel string
}

var _ ShareClient = (*shareClient)(nil)

type shareClient struct {
	shareSvc *gophercloud.ServiceClient
}

// low level methods matching the gophercloud api ================================

func (c *shareClient) ListShareNetworks(ctx context.Context, opts sharenetworks.ListOpts) ([]sharenetworks.ShareNetwork, error) {
	page, err := sharenetworks.ListDetail(c.shareSvc, opts).AllPages(ctx)
	if err != nil {
		return nil, err
	}
	arr, err := sharenetworks.ExtractShareNetworks(page)
	if err != nil {
		return nil, err
	}
	return arr, nil
}

func (c *shareClient) GetShareNetwork(ctx context.Context, shareNetworkId string) (*sharenetworks.ShareNetwork, error) {
	shareNet, err := sharenetworks.Get(ctx, c.shareSvc, shareNetworkId).Extract()
	if gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return shareNet, nil
}

func (c *shareClient) CreateShareNetwork(ctx context.Context, opts sharenetworks.CreateOpts) (*sharenetworks.ShareNetwork, error) {
	shareNet, err := sharenetworks.Create(ctx, c.shareSvc, opts).Extract()
	if err != nil {
		return nil, err
	}
	return shareNet, nil
}

func (c *shareClient) DeleteShareNetwork(ctx context.Context, shareNetworkId string) error {
	err := sharenetworks.Delete(ctx, c.shareSvc, shareNetworkId).ExtractErr()
	return err
}

func (c *shareClient) ListShares(ctx context.Context, opts shares.ListOpts) ([]shares.Share, error) {
	page, err := shares.ListDetail(c.shareSvc, opts).AllPages(ctx)
	if err != nil {
		return nil, err
	}
	arr, err := shares.ExtractShares(page)
	if err != nil {
		return nil, err
	}
	return arr, nil
}

func (c *shareClient) GetShare(ctx context.Context, shareId string) (*shares.Share, error) {
	share, err := shares.Get(ctx, c.shareSvc, shareId).Extract()
	if gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return share, nil
}

func (c *shareClient) CreateShare(ctx context.Context, opts shares.CreateOpts) (*shares.Share, error) {
	share, err := shares.Create(ctx, c.shareSvc, opts).Extract()
	if err != nil {
		return nil, err
	}
	return share, nil
}

func (c *shareClient) DeleteShare(ctx context.Context, shareId string) error {
	err := shares.Delete(ctx, c.shareSvc, shareId).ExtractErr()
	return err
}

func (c *shareClient) ShareShrink(ctx context.Context, shareId string, newSize int) error {
	err := shares.Shrink(ctx, c.shareSvc, shareId, shares.ShrinkOpts{
		NewSize: newSize,
	}).ExtractErr()
	return err
}

func (c *shareClient) ShareExtend(ctx context.Context, shareId string, newSize int) error {
	err := shares.Extend(ctx, c.shareSvc, shareId, shares.ExtendOpts{
		NewSize: newSize,
	}).ExtractErr()
	return err
}

func (c *shareClient) ListShareExportLocations(ctx context.Context, shareId string) ([]shares.ExportLocation, error) {
	page, err := shares.ListExportLocations(ctx, c.shareSvc, shareId).Extract()
	if err != nil {
		return nil, err
	}
	return page, nil
}

func NewShareAccessFromSharesAccessRight(o *shares.AccessRight) *ShareAccess {
	return &ShareAccess{
		ID:          o.ID,
		ShareID:     o.ShareID,
		AccessType:  o.AccessType,
		AccessTo:    o.AccessTo,
		AccessKey:   o.AccessKey,
		State:       o.State,
		AccessLevel: o.AccessLevel,
	}
}

func NewShareAccessFromShareAccessRulesShareAccess(o *shareaccessrules.ShareAccess) *ShareAccess {
	return &ShareAccess{
		ID:          o.ID,
		ShareID:     o.ShareID,
		AccessType:  o.AccessType,
		AccessTo:    o.AccessTo,
		AccessKey:   o.AccessKey,
		State:       o.State,
		AccessLevel: o.AccessLevel,
	}
}

func (c *shareClient) ListShareAccessRules(ctx context.Context, shareId string) ([]ShareAccess, error) {
	// https://dashboard.eu-de-1.cloud.sap/kyma/kyma-dev-02/shared-filesystem-storage/shares/d6b9995f-4c6a-4d95-8b51-6ca88c753f0f/rules

	arr, err := shareaccessrules.List(ctx, c.shareSvc, shareId).Extract()
	if gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error listing access rights: %w", err)
	}
	return pie.Map(arr, func(x shareaccessrules.ShareAccess) ShareAccess {
		x.ShareID = shareId
		return *NewShareAccessFromShareAccessRulesShareAccess(&x)
	}), nil
}

func (c *shareClient) GrantShareAccess(ctx context.Context, shareId string, cidr string) (*ShareAccess, error) {
	ar, err := shares.GrantAccess(ctx, c.shareSvc, shareId, shares.GrantAccessOpts{
		AccessType:  "ip",
		AccessTo:    cidr,
		AccessLevel: "rw",
	}).Extract()
	if err != nil {
		return nil, fmt.Errorf("error granting access to share: %w", err)
	}
	ar.ShareID = shareId
	return NewShareAccessFromSharesAccessRight(ar), nil
}

func (c *shareClient) RevokeShareAccess(ctx context.Context, shareId, accessId string) error {
	err := shares.RevokeAccess(ctx, c.shareSvc, shareId, shares.RevokeAccessOpts{
		AccessID: accessId,
	}).ExtractErr()
	if err != nil {
		return fmt.Errorf("error revoking access to share: %w", err)
	}
	return nil
}

// 	// high level opinionated ===========================================================

func (c *shareClient) ListShareNetworksByNetworkId(ctx context.Context, networkId string) ([]sharenetworks.ShareNetwork, error) {
	return c.ListShareNetworks(ctx, sharenetworks.ListOpts{
		NeutronNetID: networkId,
	})
}

func (c *shareClient) CreateShareNetworkOp(ctx context.Context, networkId, subnetId, name string) (*sharenetworks.ShareNetwork, error) {
	return c.CreateShareNetwork(ctx, sharenetworks.CreateOpts{
		NeutronNetID:    networkId,
		NeutronSubnetID: subnetId,
		Name:            name,
	})
}

func (c *shareClient) ListSharesInShareNetwork(ctx context.Context, shareNetworkId string) ([]shares.Share, error) {
	return c.ListShares(ctx, shares.ListOpts{
		ShareNetworkID: shareNetworkId,
	})
}

func (c *shareClient) CreateShareOp(ctx context.Context, shareNetworkId, name string, size int, snapshotID string, metadata map[string]string) (*shares.Share, error) {
	return c.CreateShare(ctx, shares.CreateOpts{
		ShareProto:     "NFS",
		Size:           size,
		Name:           name,
		ShareNetworkID: shareNetworkId,
		SnapshotID:     snapshotID,
		Metadata:       metadata,
	})
}
