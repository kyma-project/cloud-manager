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
)

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
	default:
		return "", fmt.Errorf("invalid provider: %s", provider)
	}
}
