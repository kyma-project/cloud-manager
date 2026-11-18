package v1beta1

import (
	"fmt"
	"strings"
)

type ProviderType string

const (
	ProviderGCP       = ProviderType("gcp")
	ProviderAzure     = ProviderType("azure")
	ProviderAws       = ProviderType("aws")
	ProviderOpenStack = ProviderType("openstack")
	ProviderAlicloud  = ProviderType("alicloud")
)

// AllProviders returns every defined ProviderType constant in canonical order.
// Callers iterating all providers use this to stay in sync when a new provider is added.
func AllProviders() []ProviderType {
	return []ProviderType{
		ProviderAws,
		ProviderGCP,
		ProviderAzure,
		ProviderOpenStack,
		ProviderAlicloud,
	}
}

func ParseProviderType(provider string) (ProviderType, error) {
	switch strings.ToLower(provider) {
	case string(ProviderGCP):
		return ProviderGCP, nil
	case string(ProviderAzure):
		return ProviderAzure, nil
	case string(ProviderAws):
		return ProviderAws, nil
	case string(ProviderOpenStack):
		return ProviderOpenStack, nil
	case string(ProviderAlicloud):
		return ProviderAlicloud, nil
	default:
		return "", fmt.Errorf("invalid provider: %s", provider)
	}
}
