package mock2

import (
	"sync"

	"cloud.google.com/go/compute/apiv1/computepb"
	"cloud.google.com/go/filestore/apiv1/filestorepb"
	"cloud.google.com/go/networkconnectivity/apiv1/networkconnectivitypb"
	"cloud.google.com/go/redis/apiv1/redispb"
	"cloud.google.com/go/redis/cluster/apiv1/clusterpb"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"google.golang.org/api/servicenetworking/v1"
)

func newStore() Store {
	result := &store{
		computeOperations: MustNewFilterableList[*computepb.Operation](),

		addressSpaces: make(map[string]*AddressSpace),
		networks:      MustNewFilterableList[*computepb.Network](),
		subnets:       MustNewFilterableList[*computepb.Subnetwork](),
		routers:       MustNewFilterableList[*computepb.Router](),
		addresses:     MustNewFilterableList[*computepb.Address](),

		serviceConnectionPolicies: MustNewFilterableList[*networkconnectivitypb.ServiceConnectionPolicy](),

		longRunningOperations: MustNewFilterableList[*OperationLongRunningBuilder](WithoutFilter),

		filestores: MustNewFilterableList[*filestorepb.Instance](),
		backups:    MustNewFilterableList[*filestorepb.Backup](),

		redisInstances: MustNewFilterableList[*redispb.Instance](),
		redisClusters:  MustNewFilterableList[*clusterpb.Cluster](),

		serviceNetworkingOperations: MustNewFilterableList[*servicenetworking.Operation](),
		serviceConnections:          MustNewFilterableList[*servicenetworking.Connection](),
	}

	return result
}

var _ Store = (*store)(nil)

type store struct {
	m sync.Mutex

	computeOperations *FilterableList[*computepb.Operation]

	addressSpaces map[string]*AddressSpace
	networks      *FilterableList[*computepb.Network]
	subnets       *FilterableList[*computepb.Subnetwork]
	routers       *FilterableList[*computepb.Router]
	addresses     *FilterableList[*computepb.Address]

	serviceConnectionPolicies *FilterableList[*networkconnectivitypb.ServiceConnectionPolicy]

	longRunningOperations *FilterableList[*OperationLongRunningBuilder]

	filestores *FilterableList[*filestorepb.Instance]
	backups    *FilterableList[*filestorepb.Backup]

	redisInstances *FilterableList[*redispb.Instance]
	redisClusters  *FilterableList[*clusterpb.Cluster]

	serviceNetworkingOperations *FilterableList[*servicenetworking.Operation]
	serviceConnections          *FilterableList[*servicenetworking.Connection]
}

var _ gcpclient.ComputeRegionalOperationsClient = (*store)(nil)
var _ gcpclient.ComputeGlobalOperationsClient = (*store)(nil)
var _ gcpclient.NetworkClient = (*store)(nil)
var _ gcpclient.SubnetClient = (*store)(nil)
var _ gcpclient.RoutersClient = (*store)(nil)
var _ gcpclient.GlobalAddressesClient = (*store)(nil)
var _ gcpclient.RegionalAddressesClient = (*store)(nil)
var _ gcpclient.FilestoreClient = (*store)(nil)
var _ gcpclient.RedisInstanceClient = (*store)(nil)
var _ gcpclient.RedisClusterClient = (*store)(nil)
var _ gcpclient.NetworkConnectivityClient = (*store)(nil)
