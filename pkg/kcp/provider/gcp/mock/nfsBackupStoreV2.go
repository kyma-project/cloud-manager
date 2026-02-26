package mock

import (
	"context"
	"fmt"
	"regexp"

	"cloud.google.com/go/filestore/apiv1/filestorepb"
	"cloud.google.com/go/longrunning/autogen/longrunningpb"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	gcpnfsbackupclientv2 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsbackup/client/v2"
	"google.golang.org/api/googleapi"
)

var (
	scopeNameRegexV2 = regexp.MustCompile(`.+labels\.scope-name="([^"]+)"`)
	cmAllowRegexV2   = regexp.MustCompile(`labels\.cm-allow-([^=\s"]+)`)
)

// FileBackupClientFakeUtilsV2 provides utility methods for v2 mock testing
type FileBackupClientFakeUtilsV2 interface {
	CreateFakeBackupV2(backup *filestorepb.Backup)
	ClearAllBackupsV2()
	GetNfsBackupV2ByName(name string) *filestorepb.Backup
	SetNfsBackupV2State(name string, state filestorepb.Backup_State)
	DeleteNfsBackupV2ByName(name string)
}

// nfsBackupStoreV2 implements FileBackupClient using protobuf types
type nfsBackupStoreV2 struct {
	backups              []*filestorepb.Backup
	backupOperationError *googleapi.Error
}

var _ gcpnfsbackupclientv2.FileBackupClient = &nfsBackupStoreV2{}

func (s *nfsBackupStoreV2) CreateFakeBackupV2(backup *filestorepb.Backup) {
	s.backups = append(s.backups, backup)
}

func (s *nfsBackupStoreV2) ClearAllBackupsV2() {
	s.backups = []*filestorepb.Backup{}
}

func (s *nfsBackupStoreV2) GetNfsBackupV2ByName(name string) *filestorepb.Backup {
	for _, backup := range s.backups {
		if backup.Name == name {
			return backup
		}
	}
	return nil
}

func (s *nfsBackupStoreV2) SetNfsBackupV2State(name string, state filestorepb.Backup_State) {
	for _, backup := range s.backups {
		if backup.Name == name {
			backup.State = state
			return
		}
	}
}

func (s *nfsBackupStoreV2) DeleteNfsBackupV2ByName(name string) {
	for i, backup := range s.backups {
		if backup.Name == name {
			s.backups = append(s.backups[:i], s.backups[i+1:]...)
			return
		}
	}
}

func (s *nfsBackupStoreV2) GetBackup(ctx context.Context, projectId, location, name string) (*filestorepb.Backup, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}

	completeName := fmt.Sprintf("projects/%s/locations/%s/backups/%s", projectId, location, name)
	logger := composed.LoggerFromCtx(ctx)

	for _, backup := range s.backups {
		if backup.Name == completeName {
			logger.WithName("GetBackup - mock v2").Info("Got Nfs Backup", "backup", backup.Name, "state", backup.State.String())
			return backup, nil
		}
	}

	logger.WithName("GetBackup - mock v2").Info(fmt.Sprintf("Backup not found, total: %d", len(s.backups)))
	return nil, &googleapi.Error{
		Code:    404,
		Message: "Not able to find the backup",
	}
}

func (s *nfsBackupStoreV2) ListBackups(ctx context.Context, projectId, filter string) ([]*filestorepb.Backup, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}

	scopeMatches := scopeNameRegexV2.FindStringSubmatch(filter)
	cmAllowMatches := cmAllowRegexV2.FindStringSubmatch(filter)

	if len(scopeMatches) == 0 && len(cmAllowMatches) == 0 {
		return []*filestorepb.Backup{}, nil
	}

	result := make([]*filestorepb.Backup, 0)

	for _, backup := range s.backups {
		projectIdFromPath, _, _ := gcpnfsbackupclientv2.GetProjectLocationNameFromFileBackupPath(backup.Name)

		// early continue if project does not match
		if projectIdFromPath != projectId {
			continue
		}

		// early continue if no labels
		if backup.Labels == nil {
			continue
		}

		// match scope-name filter (used for SKR backups)
		if len(scopeMatches) > 0 {
			scopeName := scopeMatches[1]
			if backup.Labels["scope-name"] == scopeName {
				result = append(result, backup)
			}
			continue
		}

		// match cm-allow- filter (used for shared backups discovery)
		if len(cmAllowMatches) > 0 {
			shootName := cmAllowMatches[1]
			allowLabel := fmt.Sprintf("cm-allow-%s", shootName)
			if _, ok := backup.Labels[allowLabel]; ok {
				result = append(result, backup)
			}
		}
	}

	return result, nil
}

func (s *nfsBackupStoreV2) CreateBackup(ctx context.Context, projectId, location, name string, backup *filestorepb.Backup) (string, error) {
	if isContextCanceled(ctx) {
		return "", context.Canceled
	}

	logger := composed.LoggerFromCtx(ctx)

	completeName := fmt.Sprintf("projects/%s/locations/%s/backups/%s", projectId, location, name)
	backup.Name = completeName

	for _, existing := range s.backups {
		if existing.Name == completeName {
			return "", &googleapi.Error{
				Code:    409,
				Message: "Resource already exists",
			}
		}
	}

	backup.State = filestorepb.Backup_CREATING
	s.backups = append(s.backups, backup)
	logger.WithName("CreateBackup - mock v2").Info(fmt.Sprintf("Created backup in CREATING state, total: %d", len(s.backups)))

	return fmt.Sprintf("operations/create-%s", name), nil
}

func (s *nfsBackupStoreV2) DeleteBackup(ctx context.Context, projectId, location, name string) (string, error) {
	if isContextCanceled(ctx) {
		return "", context.Canceled
	}

	logger := composed.LoggerFromCtx(ctx)

	completeName := fmt.Sprintf("projects/%s/locations/%s/backups/%s", projectId, location, name)
	for _, backup := range s.backups {
		if backup.Name == completeName {
			backup.State = filestorepb.Backup_DELETING
			logger.WithName("DeleteBackup - mock v2").Info("Set backup state to DELETING")
			return fmt.Sprintf("operations/delete-%s", name), nil
		}
	}

	return "", &googleapi.Error{
		Code:    404,
		Message: "Resource not found",
	}
}

func (s *nfsBackupStoreV2) GetBackupLROperation(ctx context.Context, operationName string) (*longrunningpb.Operation, error) {
	if s.backupOperationError != nil {
		return nil, s.backupOperationError
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

func (s *nfsBackupStoreV2) UpdateBackup(ctx context.Context, projectId, location, name string, backup *filestorepb.Backup, updateMask []string) (string, error) {
	if isContextCanceled(ctx) {
		return "", context.Canceled
	}

	logger := composed.LoggerFromCtx(ctx)

	completeName := fmt.Sprintf("projects/%s/locations/%s/backups/%s", projectId, location, name)
	for i, existing := range s.backups {
		if existing.Name == completeName {
			// Apply update mask fields
			// For simplicity in mock, we replace the entire backup but preserve the name
			backup.Name = completeName
			s.backups[i] = backup
			logger.WithName("UpdateBackup - mock v2").Info("Updated backup")
			return fmt.Sprintf("operations/update-%s", name), nil
		}
	}

	return "", &googleapi.Error{
		Code:    404,
		Message: "Resource not found",
	}
}
