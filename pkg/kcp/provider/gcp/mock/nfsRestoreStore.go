package mock

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"google.golang.org/api/file/v1"
	"google.golang.org/api/googleapi"
)

type nfsRestoreStore struct {
	restoreFileError      error
	restoreOperationError *googleapi.Error
}

func (s *nfsRestoreStore) RestoreFile(ctx context.Context, projectId, destFileFullPath, destFileShareName, srcBackupFullPath string) (*file.Operation, error) {
	if s.restoreFileError != nil {
		return nil, s.restoreFileError
	}
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	logger := composed.LoggerFromCtx(ctx)
	logger.WithName("GetFilestoreInstance - mock").Info("Restoring file ", "projectId =", projectId, "destFileFullPath =", destFileFullPath, "destFileShareName =", destFileShareName, "srcBackupFullPath =", srcBackupFullPath)
	return newOperation("", false), nil
}

func (s *nfsRestoreStore) GetRestoreOperation(ctx context.Context, _, operationName string) (*file.Operation, error) {
	if s.restoreOperationError != nil {
		return nil, s.restoreOperationError
	}
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	return &file.Operation{Name: operationName, Done: true}, nil
}
