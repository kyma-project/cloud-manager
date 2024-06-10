package mock

import (
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	"sync"
)

type VpcNetworkConfig interface {
	AddNetwork(subscription, resourceGroup, virtualNetworkName string, tags map[string]*string)
}

type networkEntry struct {
	resourceGroup string
	network       armnetwork.VirtualNetwork
}

type networkStore struct {
	m     sync.Mutex
	items []*networkEntry
}
