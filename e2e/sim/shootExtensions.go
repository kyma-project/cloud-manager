package sim

// local types for shoot extensions — json.Marshal into Raw bytes, avoids importing
// gardener-extension-shoot-cert-service and gardener-extension-shoot-dns-service packages
// whose types lack proper json struct tags.

type shootCertConfig struct {
	APIVersion   string            `json:"apiVersion"`
	Kind         string            `json:"kind"`
	ShootIssuers *shootCertIssuers `json:"shootIssuers,omitempty"`
}

type shootCertIssuers struct {
	Enabled bool `json:"enabled"`
}

type shootDNSConfig struct {
	APIVersion                    string                      `json:"apiVersion"`
	Kind                          string                      `json:"kind"`
	DNSProviderReplication        *shootDNSProviderReplication `json:"dnsProviderReplication,omitempty"`
	Providers                     []shootDNSProvider           `json:"providers,omitempty"`
	SyncProvidersFromShootSpecDNS *bool                        `json:"syncProvidersFromShootSpecDNS,omitempty"`
}

type shootDNSProviderReplication struct {
	Enabled bool `json:"enabled"`
}

type shootDNSProvider struct {
	Domains     *shootDNSIncludeExclude `json:"domains,omitempty"`
	Credentials *string                 `json:"credentials,omitempty"`
	Type        *string                 `json:"type,omitempty"`
}

type shootDNSIncludeExclude struct {
	Include []string `json:"include,omitempty"`
	Exclude []string `json:"exclude,omitempty"`
}
