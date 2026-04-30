package mock2

import (
	"context"
	"errors"
	"fmt"

	"cloud.google.com/go/filestore/apiv1/filestorepb"
	"cloud.google.com/go/longrunning/autogen/longrunningpb"
	"github.com/kyma-project/cloud-manager/pkg/common"
	gcpmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/meta"
	gcputil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
	commonpb "google.golang.org/genproto/googleapis/cloud/common"
)

type FileStoreRestoreOperationsConfig interface {
	ResolveFilestoreRestoreOperation(ctx context.Context, operationName string, opts ...FilestoreRestoreOperationOptionCall) error
}

type FilestoreRestoreOperationOptionCall func(fs *filestorepb.Instance, opBuilder *OperationLongRunningBuilder) error

func WithFilestoreRestoreOperationSimpleError(code int32, message string) FilestoreRestoreOperationOptionCall {
	return func(fs *filestorepb.Instance, opBuilder *OperationLongRunningBuilder) error {
		fs.State = filestorepb.Instance_ERROR
		opBuilder.WithSimpleError(code, message)
		return nil
	}
}

func WithFilestoreRestoreOperationError(err *longrunningpb.Operation_Error) FilestoreRestoreOperationOptionCall {
	return func(fs *filestorepb.Instance, opBuilder *OperationLongRunningBuilder) error {
		fs.State = filestorepb.Instance_ERROR
		opBuilder.WithOperationError(err)
		return nil
	}
}

func (s *store) ResolveFilestoreRestoreOperation(ctx context.Context, operationName string, opts ...FilestoreRestoreOperationOptionCall) error {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return ctx.Err()
	}

	opName, err := gcputil.ParseNameDetail(operationName)
	if err != nil {
		return gcpmeta.NewBadRequestError("operation name is invalid: %v", err)
	}
	opBuilder, found := s.longRunningOperations.FindByName(opName)
	if !found {
		return gcpmeta.NewNotFoundError("operation %s not found", opName)
	}
	fs, found := s.filestores.FindByName(opBuilder.relatedItemName)
	if !found {
		return gcpmeta.NewNotFoundError("operation target filestore %s not found", opBuilder.relatedItemName)
	}
	fs.State = filestorepb.Instance_READY
	opBuilder.WithDone(true)
	for _, op := range opts {
		if err := op(fs, opBuilder); err != nil {
			if !errors.Is(err, common.ErrLogical) {
				err = fmt.Errorf("%w: %w", common.ErrLogical, err)
			}
			return err
		}
	}
	if err := opBuilder.WithResult(fs); err != nil {
		return err
	}
	meta, err := ReadOperationMetadata[*commonpb.OperationMetadata](opBuilder)
	if err != nil {
		return err
	}
	_ = meta
	return nil
}
