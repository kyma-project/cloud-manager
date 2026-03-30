package client

import (
	"context"
	"fmt"

	"cloud.google.com/go/networkconnectivity/apiv1/networkconnectivitypb"
	"github.com/google/uuid"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

type CreateServiceConnectionPolicyRequest struct {
	ProjectId     string
	Region        string
	Name          string
	Network       string
	Subnets       []string
	IdempotenceId string
}

type DeleteServiceConnectionPolicyRequest struct {
	Name          string
	IdempotenceId string
}

type NetworkConnectivityClient interface {
	CreateServiceConnectionPolicy(ctx context.Context, request CreateServiceConnectionPolicyRequest) error
	UpdateServiceConnectionPolicy(ctx context.Context, policy *networkconnectivitypb.ServiceConnectionPolicy, updateMask []string) error
	GetServiceConnectionPolicy(ctx context.Context, name string) (*networkconnectivitypb.ServiceConnectionPolicy, error)
	DeleteServiceConnectionPolicy(ctx context.Context, request DeleteServiceConnectionPolicyRequest) error
}

func NewNetworkConnectivityClientProvider(gcpClients *gcpclient.GcpClients) gcpclient.GcpClientProvider[NetworkConnectivityClient] {
	return func(_ string) NetworkConnectivityClient {
		return NewNetworkConnectivityClient(gcpClients)
	}
}

func NewNetworkConnectivityClient(gcpClients *gcpclient.GcpClients) NetworkConnectivityClient {
	return NewNetworkConnectivityClientFromWrapped(gcpClients.NetworkConnectivityWrapped())
}

func NewNetworkConnectivityClientFromWrapped(ncClient gcpclient.NetworkConnectivityClient) NetworkConnectivityClient {
	return &networkConnectivityClient{ncClient: ncClient}
}

type networkConnectivityClient struct {
	ncClient gcpclient.NetworkConnectivityClient
}

func (ncClient *networkConnectivityClient) UpdateServiceConnectionPolicy(ctx context.Context, policy *networkconnectivitypb.ServiceConnectionPolicy, updateMask []string) error {
	_, err := ncClient.ncClient.UpdateServiceConnectionPolicy(ctx, &networkconnectivitypb.UpdateServiceConnectionPolicyRequest{
		ServiceConnectionPolicy: policy,
		RequestId:               uuid.NewString(),
		UpdateMask: &fieldmaskpb.FieldMask{
			Paths: updateMask,
		},
	})

	if err != nil {
		return err
	}

	return nil
}

func (ncClient *networkConnectivityClient) CreateServiceConnectionPolicy(ctx context.Context, request CreateServiceConnectionPolicyRequest) error {
	networkNameFull := fmt.Sprintf("projects/%s/global/networks/%s", request.ProjectId, request.Network)
	parent := fmt.Sprintf("projects/%s/locations/%s", request.ProjectId, request.Region)
	connectionPolicyNameFull := fmt.Sprintf("%s/serviceConnectionPolicies/%s", parent, request.Name)

	_, err := ncClient.ncClient.CreateServiceConnectionPolicy(ctx, &networkconnectivitypb.CreateServiceConnectionPolicyRequest{
		Parent:                    parent,
		ServiceConnectionPolicyId: request.Name,
		ServiceConnectionPolicy: &networkconnectivitypb.ServiceConnectionPolicy{
			Name:         connectionPolicyNameFull,
			Network:      networkNameFull,
			ServiceClass: "gcp-memorystore-redis",
			Description:  "Managed by cloud-manager. Used for Redis Cluster.",
			PscConfig: &networkconnectivitypb.ServiceConnectionPolicy_PscConfig{
				Subnetworks:              request.Subnets,
				ProducerInstanceLocation: networkconnectivitypb.ServiceConnectionPolicy_PscConfig_PRODUCER_INSTANCE_LOCATION_UNSPECIFIED,
			},
		},
		RequestId: request.IdempotenceId,
	})

	if err != nil {
		return err
	}

	return nil
}

func (ncClient *networkConnectivityClient) GetServiceConnectionPolicy(ctx context.Context, name string) (*networkconnectivitypb.ServiceConnectionPolicy, error) {
	connectionPolicy, err := ncClient.ncClient.GetServiceConnectionPolicy(ctx, &networkconnectivitypb.GetServiceConnectionPolicyRequest{
		Name: name,
	})

	if err != nil {
		return nil, err
	}

	return connectionPolicy, nil
}

func (ncClient *networkConnectivityClient) DeleteServiceConnectionPolicy(ctx context.Context, request DeleteServiceConnectionPolicyRequest) error {
	_, err := ncClient.ncClient.DeleteServiceConnectionPolicy(ctx, &networkconnectivitypb.DeleteServiceConnectionPolicyRequest{
		Name:      request.Name,
		RequestId: request.IdempotenceId,
	})

	if err != nil {
		return err
	}

	return nil
}
