package vpcnetwork

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/3th1nk/cidr"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/elliotchance/pie/v2"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	azurevpcnetworkclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/vpcnetwork/client"
	"k8s.io/utils/ptr"
)

type CreateInfraOption func(*createInfraOptions)

func WithName(name string) CreateInfraOption {
	return func(o *createInfraOptions) {
		o.name = name
	}
}

func WithLocation(location string) CreateInfraOption {
	return func(o *createInfraOptions) {
		o.location = location
	}
}

func WithCidrBlocks(cidrBlocks []string) CreateInfraOption {
	return func(o *createInfraOptions) {
		o.cidrBlocks = append(o.cidrBlocks, cidrBlocks...)
	}
}

func WithClient(c azurevpcnetworkclient.Client) CreateInfraOption {
	return func(o *createInfraOptions) {
		o.client = c
	}
}

func WithTimeout(t time.Duration) CreateInfraOption {
	return func(o *createInfraOptions) {
		o.timeout = t
	}
}

type createInfraOptions struct {
	name       string
	location   string
	cidrBlocks []string
	client     azurevpcnetworkclient.Client
	timeout    time.Duration
}

func (o *createInfraOptions) validate() error {
	var result error
	if o.name == "" {
		result = errors.Join(result, fmt.Errorf("name is required"))
	}
	if o.location == "" {
		result = errors.Join(result, fmt.Errorf("location is required"))
	}
	if len(o.cidrBlocks) == 0 {
		result = errors.Join(result, fmt.Errorf("at least one cidr block is required"))
	}
	for _, c := range o.cidrBlocks {
		_, err := cidr.Parse(c)
		if err != nil {
			result = errors.Join(result, fmt.Errorf("invalid cidr block %q: %w", c, err))
		}
	}
	if o.client == nil {
		result = errors.Join(result, fmt.Errorf("client is required"))
	}
	if o.timeout == 0 {
		o.timeout = 5 * time.Minute
	}
	return result
}

type CreateInfraOutput struct {
	Created        bool
	Updated        bool
	ResourceGroup  *armresources.ResourceGroup
	VirtualNetwork *armnetwork.VirtualNetwork
}

func CreateInfra(ctx context.Context, opts ...CreateInfraOption) (*CreateInfraOutput, error) {
	created := false
	updated := false
	o := &createInfraOptions{}
	for _, opt := range opts {
		opt(o)
	}
	if err := o.validate(); err != nil {
		return nil, err
	}

	// resource group

	resourceGroup, err := o.client.GetResourceGroup(ctx, o.name)
	if azuremeta.IsNotFound(err) {
		rg, err := o.client.CreateResourceGroup(ctx, o.name, o.location, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create resource group %q: %w", o.name, err)
		}
		created = true
		resourceGroup = rg
	}

	// virtual network

	virtualNetwork, err := o.client.GetNetwork(ctx, o.name, o.name)
	if azuremeta.IgnoreNotFoundError(err) != nil {
		return nil, fmt.Errorf("failed to get network: %w", err)
	}
	if azuremeta.IsNotFound(err) {
		vnetIn := armnetwork.VirtualNetwork{
			Location: ptr.To(o.location),
			Properties: &armnetwork.VirtualNetworkPropertiesFormat{
				AddressSpace: &armnetwork.AddressSpace{
					AddressPrefixes: pie.Map(o.cidrBlocks, func(x string) *string {
						return ptr.To(x)
					}),
				},
			},
		}
		poller, err := o.client.CreateOrUpdateNetwork(ctx, o.name, o.name, vnetIn, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to begin create network %q: %w", o.name, err)
		}
		vnetOut, err := func() (*armnetwork.VirtualNetwork, error) {
			toCtx, cancel := context.WithTimeout(ctx, o.timeout)
			defer cancel()
			resp, err := poller.PollUntilDone(toCtx, nil)
			if err != nil {
				return nil, fmt.Errorf("error polling network %q: %w", o.name, err)
			}
			return &resp.VirtualNetwork, nil
		}()
		if err != nil {
			return nil, err
		}
		virtualNetwork = vnetOut
	}

	// validate primary cidr block didn't change
	if ptr.Deref(virtualNetwork.Properties.AddressSpace.AddressPrefixes[0], "") != o.cidrBlocks[0] {
		return nil, fmt.Errorf("primary cidr block can not change - %s was changed to %s", ptr.Deref(virtualNetwork.Properties.AddressSpace.AddressPrefixes[0], ""), o.cidrBlocks[0])
	}

	updateNeeded := false

	if len(virtualNetwork.Properties.AddressSpace.AddressPrefixes) != len(o.cidrBlocks) {
		updateNeeded = true
	} else {
		existingCidrs := pie.Map(virtualNetwork.Properties.AddressSpace.AddressPrefixes, func(x *string) string {
			return ptr.Deref(x, "")
		})
		added, removed := pie.Diff(existingCidrs, o.cidrBlocks)
		if len(added) != 0 && len(removed) != 0 {
			updateNeeded = true
		}
	}

	if updateNeeded {
		poller, err := o.client.CreateOrUpdateNetwork(ctx, o.name, o.name, *virtualNetwork, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to begin update network: %w", err)
		}
		vnetOut, err := func() (*armnetwork.VirtualNetwork, error) {
			toCtx, cancel := context.WithTimeout(ctx, o.timeout)
			defer cancel()
			resp, err := poller.PollUntilDone(toCtx, nil)
			if err != nil {
				return nil, fmt.Errorf("error polling network %q: %w", o.name, err)
			}
			return &resp.VirtualNetwork, nil
		}()
		if err != nil {
			return nil, fmt.Errorf("failed to poll update network: %w", err)
		}
		virtualNetwork = vnetOut
		updated = true
	}

	return &CreateInfraOutput{
		Created:        created,
		Updated:        updated,
		ResourceGroup:  resourceGroup,
		VirtualNetwork: virtualNetwork,
	}, nil
}

func DeleteInfra(ctx context.Context, name string, c azurevpcnetworkclient.Client) error {
	vnetPoller, err := c.DeleteNetwork(ctx, name, name, nil)
	if azuremeta.IgnoreNotFoundError(err) != nil {
		return fmt.Errorf("failed to delete network: %w", err)
	}
	if err == nil {
		_, err = vnetPoller.PollUntilDone(ctx, nil)
		if err != nil {
			return fmt.Errorf("error polling network deletion: %w", err)
		}
	}

	err = c.DeleteResourceGroup(ctx, name)
	if azuremeta.IgnoreNotFoundError(err) != nil {
		return fmt.Errorf("failed to delete resource group: %w", err)
	}

	return nil
}
