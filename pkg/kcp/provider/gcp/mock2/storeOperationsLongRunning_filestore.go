package mock2

import (
	"context"
	"errors"
	"fmt"

	"cloud.google.com/go/filestore/apiv1/filestorepb"
	"cloud.google.com/go/longrunning/autogen/longrunningpb"
	"github.com/googleapis/gax-go/v2"
	"github.com/kyma-project/cloud-manager/pkg/common"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/meta"
	gcputil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
	commonpb "google.golang.org/genproto/googleapis/cloud/common"
)

type FileStoreOperationsConfig interface {
	ResolveFilestoreOperation(ctx context.Context, operationName string, opts ...FilestoreOperationOptionCall) error
}

// Filestore operation resolving

type FilestoreOperationOptionCall func(fs *filestorepb.Instance, opBuilder *OperationLongRunningBuilder) error

func WithFilestoreOperationSimpleError(code int32, message string) FilestoreOperationOptionCall {
	return func(fs *filestorepb.Instance, opBuilder *OperationLongRunningBuilder) error {
		fs.State = filestorepb.Instance_ERROR
		opBuilder.WithSimpleError(code, message)
		return nil
	}
}

func WithFilestoreOperationError(err *longrunningpb.Operation_Error) FilestoreOperationOptionCall {
	return func(fs *filestorepb.Instance, opBuilder *OperationLongRunningBuilder) error {
		fs.State = filestorepb.Instance_ERROR
		opBuilder.WithOperationError(err)
		return nil
	}
}

func (s *store) ResolveFilestoreOperation(ctx context.Context, operationName string, opts ...FilestoreOperationOptionCall) error {
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
	if meta != nil && meta.Verb == "delete" {
		s.filestores = s.filestores.FilterNotByCallback(func(item FilterableListItem[*filestorepb.Instance]) bool {
			return item.Name.Equal(opBuilder.relatedItemName)
		})
	}
	return nil
}

// Client Interface implementations ===================================================

func (s *store) GetFilestoreOperation(ctx context.Context, req *longrunningpb.GetOperationRequest, _ ...gax.CallOption) (*longrunningpb.Operation, error) {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return nil, ctx.Err()
	}

	return s.getLongRunningOperationNoLock(req.Name)
}

func (s *store) ListFilestoreOperations(ctx context.Context, req *longrunningpb.ListOperationsRequest, _ ...gax.CallOption) gcpclient.Iterator[*longrunningpb.Operation] {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return &iteratorMocked[*longrunningpb.Operation]{
			err: ctx.Err(),
		}
	}

	return s.listLongRunningOperationsNoLock(req.Name, req.Filter)
}
