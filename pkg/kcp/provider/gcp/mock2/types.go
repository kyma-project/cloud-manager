package mock2

import (
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpexposeddataclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/exposedData/client"
)

type Clients interface {
	gcpclient.ComputeGlobalOperationsClient
	gcpclient.ComputeRegionalOperationsClient
	gcpclient.NetworkClient
	gcpclient.FilestoreClient
	gcpclient.GlobalAddressesClient
	gcpclient.NetworkConnectivityClient
	gcpclient.RedisClusterClient
	gcpclient.RedisInstanceClient
	gcpclient.RegionalAddressesClient
	gcpclient.RoutersClient
	gcpclient.ServiceNetworkingClient
	gcpclient.SubnetClient
}

type Providers interface {
	ExposedDataProvider() gcpclient.GcpClientProvider[gcpexposeddataclient.Client]
	// all others feature's providers as they are refactored to switch using these new GCP clients
}

type Store interface {
	Clients
}

type Subscription interface {
	Store

	ProjectId() string
	Delete()
}

type Server interface {
	Providers

	// NewSubscription creates new subscription with a random projectId and given prefix. Each test scenario should
	// at the beginning create subscription and defer delete it at the end of the test scenario. For the prefix use
	// whole or part of the scenario name, which would make it easier to identify resources created by the test scenario.
	NewSubscription(prefix string) Subscription

	// GetSubscription is used from the providers methods and there's no need for direct usage.
	// It returns previously created subscription with the given projectId. If there is no subscription with
	// such projectId, nil is returned, intentionally so that reconciler fails, which would indicate invalid test setup
	// and a signal to developer to create the subscription at the beginning of the test.
	GetSubscription(projectId string) Subscription

	// DeleteSubscription deletes subscription with the given projectId from the mock server. Each test scenario should
	// at the beginning create subscription and defer delete it at the end of the test scenario.
	// That way all resources created by the test scenario are encapsulated and deleted at the end of the test.
	// This provides test isolation and prevents interference between test scenarios, which is crucial for reliable and maintainable tests.
	DeleteSubscription(projectId string)
}
