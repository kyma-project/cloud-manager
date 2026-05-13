package client

import (
	"context"
	"net/http"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack/sharedfilesystems/v2/shares"
	"github.com/gophercloud/gophercloud/v2/openstack/sharedfilesystems/v2/snapshots"
)

type SnapshotClient interface {
	CreateSnapshot(ctx context.Context, opts snapshots.CreateOpts) (*snapshots.Snapshot, error)
	GetSnapshot(ctx context.Context, snapshotId string) (*snapshots.Snapshot, error)
	DeleteSnapshot(ctx context.Context, snapshotId string) error
	ListSnapshots(ctx context.Context, opts snapshots.ListOpts) ([]snapshots.Snapshot, error)
	RevertShareToSnapshot(ctx context.Context, shareId string, snapshotId string) error
}

var _ SnapshotClient = (*snapshotClient)(nil)

type snapshotClient struct {
	shareSvc *gophercloud.ServiceClient
}

func (c *snapshotClient) CreateSnapshot(ctx context.Context, opts snapshots.CreateOpts) (*snapshots.Snapshot, error) {
	snapshot, err := snapshots.Create(ctx, c.shareSvc, opts).Extract()
	if err != nil {
		return nil, err
	}
	return snapshot, nil
}

func (c *snapshotClient) GetSnapshot(ctx context.Context, snapshotId string) (*snapshots.Snapshot, error) {
	snapshot, err := snapshots.Get(ctx, c.shareSvc, snapshotId).Extract()
	if gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return snapshot, nil
}

func (c *snapshotClient) DeleteSnapshot(ctx context.Context, snapshotId string) error {
	err := snapshots.Delete(ctx, c.shareSvc, snapshotId).ExtractErr()
	if gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
		return nil
	}
	return err
}

func (c *snapshotClient) ListSnapshots(ctx context.Context, opts snapshots.ListOpts) ([]snapshots.Snapshot, error) {
	page, err := snapshots.ListDetail(c.shareSvc, opts).AllPages(ctx)
	if err != nil {
		return nil, err
	}
	arr, err := snapshots.ExtractSnapshots(page)
	if err != nil {
		return nil, err
	}
	return arr, nil
}

func (c *snapshotClient) RevertShareToSnapshot(ctx context.Context, shareId string, snapshotId string) error {
	err := shares.Revert(ctx, c.shareSvc, shareId, shares.RevertOpts{
		SnapshotID: snapshotId,
	}).ExtractErr()
	return err
}
