package client

import (
	"context"
	"fmt"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
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
	return &computeClient{subnetworksClient: gcpClients.ComputeSubnetworks}
}

type computeClient struct {
	subnetworksClient *compute.SubnetworksClient
}

func (computeClient *computeClient) CreateSubnet(ctx context.Context, request CreateSubnetRequest) (string, error) {
	networkNameFull := fmt.Sprintf("projects/%s/global/networks/%s", request.ProjectId, request.Network)

	op, err := computeClient.subnetworksClient.Insert(ctx, &computepb.InsertSubnetworkRequest{
		Project: request.ProjectId,
		Region:  request.Region,
		SubnetworkResource: &computepb.Subnetwork{
			IpCidrRange:           new(request.Cidr),
			Name:                  new(request.Name),
			Network:               new(networkNameFull),
			PrivateIpGoogleAccess: new(request.PrivateIpGoogleAccess),
			Purpose:               new(request.Purpose),
		},
		RequestId: new(request.IdempotenceId),
	})

	if err != nil {
		return "", err
	}

	return op.Name(), nil
}

func (computeClient *computeClient) GetSubnet(ctx context.Context, request GetSubnetRequest) (*computepb.Subnetwork, error) {
	subnet, err := computeClient.subnetworksClient.Get(ctx, &computepb.GetSubnetworkRequest{
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
	_, err := computeClient.subnetworksClient.Delete(ctx, &computepb.DeleteSubnetworkRequest{
		Project:    request.ProjectId,
		Region:     request.Region,
		Subnetwork: request.Name,
		RequestId:  new(request.IdempotenceId),
	})

	if err != nil {
		return err
	}

	return nil
}
