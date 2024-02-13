package mock

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"google.golang.org/api/file/v1"
	"google.golang.org/api/googleapi"
	"strings"
)

type nfsStore struct {
	instances []*file.Instance
}

func (s *nfsStore) GetFilestoreInstance(ctx context.Context, projectId, location, instanceId string) (*file.Instance, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	completeId := fmt.Sprintf("projects/%s/locations/%s/instances/%s", projectId, location, instanceId)
	logger := composed.LoggerFromCtx(ctx)
	for i, instance := range s.instances {
		if s.instances[i].Name == completeId {
			logger.WithName("GetFilestoreInstance - mock").Info("Got Instance. ", "instance =", instance)

			result := s.instances[i]
			result.State = "READY"
			return result, nil
		}
	}
	logger.WithName("GetFilestoreInstance - mock").Info(fmt.Sprintf("Length :: %d", len(s.instances)))
	return nil, &googleapi.Error{
		Code:    404,
		Message: "Not able to find the instance",
	}
}
func (s *nfsStore) CreateFilestoreInstance(ctx context.Context, projectId, location, instanceId string, instance *file.Instance) (*file.Operation, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}

	logger := composed.LoggerFromCtx(ctx)

	completeId := fmt.Sprintf("projects/%s/locations/%s/instances/%s", projectId, location, instanceId)
	instance.Name = completeId
	s.instances = append(s.instances, instance)
	logger.WithName("CreateFilestoreInstance - mock").Info(fmt.Sprintf("Length :: %d", len(s.instances)))

	return newOperation("", false), nil
}
func (s *nfsStore) DeleteFilestoreInstance(ctx context.Context, projectId, location, instanceId string) (*file.Operation, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}

	logger := composed.LoggerFromCtx(ctx)

	completeId := fmt.Sprintf("projects/%s/locations/%s/instances/%s", projectId, location, instanceId)
	for i, instance := range s.instances {
		if completeId == instance.Name {
			s.instances = append(s.instances[:i], s.instances[i+1:]...)
			break
		}
	}

	logger.WithName("DeleteFilestoreInstance - mock").Info(fmt.Sprintf("Length :: %d", len(s.instances)))
	return newOperation("", false), nil
}

func (s *nfsStore) GetFilestoreOperation(ctx context.Context, _, operationName string) (*file.Operation, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}

	return &file.Operation{Name: operationName, Done: true}, nil
}
func (s *nfsStore) PatchFilestoreInstance(ctx context.Context, projectId, location, instanceId, updateMask string, instance *file.Instance) (*file.Operation, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}

	logger := composed.LoggerFromCtx(ctx)
	logger.WithName("PatchFilestoreInstance - mock").Info(fmt.Sprintf("Length :: %d", len(s.instances)))

	completeId := fmt.Sprintf("projects/%s/locations/%s/instances/%s", projectId, location, instanceId)
	for i, item := range s.instances {
		if item != nil && item.Name == completeId {
			for _, field := range strings.Split(updateMask, ",") {
				switch field {
				case "fileShares":
					s.instances[i].FileShares = instance.FileShares
				case "description":
					s.instances[i].Description = instance.Description
				case "labels":
					s.instances[i].Labels = instance.Labels
				default:
					return nil, &googleapi.Error{
						Code:    400,
						Message: "update_mask is not valid.",
					}
				}
			}
			return newOperation("", false), nil
		}
	}
	return nil, &googleapi.Error{
		Code:    404,
		Message: fmt.Sprintf("Resource %s was not found", completeId),
	}
}

func newOperation(error string, done bool) *file.Operation {
	name := uuid.New().String()
	if error != "" {
		return &file.Operation{Name: name, Error: &file.Status{Code: 500, Message: error}, Done: done}
	}
	return &file.Operation{Name: name, Done: done}
}
