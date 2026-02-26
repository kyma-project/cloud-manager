package mock2

import (
	"context"

	"cloud.google.com/go/filestore/apiv1/filestorepb"
	"github.com/googleapis/gax-go/v2"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
)

func (s *store) GetFilestoreInstance(ctx context.Context, req *filestorepb.GetInstanceRequest, _ ...gax.CallOption) (*filestorepb.Instance, error) {
	panic("implement me")
}

func (s *store) CreateFilestoreInstance(ctx context.Context, req *filestorepb.CreateInstanceRequest, _ ...gax.CallOption) (gcpclient.WaitableOperationWithResult[*filestorepb.Instance], error) {
	panic("implement me")
}

func (s *store) UpdateFilestoreInstance(ctx context.Context, req *filestorepb.UpdateInstanceRequest, _ ...gax.CallOption) (gcpclient.WaitableOperationWithResult[*filestorepb.Instance], error) {
	panic("implement me")
}

func (s *store) DeleteFilestoreInstance(ctx context.Context, req *filestorepb.DeleteInstanceRequest, _ ...gax.CallOption) (gcpclient.WaitableVoidOperation, error) {
	panic("implement me")
}

// Backup ======================================

func (s *store) UpdateFilestoreBackup(ctx context.Context, req *filestorepb.UpdateBackupRequest, _ ...gax.CallOption) (gcpclient.WaitableOperationWithResult[*filestorepb.Backup], error) {
	panic("implement me")
}

func (s *store) GetFilestoreBackup(ctx context.Context, req *filestorepb.GetBackupRequest, _ ...gax.CallOption) (*filestorepb.Backup, error) {
	panic("implement me")
}

func (s *store) ListFilestoreBackups(ctx context.Context, req *filestorepb.ListBackupsRequest, _ ...gax.CallOption) gcpclient.Iterator[*filestorepb.Backup] {
	panic("implement me")
}

func (s *store) CreateFilestoreBackup(ctx context.Context, req *filestorepb.CreateBackupRequest, _ ...gax.CallOption) (gcpclient.WaitableOperationWithResult[*filestorepb.Backup], error) {
	panic("implement me")
}

func (s *store) DeleteFilestoreBackup(ctx context.Context, req *filestorepb.DeleteBackupRequest, _ ...gax.CallOption) (gcpclient.WaitableVoidOperation, error) {
	panic("implement me")
}
