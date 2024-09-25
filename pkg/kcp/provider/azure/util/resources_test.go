package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestResourceDetails(t *testing.T) {

	t.Run("parse resource group id in details", func(t *testing.T) {
		id := "/subscriptions/81e52275-d267-47e4-b18d-8fd92472a201/resourceGroups/MyResourceGroup"

		rd, _ := ParseResourceID(id)
		assert.Equal(t, rd.Subscription, "81e52275-d267-47e4-b18d-8fd92472a201")
		assert.Equal(t, rd.ResourceGroup, "MyResourceGroup")
		assert.Empty(t, rd.Provider)
		assert.Empty(t, rd.ResourceType)
		assert.Empty(t, rd.ResourceName)
		assert.Empty(t, rd.SubResourceType)
		assert.Empty(t, rd.SubResourceName)
	})

	t.Run("parse resource id details", func(t *testing.T) {
		id := "/subscriptions/7dd120b2-082d-47bf-93d9-b687602929c2/resourceGroups/MyResourceGroup/providers/Microsoft.Network/virtualNetworks/MyVnet"

		rd, _ := ParseResourceID(id)
		assert.Equal(t, rd.Subscription, "7dd120b2-082d-47bf-93d9-b687602929c2")
		assert.Equal(t, rd.ResourceGroup, "MyResourceGroup")
		assert.Equal(t, rd.Provider, "Microsoft.Network")
		assert.Equal(t, rd.ResourceType, "virtualNetworks")
		assert.Equal(t, rd.ResourceName, "MyVnet")
		assert.Empty(t, rd.SubResourceType)
		assert.Empty(t, rd.SubResourceName)
	})

	t.Run("parse subresource id details", func(t *testing.T) {
		id := "/subscriptions/0eebf528-679d-47cd-aa6a-cb91c2278ce6/resourceGroups/MyResourceGroup/providers/Microsoft.Network/virtualNetworks/MyVnet/virtualNetworkPeerings/MyPeering"

		rd, _ := ParseResourceID(id)
		assert.Equal(t, rd.Subscription, "0eebf528-679d-47cd-aa6a-cb91c2278ce6")
		assert.Equal(t, rd.ResourceGroup, "MyResourceGroup")
		assert.Equal(t, rd.Provider, "Microsoft.Network")
		assert.Equal(t, rd.ResourceType, "virtualNetworks")
		assert.Equal(t, rd.ResourceName, "MyVnet")
		assert.Equal(t, rd.SubResourceType, "virtualNetworkPeerings")
		assert.Equal(t, rd.SubResourceName, "MyPeering")
	})

	t.Run("parse resource id bulk", func(t *testing.T) {
		testData := []struct {
			givenId    string
			expectedId string
			rd         ResourceDetails
			err        bool
		}{
			{
				"/subscriptions/00000000-00000000-00000000-00000001/resourceGroups/RESOURCE_GROUP_01",
				"",
				ResourceDetails{
					Subscription:  "00000000-00000000-00000000-00000001",
					ResourceGroup: "RESOURCE_GROUP_01",
				},
				false,
			},
			{
				"/subscriptions/00000000-00000000-00000000-00000002/resourceGroups/RESOURCE_GROUP_02/providers/Microsoft.Network/virtualNetworks/NETWORK_02",
				"",
				ResourceDetails{
					Subscription:  "00000000-00000000-00000000-00000002",
					ResourceGroup: "RESOURCE_GROUP_02",
					Provider:      "Microsoft.Network",
					ResourceType:  "virtualNetworks",
					ResourceName:  "NETWORK_02",
				},
				false,
			},
			{
				"/subscriptions/00000000-00000000-00000000-00000003/resourceGroups/RESOURCE_GROUP_03/providers/Microsoft.Network/virtualNetworks/NETWORK_03/virtualNetworkPeerings/PEERING_03",
				"",
				ResourceDetails{
					Subscription:    "00000000-00000000-00000000-00000003",
					ResourceGroup:   "RESOURCE_GROUP_03",
					Provider:        "Microsoft.Network",
					ResourceType:    "virtualNetworks",
					ResourceName:    "NETWORK_03",
					SubResourceType: "virtualNetworkPeerings",
					SubResourceName: "PEERING_03",
				},
				false,
			},
			{
				"//subscriptions/00000000-00000000-00000000-00000004///resourceGroups/RESOURCE_GROUP_04///",
				"/subscriptions/00000000-00000000-00000000-00000004/resourceGroups/RESOURCE_GROUP_04",
				ResourceDetails{
					Subscription:  "00000000-00000000-00000000-00000004",
					ResourceGroup: "RESOURCE_GROUP_04",
				},
				false,
			},
			{
				"/foo/bar/baz",
				"",
				ResourceDetails{},
				true,
			},
			{
				"/subscriptions/00000000-00000000-00000000-00000005/foo/RESOURCE_GROUP_03/providers/Microsoft.Network/virtualNetworks/NETWORK_03/virtualNetworkPeerings/PEERING_03",
				"",
				ResourceDetails{},
				true,
			},
			{
				"/subscriptions/00000000-00000000-00000000-00000006/resourceGroups/RESOURCE_GROUP_03/foo/Microsoft.Network/virtualNetworks/NETWORK_03/virtualNetworkPeerings/PEERING_03",
				"",
				ResourceDetails{},
				true,
			},
			{
				"abcd",
				"",
				ResourceDetails{},
				true,
			},
		}

		for _, data := range testData {
			t.Run(data.givenId, func(t *testing.T) {
				actualRD, err := ParseResourceID(data.givenId)
				if data.err {
					assert.Error(t, err, "Expected error but got none")
				} else {
					assert.NoError(t, err, "Expected no error")

					actualId := actualRD.String()
					expectedRDId := data.rd.String()
					if data.expectedId == "" {
						data.expectedId = data.givenId
					}
					assert.Equal(t, data.expectedId, actualId)
					assert.Equal(t, data.expectedId, expectedRDId)
				}
			})
		}
	})

}

