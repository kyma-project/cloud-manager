package mock

import (
	"context"
	"fmt"
	"sync"

	"cloud.google.com/go/longrunning/autogen/longrunningpb"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	gcpnfsrestoreclientv2 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsrestore/client/v2"
	"google.golang.org/api/googleapi"
)

// FileRestoreClientFakeUtilsV2 provides utility methods for v2 mock testing
type FileRestoreClientFakeUtilsV2 interface {
	SetRestoreFileErrorV2(err error)
	SetRestoreOperationErrorV2(err *googleapi.Error)
	SetRestoreOperationDoneV2(opName string)
}

// nfsRestoreStoreV2 implements FileRestoreClient using protobuf types
type nfsRestoreStoreV2 struct {
	m                     sync.Mutex
	restoreFileError      error
	restoreOperationError *googleapi.Error
	operations            map[string]*longrunningpb.Operation
	opCounter             int
}

var _ gcpnfsrestoreclientv2.FileRestoreClient = &nfsRestoreStoreV2{}

func (s *nfsRestoreStoreV2) SetRestoreFileErrorV2(err error) {
	s.m.Lock()
	defer s.m.Unlock()
	s.restoreFileError = err
}

func (s *nfsRestoreStoreV2) SetRestoreOperationErrorV2(err *googleapi.Error) {
	s.m.Lock()
	defer s.m.Unlock()
	s.restoreOperationError = err
}

// SetRestoreOperationDoneV2 marks the operation with the given name as done.
// This must be called explicitly in tests to simulate GCP completing the restore.
func (s *nfsRestoreStoreV2) SetRestoreOperationDoneV2(opName string) {
	s.m.Lock()
	defer s.m.Unlock()
	if op, ok := s.operations[opName]; ok {
		op.Done = true
	}
}

func (s *nfsRestoreStoreV2) RestoreFile(ctx context.Context, projectId, destFileFullPath, destFileShareName, srcBackupFullPath string) (string, error) {
	s.m.Lock()
	defer s.m.Unlock()
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

	s.opCounter++
	opName := fmt.Sprintf("operations/mock-restore-op-%d", s.opCounter)
	if s.operations == nil {
		s.operations = make(map[string]*longrunningpb.Operation)
	}
	s.operations[opName] = &longrunningpb.Operation{
		Name: opName,
		Done: false,
	}
	return opName, nil
}

func (s *nfsRestoreStoreV2) GetRestoreOperation(ctx context.Context, operationName string) (*longrunningpb.Operation, error) {
	s.m.Lock()
	defer s.m.Unlock()
	if s.restoreOperationError != nil {
		return nil, s.restoreOperationError
	}
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	if op, ok := s.operations[operationName]; ok {
		return op, nil
	}
	// Operation not found in store — return done to be safe for unknown ops
	return &longrunningpb.Operation{
		Name: operationName,
		Done: true,
	}, nil
}

func (s *nfsRestoreStoreV2) FindRestoreOperation(ctx context.Context, projectId, location, instanceId string) (*longrunningpb.Operation, error) {
	s.m.Lock()
	defer s.m.Unlock()
	if s.restoreOperationError != nil {
		return nil, s.restoreOperationError
	}
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	// Return the first non-done operation if any
	for _, op := range s.operations {
		if !op.Done {
			return op, nil
		}
	}
	return nil, nil
}
