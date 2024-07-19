package v1beta1

type ProviderType string

const (
	ProviderGCP       = ProviderType("gcp")
	ProviderAzure     = ProviderType("azure")
	ProviderAws       = ProviderType("aws")
	ProviderOpenStack = ProviderType("openstack")
)
