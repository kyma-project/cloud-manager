package util

import (
	"fmt"
	"github.com/elliotchance/pie/v2"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"strings"
)

type ResourceDetails struct {
	index int
	valid bool

	Subscription    string
	ResourceGroup   string
	Provider        string
	ResourceType    string
	ResourceName    string
	SubResourceType string
	SubResourceName string
}

//	subscription                                  rg                         provider            res type     res name        subRes type.       subRes name
//
// /     0       /              1                    /      2       /        3          /    4    /       5         /        6      /    7       /          8           /        9
// /subscriptions/00000000-00000000-00000000-00000000/resourceGroups/RESOURCE_GROUP_NAME
// /subscriptions/00000000-00000000-00000000-00000000/resourceGroups/RESOURCE_GROUP_NAME/providers/Microsoft.Network/virtualNetworks/NETWORK_NAME
// /subscriptions/00000000-00000000-00000000-00000000/resourceGroups/RESOURCE_GROUP_NAME/providers/Microsoft.Network/virtualNetworks/NETWORK_NAME/virtualNetworkPeerings/PEERING_NAME
func (rd *ResourceDetails) setNext(v string) error {
	switch rd.index {
	case 0:
		if v != "subscriptions" {
			return fmt.Errorf("on position %d expected subscriptions, got %s", rd.index, v)
		}
	case 1:
		rd.Subscription = v
	case 2:
		if v != "resourceGroups" {
			return fmt.Errorf("on position %d expected resourceGroups, got %s", rd.index, v)
		}
	case 3:
		rd.ResourceGroup = v
	case 4:
		if v != "providers" {
			return fmt.Errorf("on position %d expected providers, got %s", rd.index, v)
		}
	case 5:
		rd.Provider = v
	case 6:
		rd.ResourceType = v
	case 7:
		rd.ResourceName = v
	case 8:
		rd.SubResourceType = v
	case 9:
		rd.SubResourceName = v
	case 10:
		return fmt.Errorf("nothring expected on position %d, got %s", rd.index, v)
	}
	rd.index++
	return nil
}

func (rd *ResourceDetails) IsValid() bool {
	return rd.valid
}

func (rd *ResourceDetails) String() string {
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf(
		"/subscriptions/%s/resourceGroups/%s",
		rd.Subscription,
		rd.ResourceGroup,
	))
	if rd.Provider != "" {
		builder.WriteString(fmt.Sprintf(
			"/providers/%s/%s/%s",
			rd.Provider,
			rd.ResourceType,
			rd.ResourceName,
		))
	}
	if rd.SubResourceType != "" {
		builder.WriteString(fmt.Sprintf(
			"/%s/%s",
			rd.SubResourceType,
			rd.SubResourceName,
		))
	}

	return builder.String()
}

func ParseResourceID(resourceID string) (ResourceDetails, error) {
	chunks := pie.Filter(strings.Split(resourceID, "/"), func(s string) bool {
		return len(s) > 0
	})
	rd := ResourceDetails{}
	for _, v := range chunks {
		if err := rd.setNext(v); err != nil {
			return rd, err
		}
	}

	rd.valid = true

	return rd, nil
}

func NewResourceGroupResourceId(subscription, resourceGroup string) *ResourceDetails {
	return &ResourceDetails{
		Subscription:  subscription,
		ResourceGroup: resourceGroup,
		valid:         len(subscription) > 0 && len(resourceGroup) > 0,
	}
}

type NetworkResourceId struct {
	*ResourceDetails
}

func (nr *NetworkResourceId) NetworkName() string {
	return nr.ResourceName
}

func NewVirtualNetworkResourceId(subscription, resourceGroup, virtualNetworkName string) *NetworkResourceId {
	return &NetworkResourceId{
		ResourceDetails: &ResourceDetails{
			Subscription:  subscription,
			ResourceGroup: resourceGroup,
			Provider:      "Microsoft.Network",
			ResourceType:  "virtualNetworks",
			ResourceName:  virtualNetworkName,
			valid:         len(subscription) > 0 && len(resourceGroup) > 0 && len(virtualNetworkName) > 0,
		},
	}
}

func NewVirtualNetworkResourceIdFromNetworkReference(ref *cloudcontrolv1beta1.NetworkReference) *NetworkResourceId {
	if ref == nil || ref.Azure == nil {
		return &NetworkResourceId{ResourceDetails: &ResourceDetails{}}
	}
	return NewVirtualNetworkResourceId(
		ref.Azure.SubscriptionId, ref.Azure.ResourceGroup, ref.Azure.NetworkName,
	)
}

type PeeringResourceId struct {
	*ResourceDetails
}

func (pr *PeeringResourceId) PeeringName() string {
	return pr.ResourceName
}

func NewVirtualNetworkPeeringResourceId(subscription, resourceGroup, virtualNetworkName, virtualNetworkPeeringName string) *PeeringResourceId {
	return &PeeringResourceId{
		ResourceDetails: &ResourceDetails{
			Subscription:    subscription,
			ResourceGroup:   resourceGroup,
			Provider:        "Microsoft.Network",
			ResourceType:    "virtualNetworks",
			ResourceName:    virtualNetworkName,
			SubResourceType: "virtualNetworkPeerings",
			SubResourceName: virtualNetworkPeeringName,
			valid:           len(subscription) > 0 && len(resourceGroup) > 0 && len(virtualNetworkName) > 0 && len(virtualNetworkPeeringName) > 0,
		},
	}
}

// https://portal.azure.com/#@sapsharedtenant.onmicrosoft.com/resource/subscriptions/3f1d2fbd-117a-4742-8bde-6edbcdee6a04/resourceGroups/cm-redis-0a1e1caa-1d2c-4eba-848e-9bdb3ef1535c/providers/Microsoft.Cache/Redis/0a1e1caa-1d2c-4eba-848e-9bdb3ef1535c/overview
func NewRedisInstanceResourceId(subscription, resourceGroup, redisInstanceName string) *ResourceDetails {
	return &ResourceDetails{
		Subscription:  subscription,
		ResourceGroup: resourceGroup,
		Provider:      "Microsoft.Cache",
		ResourceType:  "Redis",
		ResourceName:  redisInstanceName,
		valid:         len(subscription) > 0 && len(resourceGroup) > 0 && len(redisInstanceName) > 0,
	}
}
