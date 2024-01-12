package client

import (
	"context"

	gcpclient "github.com/kyma-project/cloud-manager/components/kcp/pkg/provider/gcp/client"
	"google.golang.org/api/option"
	"google.golang.org/api/servicenetworking/v1"
)

type ServiceNetworkingClient interface {
	ListServiceConnections(ctx context.Context, projectId, vpcId string ) ([]*servicenetworking.Connection, error)
	CreateServiceConnection(ctx context.Context, projectId, vpcId, reservedIpRangeName string) (*servicenetworking.Operation, error)

}

func NewServiceNetworkingClient() gcpclient.ClientProvider[ServiceNetworkingClient] {
	return gcpclient.NewCachedClientProvider(
		func(ctx context.Context, saJsonKeyPath string) (ServiceNetworkingClient, error) {
			client, err := servicenetworking.NewService(ctx, option.WithCredentialsFile(saJsonKeyPath))
			if err != nil {
				return nil, err
			}
			return newServiceNetworkingClient(client), nil
		},
	)
}


func newServiceNetworkingClient(svcNet *servicenetworking.APIService) ServiceNetworkingClient {
	return &serviceNetworkingClient{svcNet: svcNet}
}

type serviceNetworkingClient struct {
	svcNet *servicenetworking.APIService
}


func (c *serviceNetworkingClient) ListServiceConnections(ctx context.Context, projectId, vpcId string) ([]*servicenetworking.Connection, error) {
	network := gcpclient.GetVPCPath(projectId, vpcId)
	out, err := c.svcNet.Services.Connections.List(gcpclient.ServiceNetworkingServicePath).Network(network).Do()
	if err != nil {
		return nil, err
	}
	return out.Connections, nil
}

func (c *serviceNetworkingClient) CreateServiceConnection(ctx context.Context, projectId, vpcId, reservedIpRangeName string) (*servicenetworking.Operation, error) {
	network := gcpclient.GetVPCPath(projectId, vpcId)
	var reservedIpRanges []string
	reservedIpRanges = append(reservedIpRanges, reservedIpRangeName)
	return c.svcNet.Services.Connections.Create(gcpclient.ServiceNetworkingServicePath, &servicenetworking.Connection{
		Network: network,
		ReservedPeeringRanges: reservedIpRanges,
		}).Do()
}




