package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseResourceID(t *testing.T) {

	id := "/subscriptions/9c05f3c1-314b-4c4b-bfff-b5a0650177cb/resourceGroups/MyResourceGroup/providers/Microsoft.Network/virtualNetworks/MyVnet/virtualNetworkPeerings/MyVnet-shoot--spm-test01--phx-azr-02"

	d, _ := ParseResourceID(id)
	assert.Equal(t, d.Subscription, "9c05f3c1-314b-4c4b-bfff-b5a0650177cb")
	assert.Equal(t, d.ResourceGroup, "MyResourceGroup")
	assert.Equal(t, d.ResourceName, "MyVnet")
	assert.Equal(t, d.Provider, "Microsoft.Network")
	assert.Equal(t, d.ResourceType, "virtualNetworks")
}

func TestParseSubResourceID(t *testing.T) {

	id := "/subscriptions/9c05f3c1-314b-4c4b-bfff-b5a0650177cb/resourceGroups/MyResourceGroup/providers/Microsoft.Network/virtualNetworks/MyVnet/virtualNetworkPeerings/MyVnet-shoot--spm-test01--phx-azr-02"

	d, _ := ParseResourceID(id)
	assert.Equal(t, d.Subscription, "9c05f3c1-314b-4c4b-bfff-b5a0650177cb")
	assert.Equal(t, d.ResourceGroup, "MyResourceGroup")
	assert.Equal(t, d.ResourceName, "MyVnet")
	assert.Equal(t, d.Provider, "Microsoft.Network")
	assert.Equal(t, d.ResourceType, "virtualNetworks")
	assert.Equal(t, d.SubResourceType, "virtualNetworkPeerings")
	assert.Equal(t, d.SubResourceName, "MyVnet-shoot--spm-test01--phx-azr-02")
}

func TestVirtualNetworkPeeringResourceId(t *testing.T) {

	expected := "/subscriptions/9c05f3c1-314b-4c4b-bfff-b5a0650177cb/resourceGroups/MyResourceGroup/providers/Microsoft.Network/virtualNetworks/MyVnet/virtualNetworkPeerings/MyVnet-shoot--spm-test01--phx-azr-02"

	actual := VirtualNetworkPeeringResourceId("9c05f3c1-314b-4c4b-bfff-b5a0650177cb", "MyResourceGroup", "MyVnet", "MyVnet-shoot--spm-test01--phx-azr-02")

	assert.Equal(t, expected, actual)
}
