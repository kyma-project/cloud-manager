package client

import (
	"context"
	"fmt"

	filestore "cloud.google.com/go/filestore/apiv1"
	"cloud.google.com/go/filestore/apiv1/filestorepb"
	"cloud.google.com/go/longrunning/autogen/longrunningpb"
	"github.com/googleapis/gax-go/v2"
)

type FilestoreClient interface {
	GetFilestoreInstance(ctx context.Context, req *filestorepb.GetInstanceRequest, opts ...gax.CallOption) (*filestorepb.Instance, error)
	CreateFilestoreInstance(ctx context.Context, req *filestorepb.CreateInstanceRequest, opts ...gax.CallOption) (ResultOperation[*filestorepb.Instance], error)
	UpdateFilestoreInstance(ctx context.Context, req *filestorepb.UpdateInstanceRequest, opts ...gax.CallOption) (ResultOperation[*filestorepb.Instance], error)
	DeleteFilestoreInstance(ctx context.Context, req *filestorepb.DeleteInstanceRequest, opts ...gax.CallOption) (VoidOperation, error)

	UpdateFilestoreBackup(ctx context.Context, req *filestorepb.UpdateBackupRequest, opts ...gax.CallOption) (ResultOperation[*filestorepb.Backup], error)
	GetFilestoreBackup(ctx context.Context, req *filestorepb.GetBackupRequest, opts ...gax.CallOption) (*filestorepb.Backup, error)
	ListFilestoreBackups(ctx context.Context, req *filestorepb.ListBackupsRequest, opts ...gax.CallOption) Iterator[*filestorepb.Backup]
	CreateFilestoreBackup(ctx context.Context, req *filestorepb.CreateBackupRequest, opts ...gax.CallOption) (ResultOperation[*filestorepb.Backup], error)
	DeleteFilestoreBackup(ctx context.Context, req *filestorepb.DeleteBackupRequest, opts ...gax.CallOption) (VoidOperation, error)

	GetFilestoreOperation(ctx context.Context, req *longrunningpb.GetOperationRequest, opts ...gax.CallOption) (*longrunningpb.Operation, error)
	ListFilestoreOperations(ctx context.Context, req *longrunningpb.ListOperationsRequest, opts ...gax.CallOption) Iterator[*longrunningpb.Operation]
}

var _ FilestoreClient = (*filestoreClient)(nil)

type filestoreClient struct {
	inner *filestore.CloudFilestoreManagerClient
}

func (c *filestoreClient) GetFilestoreInstance(ctx context.Context, req *filestorepb.GetInstanceRequest, opts ...gax.CallOption) (*filestorepb.Instance, error) {
	return c.inner.GetInstance(ctx, req, opts...)
}

func (c *filestoreClient) CreateFilestoreInstance(ctx context.Context, req *filestorepb.CreateInstanceRequest, opts ...gax.CallOption) (ResultOperation[*filestorepb.Instance], error) {
	return c.inner.CreateInstance(ctx, req, opts...)
}

func (c *filestoreClient) UpdateFilestoreInstance(ctx context.Context, req *filestorepb.UpdateInstanceRequest, opts ...gax.CallOption) (ResultOperation[*filestorepb.Instance], error) {
	return c.inner.UpdateInstance(ctx, req, opts...)
}

func (c *filestoreClient) DeleteFilestoreInstance(ctx context.Context, req *filestorepb.DeleteInstanceRequest, opts ...gax.CallOption) (VoidOperation, error) {
	return c.inner.DeleteInstance(ctx, req, opts...)
}

// Backup ======================================

func (c *filestoreClient) UpdateFilestoreBackup(ctx context.Context, req *filestorepb.UpdateBackupRequest, opts ...gax.CallOption) (ResultOperation[*filestorepb.Backup], error) {
	return c.inner.UpdateBackup(ctx, req, opts...)
}

func (c *filestoreClient) GetFilestoreBackup(ctx context.Context, req *filestorepb.GetBackupRequest, opts ...gax.CallOption) (*filestorepb.Backup, error) {
	return c.inner.GetBackup(ctx, req, opts...)
}

func (c *filestoreClient) ListFilestoreBackups(ctx context.Context, req *filestorepb.ListBackupsRequest, opts ...gax.CallOption) Iterator[*filestorepb.Backup] {
	return c.inner.ListBackups(ctx, req, opts...)
}

func (c *filestoreClient) CreateFilestoreBackup(ctx context.Context, req *filestorepb.CreateBackupRequest, opts ...gax.CallOption) (ResultOperation[*filestorepb.Backup], error) {
	return c.inner.CreateBackup(ctx, req, opts...)
}

func (c *filestoreClient) DeleteFilestoreBackup(ctx context.Context, req *filestorepb.DeleteBackupRequest, opts ...gax.CallOption) (VoidOperation, error) {
	return c.inner.DeleteBackup(ctx, req, opts...)
}

// Restore =====================================

func (c *filestoreClient) RestoreFilestoreInstance(ctx context.Context, req *filestorepb.RestoreInstanceRequest, opts ...gax.CallOption) (ResultOperation[*filestorepb.Instance], error) {
	return c.inner.RestoreInstance(ctx, req, opts...)
}

// Operation =====================================

func (c *filestoreClient) GetFilestoreOperation(ctx context.Context, req *longrunningpb.GetOperationRequest, opts ...gax.CallOption) (*longrunningpb.Operation, error) {
	return c.inner.GetOperation(ctx, req, opts...)
}

func (c *filestoreClient) ListFilestoreOperations(ctx context.Context, req *longrunningpb.ListOperationsRequest, opts ...gax.CallOption) Iterator[*longrunningpb.Operation] {
	return c.inner.ListOperations(ctx, req, opts...)
}

// High level functions =====================================

func (c *filestoreClient) FindFilestoreRestoreOperation(ctx context.Context, projectId, location, instanceId string) (*longrunningpb.Operation, error) {
	parent := fmt.Sprintf("projects/%s/locations/%s", projectId, location)
	destination := fmt.Sprintf("projects/%s/locations/%s/instances/%s", projectId, location, instanceId)
	targetFilter := fmt.Sprintf("metadata.target = \"%s\"", destination)
	verbFilter := "metadata.verb = \"restore\""
	filters := fmt.Sprintf("%s AND %s AND done = false", targetFilter, verbFilter)

	it := c.ListFilestoreOperations(ctx, &longrunningpb.ListOperationsRequest{
		Name:   parent,
		Filter: filters,
	})
	var runningOperation *longrunningpb.Operation
	for op, err := range it.All() {
		if err != nil {
			return nil, err
		}
		if !op.GetDone() {
			runningOperation = op
		}
	}
	if runningOperation == nil || runningOperation.Name == "" {
		return nil, nil
	}
	return runningOperation, nil
}
