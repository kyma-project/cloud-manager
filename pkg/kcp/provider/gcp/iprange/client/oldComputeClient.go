package client

import (
	"context"
	"fmt"

	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

// OldComputeClient provides the original OLD pattern implementation using Google Discovery API.
// This is for v2 legacy code to maintain exact original behavior.
// Uses google.golang.org/api/compute/v1 (Discovery API, not gRPC).
type OldComputeClient interface {
	ListGlobalAddresses(ctx context.Context, projectId, vpc string) (*compute.AddressList, error)
	CreatePscIpRange(ctx context.Context, projectId, vpcName, name, description, address string, prefixLength int64) (*compute.Operation, error)
	DeleteIpRange(ctx context.Context, projectId, name string) (*compute.Operation, error)
	GetIpRange(ctx context.Context, projectId, name string) (*compute.Address, error)
	GetGlobalOperation(ctx context.Context, projectId, operationName string) (*compute.Operation, error)
}

type oldComputeClient struct {
	computeService *compute.Service
}

// NewOldComputeClientProviderV2 creates a ClientProvider for OldComputeClient (v2 legacy).
// For v2 legacy code, we need the Discovery API compute.Service. Since GcpClients only has
// modern gRPC clients, we create a Discovery API service but wrap the existing clients
// from GcpClients to avoid creating separate auth infrastructure.
func NewOldComputeClientProviderV2(gcpClients *gcpclient.GcpClients) gcpclient.ClientProvider[OldComputeClient] {
	return gcpclient.NewCachedClientProvider(
		func(ctx context.Context, credentialsFile string) (OldComputeClient, error) {
			// For Discovery API, we need to create compute.Service with HTTP client
			// Reuse the same credentials/auth that GcpClients uses
			baseClient, err := gcpclient.GetCachedGcpClient(ctx, credentialsFile)
			if err != nil {
				return nil, fmt.Errorf("error obtaining GCP HTTP Client: %w", err)
			}

			httpClient := gcpclient.NewMetricsHTTPClient(baseClient.Transport)

			computeService, err := compute.NewService(ctx, option.WithHTTPClient(httpClient))
			if err != nil {
				return nil, fmt.Errorf("error obtaining GCP Compute Service: %w", err)
			}

			return &oldComputeClient{computeService: computeService}, nil
		},
	)
}

func (c *oldComputeClient) GetIpRange(ctx context.Context, projectId, name string) (*compute.Address, error) {
	return c.computeService.GlobalAddresses.Get(projectId, name).Context(ctx).Do()
}

func (c *oldComputeClient) DeleteIpRange(ctx context.Context, projectId, name string) (*compute.Operation, error) {
	return c.computeService.GlobalAddresses.Delete(projectId, name).Context(ctx).Do()
}

func (c *oldComputeClient) CreatePscIpRange(ctx context.Context, projectId, vpcName, name, description, address string, prefixLength int64) (*compute.Operation, error) {
	addr := &compute.Address{
		Name:         name,
		Description:  description,
		Address:      address,
		PrefixLength: prefixLength,
		AddressType:  "INTERNAL",
		Purpose:      "VPC_PEERING",
		Network:      fmt.Sprintf("projects/%s/global/networks/%s", projectId, vpcName),
	}

	return c.computeService.GlobalAddresses.Insert(projectId, addr).Context(ctx).Do()
}

func (c *oldComputeClient) ListGlobalAddresses(ctx context.Context, projectId, vpc string) (*compute.AddressList, error) {
	filter := fmt.Sprintf("network=\"https://www.googleapis.com/compute/v1/projects/%s/global/networks/%s\"", projectId, vpc)
	return c.computeService.GlobalAddresses.List(projectId).Filter(filter).Context(ctx).Do()
}

func (c *oldComputeClient) GetGlobalOperation(ctx context.Context, projectId, operationName string) (*compute.Operation, error) {
	return c.computeService.GlobalOperations.Get(projectId, operationName).Context(ctx).Do()
}
