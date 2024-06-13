package mock

import (
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	"sync"
)

type peeringEntry struct {
	resourceGroupName  string
	virtualNetworkName string
	peering            armnetwork.VirtualNetworkPeering
}

type peeringStore struct {
	m     sync.Mutex
	items []*peeringEntry
}
