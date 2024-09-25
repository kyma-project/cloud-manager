package mock

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/serviceusage/v1"
)

type serviceUsageStore struct {
	services         map[string]bool
	suEnableError    *googleapi.Error
	suDisableError   *googleapi.Error
	suOperationError *googleapi.Error
	suIsEnabledError *googleapi.Error
}

func initServicesMap(sus *serviceUsageStore) {
	if sus.services == nil {
		sus.services = make(map[string]bool)
	}
}

func newServiceUsageOperation(msg string, done bool) *serviceusage.Operation {
	name := uuid.New().String()
	if msg != "" {
		return &serviceusage.Operation{Name: name, Error: &serviceusage.Status{Code: 500, Message: msg}, Done: done}
	}
	return &serviceusage.Operation{Name: name, Done: done}
}

func (s *serviceUsageStore) EnableService(ctx context.Context, projectId string, serviceName client.GcpServiceName) (*serviceusage.Operation, error) {
	initServicesMap(s)
	if s.suEnableError != nil {
		return nil, s.suEnableError
	}
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}

	logger := composed.LoggerFromCtx(ctx)

	completeId := fmt.Sprintf("projects/%s/services/%s", projectId, serviceName)
	s.services[completeId] = true
	logger.WithName("EnableService - mock").Info(fmt.Sprintf("Length :: %d", len(s.services)))

	return newServiceUsageOperation("", false), nil
}

func (s *serviceUsageStore) DisableService(ctx context.Context, projectId string, serviceName client.GcpServiceName) (*serviceusage.Operation, error) {
	initServicesMap(s)
	if s.suDisableError != nil {
		return nil, s.suDisableError
	}
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}

	logger := composed.LoggerFromCtx(ctx)
	completeId := fmt.Sprintf("projects/%s/services/%s", projectId, serviceName)
	s.services[completeId] = false
	logger.WithName("DisableService - mock").Info(fmt.Sprintf("Length :: %d", len(s.services)))

	return newServiceUsageOperation("", false), nil
}

func (s *serviceUsageStore) IsServiceEnabled(ctx context.Context, projectId string, serviceName client.GcpServiceName) (bool, error) {
	initServicesMap(s)
	if s.suIsEnabledError != nil {
		return false, s.suIsEnabledError
	}
	if isContextCanceled(ctx) {
		return false, context.Canceled
	}
	completeId := fmt.Sprintf("projects/%s/services/%s", projectId, serviceName)
	enabled := s.services[completeId]
	logger := composed.LoggerFromCtx(ctx)
	logger.WithName("IsServiceEnabled - mock").Info(fmt.Sprintf("service :: %s , enabled :: %t", completeId, enabled))
	return enabled, nil
}

func (s *serviceUsageStore) GetServiceUsageOperation(ctx context.Context, operationName string) (*serviceusage.Operation, error) {
	if s.suOperationError != nil {
		return nil, s.suOperationError
	}
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	logger := composed.LoggerFromCtx(ctx)
	logger.WithName("GetServiceUsageOperation - mock").Info(fmt.Sprintf("operationName :: %s", operationName))
	return newServiceUsageOperation("", true), nil
}
