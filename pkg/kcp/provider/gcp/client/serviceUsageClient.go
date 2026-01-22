package client

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"google.golang.org/api/option"
	"google.golang.org/api/serviceusage/v1"
)

type ServiceUsageClient interface {
	EnableService(ctx context.Context, projectId string, serviceName GcpServiceName) (*serviceusage.Operation, error)
	DisableService(ctx context.Context, projectId string, serviceName GcpServiceName) (*serviceusage.Operation, error)
	IsServiceEnabled(ctx context.Context, projectId string, serviceName GcpServiceName) (bool, error)
	GetServiceUsageOperation(ctx context.Context, operationName string) (*serviceusage.Operation, error)
}

func NewServiceUsageClientProvider() ClientProvider[ServiceUsageClient] {
	return NewCachedClientProvider(
		func(ctx context.Context, credentialsFile string) (ServiceUsageClient, error) {
			baseClient, err := GetCachedGcpClient(ctx, credentialsFile)
			if err != nil {
				return nil, err
			}

			httpClient := NewMetricsHTTPClient(baseClient.Transport)

			serviceUsageClient, err := serviceusage.NewService(ctx, option.WithHTTPClient(httpClient))
			if err != nil {
				return nil, fmt.Errorf("error obtaining GCP Compute Client: [%w]", err)
			}
			return NewServiceUsageClient(serviceUsageClient), nil
		},
	)
}

func NewServiceUsageClient(svcServiceUsage *serviceusage.Service) ServiceUsageClient {
	return &serviceUsageClient{svcServiceUsage: svcServiceUsage}
}

type serviceUsageClient struct {
	svcServiceUsage *serviceusage.Service
}

func (s serviceUsageClient) GetServiceUsageOperation(ctx context.Context, operationName string) (*serviceusage.Operation, error) {
	logger := composed.LoggerFromCtx(ctx)
	operation, err := s.svcServiceUsage.Operations.Get(operationName).Do()
	if err != nil {
		logger.Info("GetOperation", "err", err)
		return nil, err
	}
	return operation, err
}

func (s serviceUsageClient) EnableService(ctx context.Context, projectId string, serviceName GcpServiceName) (*serviceusage.Operation, error) {
	logger := composed.LoggerFromCtx(ctx)
	completeServiceName := GetCompleteServiceName(projectId, serviceName)
	enableServiceRequest := &serviceusage.EnableServiceRequest{}
	operation, err := s.svcServiceUsage.Services.Enable(completeServiceName, enableServiceRequest).Do()
	if err != nil {
		logger.Info("EnableService", "err", err)
	}
	return operation, err
}

func (s serviceUsageClient) DisableService(ctx context.Context, projectId string, serviceName GcpServiceName) (*serviceusage.Operation, error) {
	logger := composed.LoggerFromCtx(ctx)
	completeServiceName := GetCompleteServiceName(projectId, serviceName)
	disableServiceRequest := &serviceusage.DisableServiceRequest{}
	operation, err := s.svcServiceUsage.Services.Disable(completeServiceName, disableServiceRequest).Do()
	if err != nil {
		logger.Info("DisableService", "err", err)
	}
	return operation, err
}

func (s serviceUsageClient) IsServiceEnabled(ctx context.Context, projectId string, serviceName GcpServiceName) (bool, error) {
	logger := composed.LoggerFromCtx(ctx)
	completeServiceName := GetCompleteServiceName(projectId, serviceName)
	service, err := s.svcServiceUsage.Services.Get(completeServiceName).Do()
	if err != nil {
		logger.Info("isServiceEnabled", "err", err)
		return false, err
	}
	return service.State == "ENABLED", nil
}
