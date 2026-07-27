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

type FileStoreBackupOperationsConfig interface {
	ResolveFilestoreBackupOperation(ctx context.Context, operationName string, opts ...FilestoreBackupOperationOptionCall) error
	ResolvePendingBackupDeleteOperations(ctx context.Context) error
	AddFilestoreBackupDirectly(backup *filestorepb.Backup) error
	SetListFilestoreBackupsError(err error)
}

// Filestore operation resolving

type FilestoreBackupOperationOptionCall func(fs *filestorepb.Backup, opBuilder *OperationLongRunningBuilder) error

func WithFilestoreBackupOperationSimpleError(code int32, message string) FilestoreBackupOperationOptionCall {
	return func(backup *filestorepb.Backup, opBuilder *OperationLongRunningBuilder) error {
		backup.State = filestorepb.Backup_INVALID
		opBuilder.WithSimpleError(code, message)
		return nil
	}
}

func WithFilestoreBackupOperationError(err *longrunningpb.Operation_Error) FilestoreBackupOperationOptionCall {
	return func(backup *filestorepb.Backup, opBuilder *OperationLongRunningBuilder) error {
		backup.State = filestorepb.Backup_INVALID
		opBuilder.WithOperationError(err)
		return nil
	}
}

func (s *store) ResolveFilestoreBackupOperation(ctx context.Context, operationName string, opts ...FilestoreBackupOperationOptionCall) error {
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
	backup, found := s.backups.FindByName(opBuilder.relatedItemName)
	if !found {
		return gcpmeta.NewNotFoundError("operation target filestore backup %s not found", opBuilder.relatedItemName)
	}
	backup.State = filestorepb.Backup_READY
	opBuilder.WithDone(true)
	for _, op := range opts {
		if err := op(backup, opBuilder); err != nil {
			if !errors.Is(err, common.ErrLogical) {
				err = fmt.Errorf("%w: %w", common.ErrLogical, err)
			}
			return err
		}
	}
	if err := opBuilder.WithResult(backup); err != nil {
		return err
	}
	meta, err := ReadOperationMetadata[*commonpb.OperationMetadata](opBuilder)
	if err != nil {
		return err
	}
	if meta != nil && meta.Verb == "delete" {
		s.backups = s.backups.FilterNotByCallback(func(item FilterableListItem[*filestorepb.Backup]) bool {
			return item.Name.Equal(opBuilder.relatedItemName)
		})
	}
	return nil
}

// ResolvePendingBackupDeleteOperations completes every not-yet-done backup delete operation,
// removing the corresponding backup from the store. It models GCP's asynchronous completion of
// delete operations for fire-and-forget callers (e.g. the KCP nuke reconciler) that do not track
// operation names. Tests drive it in an Eventually loop while the reconciler runs.
func (s *store) ResolvePendingBackupDeleteOperations(ctx context.Context) error {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return ctx.Err()
	}

	for _, opBuilder := range s.longRunningOperations.GetItems() {
		if opBuilder.GetOperationPB().Done {
			continue
		}
		meta, err := ReadOperationMetadata[*commonpb.OperationMetadata](opBuilder)
		if err != nil {
			return err
		}
		if meta == nil || meta.Verb != "delete" {
			continue
		}
		backupName := opBuilder.relatedItemName
		if _, found := s.backups.FindByName(backupName); !found {
			continue
		}
		opBuilder.WithDone(true)
		s.backups = s.backups.FilterNotByCallback(func(item FilterableListItem[*filestorepb.Backup]) bool {
			return item.Name.Equal(backupName)
		})
	}
	return nil
}
