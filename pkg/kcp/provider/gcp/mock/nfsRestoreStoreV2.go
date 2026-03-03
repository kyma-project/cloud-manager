package mock

import (
	"context"

	"cloud.google.com/go/longrunning/autogen/longrunningpb"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	gcpnfsrestoreclientv2 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsrestore/client/v2"
	"google.golang.org/api/googleapi"
)

// FileRestoreClientFakeUtilsV2 provides utility methods for v2 mock testing
type FileRestoreClientFakeUtilsV2 interface {
	SetRestoreFileErrorV2(err error)
	SetRestoreOperationErrorV2(err *googleapi.Error)
}

// nfsRestoreStoreV2 implements FileRestoreClient using protobuf types
type nfsRestoreStoreV2 struct {
	restoreFileError      error
	restoreOperationError *googleapi.Error
}

var _ gcpnfsrestoreclientv2.FileRestoreClient = &nfsRestoreStoreV2{}

func (s *nfsRestoreStoreV2) SetRestoreFileErrorV2(err error) {
	s.restoreFileError = err
}

func (s *nfsRestoreStoreV2) SetRestoreOperationErrorV2(err *googleapi.Error) {
	s.restoreOperationError = err
}

func (s *nfsRestoreStoreV2) RestoreFile(ctx context.Context, projectId, destFileFullPath, destFileShareName, srcBackupFullPath string) (string, error) {
	if s.restoreFileError != nil {
		return "", s.restoreFileError
	}
	if isContextCanceled(ctx) {
		return "", context.Canceled
	}
	logger := composed.LoggerFromCtx(ctx)
	logger.WithName("RestoreFile - mock v2").Info("Restoring file",
		"projectId", projectId,
		"destFileFullPath", destFileFullPath,
		"destFileShareName", destFileShareName,
		"srcBackupFullPath", srcBackupFullPath)
	return "operations/mock-restore-op", nil
}

func (s *nfsRestoreStoreV2) GetRestoreOperation(ctx context.Context, operationName string) (*longrunningpb.Operation, error) {
	if s.restoreOperationError != nil {
		return nil, s.restoreOperationError
	}
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	// Mock always returns done with no error
	return &longrunningpb.Operation{
		Name: operationName,
		Done: true,
	}, nil
}

func (s *nfsRestoreStoreV2) FindRestoreOperation(ctx context.Context, projectId, location, instanceId string) (*longrunningpb.Operation, error) {
	if s.restoreOperationError != nil {
		return nil, s.restoreOperationError
	}
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	// Mock returns nil to indicate no running operation found
	return nil, nil
}
