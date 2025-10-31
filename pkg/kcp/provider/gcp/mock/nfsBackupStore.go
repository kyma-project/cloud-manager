package mock

import (
	"context"
	"fmt"
	"regexp"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"google.golang.org/api/file/v1"
	"google.golang.org/api/googleapi"
)

type FileBackupClientFakeUtils interface {
	CreateFakeBackup(backup *file.Backup)
	ClearAllBackups()
}

type nfsBackupStore struct {
	backups              []*file.Backup
	backupOperationError *googleapi.Error
}

func (s *nfsBackupStore) CreateFakeBackup(backup *file.Backup) {
	s.backups = append(s.backups, backup)
}

func (s *nfsBackupStore) ClearAllBackups() {
	s.backups = []*file.Backup{}
}

func (s *nfsBackupStore) GetFileBackup(ctx context.Context, projectId, location, name string) (*file.Backup, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	completeId := fmt.Sprintf("projects/%s/locations/%s/backups/%s", projectId, location, name)
	logger := composed.LoggerFromCtx(ctx)
	for i, backup := range s.backups {
		if s.backups[i].Name == completeId {
			logger.WithName("GetFileBackup - mock").Info("Got Nfs Backup. ", "backup =", backup)

			result := s.backups[i]
			result.State = "READY"
			return result, nil
		}
	}
	logger.WithName("GetFileBackup - mock").Info(fmt.Sprintf("Length :: %d", len(s.backups)))
	return nil, &googleapi.Error{
		Code:    404,
		Message: "Not able to find the backup",
	}
}

func (s *nfsBackupStore) ListFilesBackups(ctx context.Context, project, filter string) ([]*file.Backup, error) {
	regex := `.+labels.scope-name="(?P<Scope>.+)"`
	re := regexp.MustCompile(regex)
	matches := re.FindStringSubmatch(filter)
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	if len(matches) == 0 {
		return s.backups, nil
	}
	result := make([]*file.Backup, 0)
	for _, backup := range s.backups {
		projectId, _, _ := client.GetProjectLocationNameFromFileBackupPath(backup.Name)
		scopeName := matches[re.SubexpIndex("Scope")]
		if projectId == project && backup.Labels != nil && backup.Labels["scope-name"] == scopeName {
			result = append(result, backup)
		}
	}
	return result, nil
}

func (s *nfsBackupStore) PatchFileBackup(ctx context.Context, projectId, location, name, _ string, backup *file.Backup) (*file.Operation, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}

	logger := composed.LoggerFromCtx(ctx)

	completeId := fmt.Sprintf("projects/%s/locations/%s/backups/%s", projectId, location, name)
	for i, instance := range s.backups {
		if completeId == instance.Name {
			s.backups[i] = backup
			logger.WithName("PatchFileBackup - mock").Info(fmt.Sprintf("Length :: %d", len(s.backups)))
			return newOperation("", false), nil
		}
	}
	return nil, &googleapi.Error{
		Code:    404,
		Message: "Resource not found",
	}
}

func (s *nfsBackupStore) CreateFileBackup(ctx context.Context, projectId, location, name string, backup *file.Backup) (*file.Operation, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}

	logger := composed.LoggerFromCtx(ctx)

	completeId := fmt.Sprintf("projects/%s/locations/%s/backups/%s", projectId, location, name)
	backup.Name = completeId
	for _, existing := range s.backups {
		if existing.Name == completeId {
			return nil, &googleapi.Error{
				Code:    409,
				Message: "Resource already exists",
			}
		}
	}
	s.backups = append(s.backups, backup)
	logger.WithName("CreateFileBackup - mock").Info(fmt.Sprintf("Length :: %d", len(s.backups)))

	return newOperation("", false), nil
}
func (s *nfsBackupStore) DeleteFileBackup(ctx context.Context, projectId, location, name string) (*file.Operation, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}

	logger := composed.LoggerFromCtx(ctx)

	completeId := fmt.Sprintf("projects/%s/locations/%s/backups/%s", projectId, location, name)
	for i, instance := range s.backups {
		if completeId == instance.Name {
			s.backups = append(s.backups[:i], s.backups[i+1:]...)
			logger.WithName("DeleteFileBackup - mock").Info(fmt.Sprintf("Length :: %d", len(s.backups)))
			return newOperation("", false), nil
		}
	}
	return nil, &googleapi.Error{
		Code:    404,
		Message: "Resource not found",
	}

}

func (s *nfsBackupStore) GetBackupOperation(ctx context.Context, _, operationName string) (*file.Operation, error) {
	if s.backupOperationError != nil {
		return nil, s.backupOperationError
	}
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	return &file.Operation{Name: operationName, Done: true}, nil
}
