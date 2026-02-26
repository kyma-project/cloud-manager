package mock2

import (
	"fmt"
	"sync"

	"cloud.google.com/go/compute/apiv1/computepb"
	"cloud.google.com/go/filestore/apiv1/filestorepb"
	"cloud.google.com/go/longrunning/autogen/longrunningpb"
	"cloud.google.com/go/redis/apiv1/redispb"
	"cloud.google.com/go/redis/cluster/apiv1/clusterpb"
	"github.com/elliotchance/pie/v2"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/mock2/filter"
	gcputil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func newStore() Store {
	return &store{
		computeOperations:        mustNewList[*computepb.Operation](),
		reactorComputeOperations: newReactor(reactComputeOperationDone),

		addressSpaces: make(map[string]*AddressSpace),
		networks:      mustNewList[*computepb.Network](),
		subnets:       mustNewList[*computepb.Subnetwork](),
		routers:       mustNewList[*computepb.Router](),
		addresses:     mustNewList[*computepb.Address](),

		longRunningOperations:        mustNewList[*longrunningpb.Operation](),
		reactorLongRunningOperations: newReactor[*longrunningpb.Operation]( /* TODO: reactLongRunningOperationDone */ ),

		filestores:     mustNewList[*filestorepb.Instance](),
		redisInstances: mustNewList[*redispb.Instance](),
		redisClusters:  mustNewList[*clusterpb.Cluster](),
	}
}

var _ Store = (*store)(nil)

type store struct {
	m sync.Mutex

	computeOperations        *list[*computepb.Operation]
	reactorComputeOperations Reactor[*computepb.Operation]

	addressSpaces map[string]*AddressSpace
	networks      *list[*computepb.Network]
	subnets       *list[*computepb.Subnetwork]
	routers       *list[*computepb.Router]
	addresses     *list[*computepb.Address]

	longRunningOperations        *list[*longrunningpb.Operation]
	reactorLongRunningOperations Reactor[*longrunningpb.Operation]

	filestores     *list[*filestorepb.Instance]
	redisInstances *list[*redispb.Instance]
	redisClusters  *list[*clusterpb.Cluster]
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

type listItem[T any] struct {
	obj  T
	name gcputil.NameDetail
}

type list[T any] struct {
	items        []listItem[T]
	filterEngine *filter.FilterEngine[T]
}

func mustNewList[T any]() *list[T] {
	return util.Must(newList[T]())
}

func newList[T any]() (*list[T], error) {
	fe, err := filter.NewFilterEngine[T]()
	if err != nil {
		return nil, err
	}
	return &list[T]{
		filterEngine: fe,
	}, nil
}

func (l *list[T]) add(obj T, name gcputil.NameDetail) {
	l.items = append(l.items, listItem[T]{
		obj:  obj,
		name: name,
	})
}

func (l *list[T]) getItems() []T {
	return pie.Map(l.items, func(i listItem[T]) T {
		return i.obj
	})
}

func (l *list[T]) findByName(name gcputil.NameDetail) (T, bool) {
	for _, item := range l.items {
		if item.name.Equal(name) {
			return item.obj, true
		}
	}
	var zero T
	return zero, false
}

func (l *list[T]) filterByExpression(f *string) (*list[T], error) {
	if f == nil {
		return l, nil
	}
	result := &list[T]{
		filterEngine: l.filterEngine,
	}
	for _, item := range l.items {
		match, err := l.filterEngine.Match(*f, item.obj)
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate filter expression: %w", err)
		}
		if match {
			result.items = append(result.items, item)
		}
	}
	return result, nil
}

func (l *list[T]) filterByCallback(f func(l listItem[T]) bool) *list[T] {
	result := &list[T]{
		filterEngine: l.filterEngine,
	}
	for _, item := range l.items {
		if f(item) {
			result.items = append(result.items, item)
		}
	}
	return result
}

func (l *list[T]) filterNotByCallback(f func(l listItem[T]) bool) *list[T] {
	result := &list[T]{
		filterEngine: l.filterEngine,
	}
	for _, item := range l.items {
		if !f(item) {
			result.items = append(result.items, item)
		}
	}
	return result
}

func (l *list[T]) toIterator() gcpclient.Iterator[T] {
	return &iteratorMocked[T]{
		items: pie.Map(l.items, func(i listItem[T]) T {
			return i.obj
		}),
	}
}
