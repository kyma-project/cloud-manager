package client

import (
	"context"
	"fmt"

	"cloud.google.com/go/compute/apiv1/computepb"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"k8s.io/utils/ptr"
)

type CreateSubnetRequest struct {
	ProjectId             string
	Region                string
	Network               string
	Name                  string
	Cidr                  string
	IdempotenceId         string
	PrivateIpGoogleAccess bool
	Purpose               string
}

type GetSubnetRequest struct {
	ProjectId string
	Region    string
	Name      string
}

type DeleteSubnetRequest struct {
	ProjectId     string
	Region        string
	Name          string
	IdempotenceId string
}

type ComputeClient interface {
	CreateSubnet(ctx context.Context, request CreateSubnetRequest) (string, error)
	GetSubnet(ctx context.Context, request GetSubnetRequest) (*computepb.Subnetwork, error)
	DeleteSubnet(ctx context.Context, request DeleteSubnetRequest) error
}

func NewComputeClientProvider(gcpClients *gcpclient.GcpClients) gcpclient.GcpClientProvider[ComputeClient] {
	return func(_ string) ComputeClient {
		return NewComputeClient(gcpClients)
	}
}

func NewComputeClient(gcpClients *gcpclient.GcpClients) ComputeClient {
	return NewComputeClientFromSubnetClient(gcpClients.SubnetWrapped())
}

func NewComputeClientFromSubnetClient(subnetClient gcpclient.SubnetClient) ComputeClient {
	return &computeClient{subnetClient: subnetClient}
}

type computeClient struct {
	subnetClient gcpclient.SubnetClient
}

func (computeClient *computeClient) CreateSubnet(ctx context.Context, request CreateSubnetRequest) (string, error) {
	networkNameFull := fmt.Sprintf("projects/%s/global/networks/%s", request.ProjectId, request.Network)

	op, err := computeClient.subnetClient.InsertSubnet(ctx, &computepb.InsertSubnetworkRequest{
		Project: request.ProjectId,
		Region:  request.Region,
		SubnetworkResource: &computepb.Subnetwork{
			IpCidrRange:           ptr.To(request.Cidr),
			Name:                  ptr.To(request.Name),
			Network:               ptr.To(networkNameFull),
			PrivateIpGoogleAccess: ptr.To(request.PrivateIpGoogleAccess),
			Purpose:               ptr.To(request.Purpose),
		},
		RequestId: ptr.To(request.IdempotenceId),
	})

	if err != nil {
		return "", err
	}

	return op.Name(), nil
}

func (computeClient *computeClient) GetSubnet(ctx context.Context, request GetSubnetRequest) (*computepb.Subnetwork, error) {
	subnet, err := computeClient.subnetClient.GetSubnet(ctx, &computepb.GetSubnetworkRequest{
		Project:    request.ProjectId,
		Region:     request.Region,
		Subnetwork: request.Name,
	})

	if err != nil {
		return nil, err
	}

	return subnet, nil
}

func (computeClient *computeClient) DeleteSubnet(ctx context.Context, request DeleteSubnetRequest) error {
	_, err := computeClient.subnetClient.DeleteSubnet(ctx, &computepb.DeleteSubnetworkRequest{
		Project:    request.ProjectId,
		Region:     request.Region,
		Subnetwork: request.Name,
		RequestId:  ptr.To(request.IdempotenceId),
	})

	if err != nil {
		return err
	}

	return nil
}
