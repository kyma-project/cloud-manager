package client

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"net/http"
	"strconv"

	"github.com/kyma-project/cloud-manager/pkg/composed"

	"google.golang.org/api/option"
	"google.golang.org/api/servicenetworking/v1"
)

type ServiceNetworkingClient interface {
	ListServiceConnections(ctx context.Context, projectId, vpcId string) ([]*servicenetworking.Connection, error)
	CreateServiceConnection(ctx context.Context, projectId, vpcId string, reservedIpRanges []string) (*servicenetworking.Operation, error)
	// DeleteServiceConnection: Deletes a private service access connection.
	// projectNumber: Project number which is different from project id. Get it by calling client.GetProjectNumber(ctx, projectId)
	DeleteServiceConnection(ctx context.Context, projectId, vpcId string) (*servicenetworking.Operation, error)
	PatchServiceConnection(ctx context.Context, projectId, vpcId string, reservedIpRanges []string) (*servicenetworking.Operation, error)
	GetServiceNetworkingOperation(ctx context.Context, operationName string) (*servicenetworking.Operation, error)
}

func NewServiceNetworkingClient() client.ClientProvider[ServiceNetworkingClient] {
	return client.NewCachedClientProvider(
		func(ctx context.Context, saJsonKeyPath string) (ServiceNetworkingClient, error) {
			httpClient, err := client.GetCachedGcpClient(ctx, saJsonKeyPath)
			if err != nil {
				return nil, err
			}
			svcNetClient, err := servicenetworking.NewService(ctx, option.WithHTTPClient(httpClient))
			if err != nil {
				return nil, err
			}
			return newServiceNetworkingClient(svcNetClient, httpClient), nil
		},
	)
}

func newServiceNetworkingClient(svcNet *servicenetworking.APIService, httpClient *http.Client) ServiceNetworkingClient {
	return &serviceNetworkingClient{svcNet: svcNet, httpClient: httpClient}
}

type serviceNetworkingClient struct {
	svcNet     *servicenetworking.APIService
	httpClient *http.Client
}

func (c *serviceNetworkingClient) PatchServiceConnection(ctx context.Context, projectId, vpcId string, reservedIpRanges []string) (*servicenetworking.Operation, error) {
	logger := composed.LoggerFromCtx(ctx)
	network := client.GetVPCPath(projectId, vpcId)
	operation, err := c.svcNet.Services.Connections.Patch(client.ServiceNetworkingServiceConnectionName, &servicenetworking.Connection{
		Network:               network,
		ReservedPeeringRanges: reservedIpRanges,
	}).Force(true).Do()
	logger.V(4).Info("PatchServiceConnection", "operation", operation, "err", err)
	return operation, err
}

func (c *serviceNetworkingClient) DeleteServiceConnection(ctx context.Context, projectId, vpcId string) (*servicenetworking.Operation, error) {
	logger := composed.LoggerFromCtx(ctx)
	ProjectNumber, err := client.GetCachedProjectNumber(ctx, projectId, c.httpClient)
	if err != nil {
		return nil, err
	}
	network := client.GetVPCPath(strconv.FormatInt(ProjectNumber, 10), vpcId)
	operation, err := c.svcNet.Services.Connections.DeleteConnection(client.ServiceNetworkingServiceConnectionName, &servicenetworking.DeleteConnectionRequest{
		ConsumerNetwork: network,
	}).Do()
	logger.V(4).Info("DeleteServiceConnection", "operation", operation, "err", err)
	return operation, err
}

func (c *serviceNetworkingClient) ListServiceConnections(ctx context.Context, projectId, vpcId string) ([]*servicenetworking.Connection, error) {
	logger := composed.LoggerFromCtx(ctx)
	network := client.GetVPCPath(projectId, vpcId)
	out, err := c.svcNet.Services.Connections.List(client.ServiceNetworkingServicePath).Network(network).Do()
	if err != nil {
		logger.Error(err, "ListServiceConnections", "projectId", projectId, "vpcId", vpcId)
		return nil, err
	}
	return out.Connections, nil
}

func (c *serviceNetworkingClient) CreateServiceConnection(ctx context.Context, projectId, vpcId string, reservedIpRanges []string) (*servicenetworking.Operation, error) {
	logger := composed.LoggerFromCtx(ctx)
	network := client.GetVPCPath(projectId, vpcId)
	operation, err := c.svcNet.Services.Connections.Create(client.ServiceNetworkingServicePath, &servicenetworking.Connection{
		Network:               network,
		ReservedPeeringRanges: reservedIpRanges,
	}).Do()
	logger.V(4).Info("CreateServiceConnection", "operation", operation, "err", err)
	return operation, err
}

func (c *serviceNetworkingClient) GetServiceNetworkingOperation(ctx context.Context, operationName string) (*servicenetworking.Operation, error) {
	logger := composed.LoggerFromCtx(ctx)
	operation, err := c.svcNet.Operations.Get(operationName).Do()
	if err != nil {
		logger.Error(err, "GetServiceNetworkingOperation", "operationName", operationName)
		return nil, err
	}
	return operation, nil
}
