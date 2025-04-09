package v3

import (
	"context"
	"fmt"

	compute "cloud.google.com/go/compute/apiv1"
	computepb "cloud.google.com/go/compute/apiv1/computepb"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"google.golang.org/api/option"
	"k8s.io/utils/ptr"
)

type CreateSubnetRequest struct {
	ProjectId     string
	Region        string
	Network       string
	Name          string
	Cidr          string
	IdempotenceId string
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
	CreatePrivateSubnet(ctx context.Context, request CreateSubnetRequest) error
	GetSubnet(ctx context.Context, request GetSubnetRequest) (*computepb.Subnetwork, error)
	DeleteSubnet(ctx context.Context, request DeleteSubnetRequest) error
}

func NewComputeClientProvider() client.ClientProvider[ComputeClient] {
	return func(ctx context.Context, saJsonKeyPath string) (ComputeClient, error) {
		return NewComputeClient(saJsonKeyPath), nil
	}
}

func NewComputeClient(saJsonKeyPath string) ComputeClient {
	return &computeClient{saJsonKeyPath: saJsonKeyPath}
}

type computeClient struct {
	saJsonKeyPath string
}

// Creates a Subnet with Purpose set to PRIVATE and PrivateIpGoogleAccess set to true
func (computeClient *computeClient) CreatePrivateSubnet(ctx context.Context, request CreateSubnetRequest) error {
	subnetClient, err := compute.NewSubnetworksRESTClient(ctx, option.WithCredentialsFile(computeClient.saJsonKeyPath))
	if err != nil {
		return err
	}
	defer subnetClient.Close()

	networkNameFull := fmt.Sprintf("projects/%s/global/networks/%s", request.ProjectId, request.Network)

	_, err = subnetClient.Insert(ctx, &computepb.InsertSubnetworkRequest{
		Project: request.ProjectId,
		Region:  request.Region,
		SubnetworkResource: &computepb.Subnetwork{
			IpCidrRange:           ptr.To(request.Cidr),
			Name:                  ptr.To(request.Name),
			Network:               ptr.To(networkNameFull),
			PrivateIpGoogleAccess: ptr.To(true),
			Purpose:               ptr.To("PRIVATE"),
		},
		RequestId: ptr.To(request.IdempotenceId),
	})

	if err != nil {
		return err
	}

	return nil
}

func (computeClient *computeClient) GetSubnet(ctx context.Context, request GetSubnetRequest) (*computepb.Subnetwork, error) {
	subnetClient, err := compute.NewSubnetworksRESTClient(ctx, option.WithCredentialsFile(computeClient.saJsonKeyPath))
	if err != nil {
		return nil, err
	}
	defer subnetClient.Close()

	subnet, err := subnetClient.Get(ctx, &computepb.GetSubnetworkRequest{
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
	subnetClient, err := compute.NewSubnetworksRESTClient(ctx, option.WithCredentialsFile(computeClient.saJsonKeyPath))
	if err != nil {
		return err
	}
	defer subnetClient.Close()

	_, err = subnetClient.Delete(ctx, &computepb.DeleteSubnetworkRequest{
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
