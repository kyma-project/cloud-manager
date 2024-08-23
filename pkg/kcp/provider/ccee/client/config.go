package client

import (
	"context"
	"fmt"
	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack/config"
)

type ProvidedInfo struct {
	ProviderClient  *gophercloud.ProviderClient
	EndpointOptions gophercloud.EndpointOpts
}

func NewProviderClient(_ context.Context, pp ProviderParams) (*ProvidedInfo, error) {
	authUrl := fmt.Sprintf("https://identity-3.%s.cloud.sap/v3/", pp.RegionName)
	authOptions := gophercloud.AuthOptions{
		IdentityEndpoint: authUrl,
		Username:         pp.Username,
		Password:         pp.Password,
		DomainName:       pp.ProjectDomainName,
		TenantName:       pp.ProjectName,
		AllowReauth:      true,
	}

	endpointOptions := gophercloud.EndpointOpts{
		Region:       pp.RegionName,
		Availability: gophercloud.AvailabilityPublic,
	}

	var providerClient *gophercloud.ProviderClient
	var err error
	httpClient := monitoredHttpClient()
	providerClient, err = config.NewProviderClient(context.Background(), authOptions, config.WithHTTPClient(*httpClient))

	if err != nil {
		return nil, fmt.Errorf("failed to create cloud client: %v", err)
	}

	return &ProvidedInfo{
		ProviderClient:  providerClient,
		EndpointOptions: endpointOptions,
	}, nil
	//serviceClient, err := openstack.NewComputeV2(providerClient, endpointOptions)
	//if err != nil {
	//	return nil, fmt.Errorf("failed to create cloud client: %v", err)
	//}
	//
	//return serviceClient, nil
}
