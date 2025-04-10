package v3

import (
	"context"
	"fmt"

	networkconnectivity "cloud.google.com/go/networkconnectivity/apiv1"
	"cloud.google.com/go/networkconnectivity/apiv1/networkconnectivitypb"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"google.golang.org/api/option"
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
	GetServiceConnectionPolicy(ctx context.Context, name string) (*networkconnectivitypb.ServiceConnectionPolicy, error)
	DeleteServiceConnectionPolicy(ctx context.Context, request DeleteServiceConnectionPolicyRequest) error
}

func NewNetworkConnectivityClientProvider() client.ClientProvider[NetworkConnectivityClient] {
	return func(ctx context.Context, saJsonKeyPath string) (NetworkConnectivityClient, error) {
		return NewNetworkConnectivityClient(saJsonKeyPath), nil
	}
}

func NewNetworkConnectivityClient(saJsonKeyPath string) NetworkConnectivityClient {
	return &networkConnectivityClient{saJsonKeyPath: saJsonKeyPath}
}

type networkConnectivityClient struct {
	saJsonKeyPath string
}

func (ncClient *networkConnectivityClient) CreateServiceConnectionPolicy(ctx context.Context, request CreateServiceConnectionPolicyRequest) error {
	client, err := networkconnectivity.NewCrossNetworkAutomationClient(ctx, option.WithCredentialsFile(ncClient.saJsonKeyPath))
	if err != nil {
		return err
	}
	defer client.Close() // nolint: errcheck

	networkNameFull := fmt.Sprintf("projects/%s/global/networks/%s", request.ProjectId, request.Network)
	parent := fmt.Sprintf("projects/%s/locations/%s", request.ProjectId, request.Region)
	connectionPolicyNameFull := fmt.Sprintf("%s/serviceConnectionPolicies/%s", parent, request.Name)

	_, err = client.CreateServiceConnectionPolicy(ctx, &networkconnectivitypb.CreateServiceConnectionPolicyRequest{
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
	client, err := networkconnectivity.NewCrossNetworkAutomationClient(ctx, option.WithCredentialsFile(ncClient.saJsonKeyPath))
	if err != nil {
		return nil, err
	}
	defer client.Close() // nolint: errcheck

	connectionPolicy, err := client.GetServiceConnectionPolicy(ctx, &networkconnectivitypb.GetServiceConnectionPolicyRequest{
		Name: name,
	})

	if err != nil {
		return nil, err
	}

	return connectionPolicy, nil
}

func (ncClient *networkConnectivityClient) DeleteServiceConnectionPolicy(ctx context.Context, request DeleteServiceConnectionPolicyRequest) error {
	client, err := networkconnectivity.NewCrossNetworkAutomationClient(ctx, option.WithCredentialsFile(ncClient.saJsonKeyPath))
	if err != nil {
		return err
	}
	defer client.Close() // nolint: errcheck

	_, err = client.DeleteServiceConnectionPolicy(ctx, &networkconnectivitypb.DeleteServiceConnectionPolicyRequest{
		Name:      request.Name,
		RequestId: request.IdempotenceId,
	})

	if err != nil {
		return err
	}

	return nil
}
