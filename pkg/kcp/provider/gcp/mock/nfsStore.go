package mock

import (
	"context"
	"google.golang.org/api/file/v1"
)

type nfsStore struct {
}

func (s *nfsStore) GetFilestoreInstance(ctx context.Context, projectId, location, instanceId string) (*file.Instance, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}

	return nil, nil
}
func (s *nfsStore) CreateFilestoreInstance(ctx context.Context, projectId, location, instanceId string, instance *file.Instance) (*file.Operation, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}

	return nil, nil
}
func (s *nfsStore) DeleteFilestoreInstance(ctx context.Context, projectId, location, instanceId string) (*file.Operation, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}

	return nil, nil
}
func (s *nfsStore) GetFilestoreOperation(ctx context.Context, projectId, operationName string) (*file.Operation, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}

	return nil, nil
}
func (s *nfsStore) PatchFilestoreInstance(ctx context.Context, projectId, location, instanceId, updateMask string, instance *file.Instance) (*file.Operation, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}

	return nil, nil
}
