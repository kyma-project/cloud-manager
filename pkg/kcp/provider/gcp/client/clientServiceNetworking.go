package client

import (
	"context"
	"fmt"

	"github.com/hashicorp/golang-lru/v2"
	gcputil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/util"
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/servicenetworking/v1"
)

type ServiceNetworkingClient interface {
	PatchServiceConnection(ctx context.Context, projectId, vpcId string, reservedIpRanges []string) (*servicenetworking.Operation, error)
	DeleteServiceConnection(ctx context.Context, projectId, vpcId string) (*servicenetworking.Operation, error)
	ListServiceConnections(ctx context.Context, projectId, vpcId string) ([]*servicenetworking.Connection, error)
	CreateServiceConnection(ctx context.Context, projectId, vpcId string, reservedIpRanges []string) (*servicenetworking.Operation, error)

	GetServiceNetworkingOperation(_ context.Context, operationName string) (*servicenetworking.Operation, error)
}

var _ ServiceNetworkingClient = &serviceNetworkingClient{}

type serviceNetworkingClient struct {
	inner *servicenetworking.APIService
	crm   *cloudresourcemanager.Service
}

var projectNumbersCache, _ = lru.New[string, int64](256)

func (c *serviceNetworkingClient) PatchServiceConnection(_ context.Context, projectId, vpcId string, reservedIpRanges []string) (*servicenetworking.Operation, error) {
	networkName := gcputil.NewGlobalNetworkName(projectId, vpcId).String()
	return c.inner.Services.Connections.Patch(ServiceNetworkingServiceConnectionName, &servicenetworking.Connection{
		Network:               networkName,
		ReservedPeeringRanges: reservedIpRanges,
	}).Force(true).Do()
}

func (c *serviceNetworkingClient) DeleteServiceConnection(_ context.Context, projectId, vpcId string) (*servicenetworking.Operation, error) {
	projectNumber, ok := projectNumbersCache.Get(projectId)
	if !ok {
		project, err := c.crm.Projects.Get(projectId).Do()
		if err != nil {
			return nil, err
		}
		projectNumber = project.ProjectNumber
		projectNumbersCache.Add(projectId, projectNumber)
	}

	networkName := gcputil.NewGlobalNetworkName(fmt.Sprintf("%d", projectNumber), vpcId).String()

	return c.inner.Services.Connections.DeleteConnection(ServiceNetworkingServiceConnectionName, &servicenetworking.DeleteConnectionRequest{
		ConsumerNetwork: networkName,
	}).Do()
}

func (c *serviceNetworkingClient) ListServiceConnections(_ context.Context, projectId, vpcId string) ([]*servicenetworking.Connection, error) {
	networkName := gcputil.NewGlobalNetworkName(projectId, vpcId).String()
	out, err := c.inner.Services.Connections.List(ServiceNetworkingServicePath).Network(networkName).Do()
	if err != nil {
		return nil, err
	}
	return out.Connections, nil
}

func (c *serviceNetworkingClient) CreateServiceConnection(_ context.Context, projectId, vpcId string, reservedIpRanges []string) (*servicenetworking.Operation, error) {
	networkName := gcputil.NewGlobalNetworkName(projectId, vpcId).String()
	return c.inner.Services.Connections.Create(ServiceNetworkingServicePath, &servicenetworking.Connection{
		Network:               networkName,
		ReservedPeeringRanges: reservedIpRanges,
	}).Do()
}

func (c *serviceNetworkingClient) GetServiceNetworkingOperation(_ context.Context, operationName string) (*servicenetworking.Operation, error) {
	return c.inner.Operations.Get(operationName).Do()
}
