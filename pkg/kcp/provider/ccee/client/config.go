package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack"
	"github.com/gophercloud/gophercloud/v2/openstack/config"
	"os"
)

func NewProviderClient(_ context.Context, pp ProviderParams) (*gophercloud.ServiceClient, error) {
	authUrl := fmt.Sprintf("https://identity-3.%s.cloud.sap/v3/", pp.RegionName)
	authOptions := gophercloud.AuthOptions{
		IdentityEndpoint: authUrl,
		Username:         pp.Username,
		Password:         pp.Password,
		DomainName:       pp.ProjectDomainName,
		TenantName:       pp.ProjectName,
	}

	endpointOptions := gophercloud.EndpointOpts{
		Region:       pp.RegionName,
		Availability: gophercloud.AvailabilityPublic,
	}

	var tlsConfig *tls.Config
	if pp.TlsCaCertPath != "" {
		caCert, err := os.ReadFile(pp.TlsCaCertPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read tls ca cert file: %w", err)
		}
		caCertPool := x509.NewCertPool()
		if ok := caCertPool.AppendCertsFromPEM(bytes.TrimSpace(caCert)); !ok {
			return nil, fmt.Errorf("failed to parse the CA Cert from %q", pp.TlsCaCertPath)
		}
		tlsConfig = new(tls.Config)
		tlsConfig.RootCAs = caCertPool
	}
	if pp.TlsCertPath != "" && pp.TlsKeyPath != "" {
		clientCert, err := os.ReadFile(pp.TlsCertPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read the client cert file: %w", err)
		}

		clientKey, err := os.ReadFile(pp.TlsKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read the client cert key file: %w", err)
		}

		cert, err := tls.X509KeyPair(clientCert, clientKey)
		if err != nil {
			return nil, fmt.Errorf("failed to parse the client cert and client key pair: %w", err)
		}

		if tlsConfig == nil {
			tlsConfig = new(tls.Config)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	} else if pp.TlsCertPath != "" && pp.TlsKeyPath == "" {
		return nil, fmt.Errorf("client cert is set, but client cert key is missing")
	} else if pp.TlsCertPath == "" && pp.TlsKeyPath != "" {
		return nil, fmt.Errorf("client cert key is set, but client cert is missing")
	}

	var providerClient *gophercloud.ProviderClient
	var err error
	httpClient := monitoredHttpClient()
	if tlsConfig != nil {
		providerClient, err = config.NewProviderClient(context.Background(), authOptions, config.WithHTTPClient(*httpClient), config.WithTLSConfig(tlsConfig))
	} else {
		providerClient, err = config.NewProviderClient(context.Background(), authOptions, config.WithHTTPClient(*httpClient))
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create cloud client: %v", err)
	}

	serviceClient, err := openstack.NewComputeV2(providerClient, endpointOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create cloud client: %v", err)
	}

	return serviceClient, nil
}
