package client

import (
	"context"
	"fmt"

	"github.com/hashicorp/golang-lru/v2"
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/servicenetworking/v1"
)

type ServiceNetworkingClient interface {
}

type serviceNetworkingClient struct {
	inner *servicenetworking.APIService
	crm   *cloudresourcemanager.Service
}

var projectNumbersCache, _ = lru.New[string, int64](256)

func (c *serviceNetworkingClient) PatchServiceConnection(_ context.Context, projectId, vpcId string, reservedIpRanges []string) (*servicenetworking.Operation, error) {
	network := fmt.Sprintf("projects/%s/global/networks/%s", projectId, vpcId)
	return c.inner.Services.Connections.Patch(ServiceNetworkingServiceConnectionName, &servicenetworking.Connection{
		Network:               network,
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

	network := fmt.Sprintf("projects/%d/global/networks/%s", projectNumber, vpcId)

	return c.inner.Services.Connections.DeleteConnection(ServiceNetworkingServiceConnectionName, &servicenetworking.DeleteConnectionRequest{
		ConsumerNetwork: network,
	}).Do()
}

func (c *serviceNetworkingClient) ListServiceConnections(_ context.Context, projectId, vpcId string) ([]*servicenetworking.Connection, error) {
	network := fmt.Sprintf("projects/%s/global/networks/%s", projectId, vpcId)
	out, err := c.inner.Services.Connections.List(ServiceNetworkingServicePath).Network(network).Do()
	if err != nil {
		return nil, err
	}
	return out.Connections, nil
}

func (c *serviceNetworkingClient) CreateServiceConnection(_ context.Context, projectId, vpcId string, reservedIpRanges []string) (*servicenetworking.Operation, error) {
	network := fmt.Sprintf("projects/%s/global/networks/%s", projectId, vpcId)
	return c.inner.Services.Connections.Create(ServiceNetworkingServicePath, &servicenetworking.Connection{
		Network:               network,
		ReservedPeeringRanges: reservedIpRanges,
	}).Do()
}

func (c *serviceNetworkingClient) GetServiceNetworkingOperation(ctx context.Context, operationName string) (*servicenetworking.Operation, error) {
	return c.inner.Operations.Get(operationName).Do()
}
