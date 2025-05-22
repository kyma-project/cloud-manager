package mock

import (
	"context"
	"sync"

	"cloud.google.com/go/networkconnectivity/apiv1/networkconnectivitypb"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/subnet"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/subnet/client"
	"google.golang.org/api/googleapi"
)

type networkConnectivityClientFake struct {
	mutex              sync.Mutex
	connectionPolicies map[string]*networkconnectivitypb.ServiceConnectionPolicy
}

func (ncClientFake *networkConnectivityClientFake) CreateServiceConnectionPolicy(ctx context.Context, request client.CreateServiceConnectionPolicyRequest) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	ncClientFake.mutex.Lock()
	defer ncClientFake.mutex.Unlock()

	name := subnet.GetServiceConnectionPolicyFullName(request.ProjectId, request.Region, request.Network)
	connectionPolicy := &networkconnectivitypb.ServiceConnectionPolicy{
		Name:         name,
		Network:      request.Network,
		ServiceClass: "gcp-memorystore-redis",
		Description:  "Managed by cloud-manager. Used for Redis Cluster.",
		PscConfig: &networkconnectivitypb.ServiceConnectionPolicy_PscConfig{
			Subnetworks:              request.Subnets,
			ProducerInstanceLocation: networkconnectivitypb.ServiceConnectionPolicy_PscConfig_PRODUCER_INSTANCE_LOCATION_UNSPECIFIED,
		},
	}

	ncClientFake.connectionPolicies[name] = connectionPolicy

	return nil
}

func (ncClientFake *networkConnectivityClientFake) UpdateServiceConnectionPolicy(ctx context.Context, policy *networkconnectivitypb.ServiceConnectionPolicy, updateMask []string) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	ncClientFake.mutex.Lock()
	defer ncClientFake.mutex.Unlock()

	ncClientFake.connectionPolicies[policy.Name] = policy

	return nil
}

func (ncClientFake *networkConnectivityClientFake) GetServiceConnectionPolicy(ctx context.Context, name string) (*networkconnectivitypb.ServiceConnectionPolicy, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	ncClientFake.mutex.Lock()
	defer ncClientFake.mutex.Unlock()

	if instance, ok := ncClientFake.connectionPolicies[name]; ok {
		return instance, nil
	}

	return nil, &googleapi.Error{
		Code:    404,
		Message: "Not Found",
	}
}

func (ncClientFake *networkConnectivityClientFake) DeleteServiceConnectionPolicy(ctx context.Context, request client.DeleteServiceConnectionPolicyRequest) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	ncClientFake.mutex.Lock()
	defer ncClientFake.mutex.Unlock()

	if _, ok := ncClientFake.connectionPolicies[request.Name]; ok {
		delete(ncClientFake.connectionPolicies, request.Name)
		return nil
	}

	return &googleapi.Error{
		Code:    404,
		Message: "Not Found",
	}
}
