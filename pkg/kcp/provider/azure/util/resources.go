package util

import (
	"fmt"
	"regexp"
	"strings"
)

type ResourceDetails struct {
	Subscription    string
	ResourceGroup   string
	Provider        string
	ResourceType    string
	ResourceName    string
	SubResourceType string
	SubResourceName string
}

func ParseResourceID(resourceID string) (ResourceDetails, error) {

	const resourceIDPatternText = `/subscriptions/(?P<subscription>.+)/resourceGroups/(?P<resourceGroup>.+)/providers/(?P<provider>[^/]+)/(?P<type>[^/]+)/(?P<name>[^/]+)/?(?P<subType>[^/]+)?/?(?P<subName>.+)?`
	resourceIDPattern := regexp.MustCompile(resourceIDPatternText)
	match := resourceIDPattern.FindStringSubmatch(resourceID)

	groups := make(map[string]string)

	groups["subType"] = ""
	groups["subName"] = ""

	for i, name := range resourceIDPattern.SubexpNames() {
		if i != 0 && name != "" {
			groups[name] = match[i]
		}
	}

	if len(match) == 0 {
		return ResourceDetails{}, fmt.Errorf("parsing failed for %s. Invalid resource Id format", resourceID)
	}

	result := ResourceDetails{
		Subscription:    groups["subscription"],
		ResourceGroup:   groups["resourceGroup"],
		Provider:        groups["provider"],
		ResourceType:    groups["type"],
		ResourceName:    groups["name"],
		SubResourceType: groups["subType"],
		SubResourceName: groups["subName"],
	}

	return result, nil
}

func ResourceID(details ResourceDetails) string {
	return fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/%s/%s",
		details.Subscription,
		details.ResourceGroup,
		details.Provider,
		strings.Join([]string{
			details.ResourceType,
			details.ResourceName,
			details.SubResourceType,
			details.SubResourceName}, "/"))

}
func VirtualNetworkPeeringResourceId(subscription, resourceGroup, virtualNetworkName, virtualNetworkPeeringName string) string {
	return ResourceID(ResourceDetails{
		Subscription:    subscription,
		ResourceGroup:   resourceGroup,
		Provider:        "Microsoft.Network",
		ResourceType:    "virtualNetworks",
		ResourceName:    virtualNetworkName,
		SubResourceType: "virtualNetworkPeerings",
		SubResourceName: virtualNetworkPeeringName,
	})
}

func VirtualNetworkResourceId(subscription, resourceGroup, virtualNetworkName string) string {
	return ResourceID(ResourceDetails{
		Subscription:  subscription,
		ResourceGroup: resourceGroup,
		Provider:      "Microsoft.Network",
		ResourceType:  "virtualNetworks",
		ResourceName:  virtualNetworkName,
	})
}