func TestResourceGroupResourceId(t *testing.T) {
	expected := "/subscriptions/30e0eb68-9aa7-45e9-8c47-57c1d0d6fa75/resourceGroups/MyResourceGroup"
	actual := NewResourceGroupResourceId("30e0eb68-9aa7-45e9-8c47-57c1d0d6fa75", "MyResourceGroup").String()
	assert.Equal(t, expected, actual)
}

func TestVirtualNetworkResourceId(t *testing.T) {
	expected := "/subscriptions/ecf40052-d84d-4282-9c4a-78bd50d6d59f/resourceGroups/MyResourceGroup/providers/Microsoft.Network/virtualNetworks/MyVnet"
	actual := NewVirtualNetworkResourceId("ecf40052-d84d-4282-9c4a-78bd50d6d59f", "MyResourceGroup", "MyVnet").String()
	assert.Equal(t, expected, actual)
}

func TestVirtualNetworkPeeringResourceId(t *testing.T) {
	expected := "/subscriptions/4f12b37a-2351-4973-95a3-0ee1bdcd9f5d/resourceGroups/MyResourceGroup/providers/Microsoft.Network/virtualNetworks/MyVnet/virtualNetworkPeerings/MyVnet-123"
	actual := NewVirtualNetworkPeeringResourceId("4f12b37a-2351-4973-95a3-0ee1bdcd9f5d", "MyResourceGroup", "MyVnet", "MyVnet-123").String()
	assert.Equal(t, expected, actual)
}

func TestRedisResourceId(t *testing.T) {
	expected := "/subscriptions/10375acd-7dc2-47f3-adf2-4627c93b7d36/resourceGroups/MyRG08/providers/Microsoft.Cache/Redis/MyRedis08"
	actual := NewRedisInstanceResourceId("10375acd-7dc2-47f3-adf2-4627c93b7d36", "MyRG08", "MyRedis08").String()
	assert.Equal(t, expected, actual)
}
