package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseResourceID(t *testing.T) {

	id := "/subscriptions/afdbc79f-de19-4df4-94cd-6be2739dc0e0/resourceGroups/MyResourceGroup/providers/Microsoft.Network/virtualNetworks/MyVnet/virtualNetworkPeerings/MyVnet-shoot--spm-test01--phx-azr-02"

	d, _ := ParseResourceID(id)
	assert.Equal(t, d.Subscription, "afdbc79f-de19-4df4-94cd-6be2739dc0e0")
	assert.Equal(t, d.ResourceGroup, "MyResourceGroup")
	assert.Equal(t, d.ResourceName, "MyVnet")
	assert.Equal(t, d.Provider, "Microsoft.Network")
	assert.Equal(t, d.ResourceType, "virtualNetworks")
}

func TestParseSubResourceID(t *testing.T) {

	id := "/subscriptions/afdbc79f-de19-4df4-94cd-6be2739dc0e0/resourceGroups/MyResourceGroup/providers/Microsoft.Network/virtualNetworks/MyVnet/virtualNetworkPeerings/MyVnet-shoot--spm-test01--phx-azr-02"

	d, _ := ParseResourceID(id)
	assert.Equal(t, d.Subscription, "afdbc79f-de19-4df4-94cd-6be2739dc0e0")
	assert.Equal(t, d.ResourceGroup, "MyResourceGroup")
	assert.Equal(t, d.ResourceName, "MyVnet")
	assert.Equal(t, d.Provider, "Microsoft.Network")
	assert.Equal(t, d.ResourceType, "virtualNetworks")
	assert.Equal(t, d.SubResourceType, "virtualNetworkPeerings")
	assert.Equal(t, d.SubResourceName, "MyVnet-shoot--spm-test01--phx-azr-02")
}

func TestVirtualNetworkPeeringResourceId(t *testing.T) {

	expected := "/subscriptions/afdbc79f-de19-4df4-94cd-6be2739dc0e0/resourceGroups/MyResourceGroup/providers/Microsoft.Network/virtualNetworks/MyVnet/virtualNetworkPeerings/MyVnet-shoot--spm-test01--phx-azr-02"

	actual := VirtualNetworkPeeringResourceId("afdbc79f-de19-4df4-94cd-6be2739dc0e0", "MyResourceGroup", "MyVnet", "MyVnet-shoot--spm-test01--phx-azr-02")

	assert.Equal(t, expected, actual)
}
