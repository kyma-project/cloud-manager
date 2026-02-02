package vpcnetwork

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"testing"
	"time"

	"github.com/3th1nk/cidr"
	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/extensions/layer3/routers"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/networks"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/subnets"
	sapconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/config"
	sapvpcnetworkclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/vpcnetwork/client"
	"k8s.io/apimachinery/pkg/util/wait"
)

type CreateInfraOption func(*createInfraOptions)

func WithName(name string) CreateInfraOption {
	return func(o *createInfraOptions) {
		o.name = name
	}
}

func WithCidrBlocks(cidrBlocks []string) CreateInfraOption {
	return func(o *createInfraOptions) {
		o.cidrBlocks = append(o.cidrBlocks, cidrBlocks...)
	}
}

func WithClient(c sapvpcnetworkclient.Client) CreateInfraOption {
	return func(o *createInfraOptions) {
		o.client = c
	}
}

func WithTimeout(t time.Duration) CreateInfraOption {
	return func(o *createInfraOptions) {
		o.timeout = t
	}
}

func WithInterval(t time.Duration) CreateInfraOption {
	return func(o *createInfraOptions) {
		o.interval = t
	}
}

type createInfraOptions struct {
	name       string
	cidrBlocks []string
	client     sapvpcnetworkclient.Client
	timeout    time.Duration
	interval   time.Duration
}

type CreateInfraOutput struct {
	Created         bool
	Updated         bool
	Router          *routers.Router
	InternalNetwork *networks.Network
}

func (o *createInfraOptions) validate() error {
	var result error
	if o.name == "" {
		result = errors.Join(result, fmt.Errorf("name is required"))
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
		if testing.Testing() {
			o.timeout = time.Second
		}
	}
	if o.interval == 0 {
		o.interval = time.Second
		if testing.Testing() {
			o.interval = time.Millisecond
		}
	}
	return result
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

	// load external FIP network

	if sapconfig.SapConfig.FloatingPoolNetwork == "" {
		return nil, errors.New("floating pool network is not configured")
	}
	if sapconfig.SapConfig.FloatingPoolSubnet == "" {
		return nil, errors.New("floating pool subnet is not configured")
	}

	arrFipNets, err := o.client.ListExternalNetworks(ctx, networks.ListOpts{
		Name: sapconfig.SapConfig.FloatingPoolNetwork,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list external fip networks: %w", err)
	}
	if len(arrFipNets) == 0 {
		return nil, errors.New("no external networks found matching configured floating pool network name")
	}

	externalNetwork := arrFipNets[0]

	// load external FIP subnet

	arrFipSubnets, err := o.client.ListSubnets(ctx, subnets.ListOpts{
		NetworkID: externalNetwork.ID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list fip subnets: %w", err)
	}
	var externalSubnetCandidateIds []string
	externalSubnetRegex, err := regexp.Compile(sapconfig.SapConfig.FloatingPoolSubnet)
	if err != nil {
		return nil, fmt.Errorf("failed to parse floating pool subnet regexp: %w", err)
	}
	for _, s := range arrFipSubnets {
		if externalSubnetRegex.MatchString(s.Name) {
			externalSubnetCandidateIds = append(externalSubnetCandidateIds, s.ID)
		}
	}
	if len(externalSubnetCandidateIds) == 0 {
		return nil, errors.New("no external subnet found matching configured floating pool subnet name")
	}

	// load interval network

	internalNetwork, err := o.client.GetNetworkByName(ctx, o.name)
	if err != nil {
		return nil, fmt.Errorf("failed to get internal network by name: %w", err)
	}

	// load router

	router, err := o.client.GetRouterByName(ctx, o.name)
	if err != nil {
		return nil, fmt.Errorf("failed to get router by name: %w", err)
	}

	// router create/update ============================================================

	if router == nil {

		// create router ----------------------------------------------

		isCreated := false

		var allCreateRouterErrors []error
		for _, externalSubnetCandidateId := range externalSubnetCandidateIds {
			router, err = o.client.CreateRouter(ctx, routers.CreateOpts{
				Name:         o.name,
				AdminStateUp: gophercloud.Enabled,
				GatewayInfo: &routers.GatewayInfo{
					NetworkID: externalNetwork.ID,
					ExternalFixedIPs: []routers.ExternalFixedIP{
						{
							SubnetID: externalSubnetCandidateId,
						},
					},
				},
			})
			if err != nil {
				allCreateRouterErrors = append(allCreateRouterErrors, err)
				continue
			}
			isCreated = true
			break
		}

		if !isCreated {
			return nil, fmt.Errorf("failed to create router: %w", errors.Join(allCreateRouterErrors...))
		}

		created = true

	} else {

		// update router ----------------------------------------------

		updateNeeded := false

		isSubnetOK := false
		for _, ext := range router.GatewayInfo.ExternalFixedIPs {
			if slices.Contains(externalSubnetCandidateIds, ext.SubnetID) {
				isSubnetOK = true
				break
			}
		}
		if !isSubnetOK {
			updateNeeded = true
		}

		if updateNeeded {
			var allUpdateRouterErrors []error
			isUpdated := false
			for _, externalSubnetCandidateId := range externalSubnetCandidateIds {
				_, err = o.client.UpdateRouter(ctx, router.ID, routers.UpdateOpts{
					GatewayInfo: &routers.GatewayInfo{
						NetworkID: externalNetwork.ID,
						ExternalFixedIPs: []routers.ExternalFixedIP{
							{
								SubnetID: externalSubnetCandidateId,
							},
						},
					},
				})
				if err != nil {
					allUpdateRouterErrors = append(allUpdateRouterErrors, err)
					continue
				}
				isUpdated = true
				break
			}
			if !isUpdated {
				return nil, fmt.Errorf("failed to update router: %w", errors.Join(allUpdateRouterErrors...))
			}

			updated = true

		} // if updateNeeded

	} // else router != nil - check if update is needed

	if created || updated {
		// wait router is active

		err := wait.PollUntilContextTimeout(ctx, o.interval, o.timeout, false, func(ctx context.Context) (done bool, err error) {
			r, err := o.client.GetRouterByName(ctx, o.name)
			if err != nil {
				return false, err
			}
			if r == nil {
				return false, fmt.Errorf("router %q not found", o.name)
			}
			if r.Status == "ACTIVE" {
				router = r
				return true, nil
			}
			if r.Status == "ERROR" {
				return false, fmt.Errorf("router is in error state")
			}
			return false, nil
		})
		if err != nil {
			return nil, fmt.Errorf("error waiting router to become active: %w", err)
		}
	}

	// interval network create/update ============================================================

	if internalNetwork == nil {

		// create interval network

		internalNetwork, err = o.client.CreateNetwork(ctx, networks.CreateOpts{
			Name:         o.name,
			AdminStateUp: gophercloud.Enabled,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create internal network: %w", err)
		}

		created = true

		// wait interval network is active

		err := wait.PollUntilContextTimeout(ctx, o.interval, o.timeout, false, func(ctx context.Context) (done bool, err error) {
			net, err := o.client.GetNetworkByName(ctx, o.name)
			if err != nil {
				return false, err
			}
			if net.Status == networks.StatusActive {
				internalNetwork = net
				return true, nil
			}
			if net.Status == networks.StatusError {
				return false, fmt.Errorf("internal network is in error state")
			}
			if net.Status == networks.StatusDown {
				return false, fmt.Errorf("internal network is down")
			}
			return false, nil
		})
		if err != nil {
			return nil, fmt.Errorf("error waiting internal network to become active: %w", err)
		}
	}

	return &CreateInfraOutput{
		Created:         created,
		Updated:         updated,
		Router:          router,
		InternalNetwork: internalNetwork,
	}, nil
}

func DeleteInfra(ctx context.Context, name string, c sapvpcnetworkclient.Client) error {
	internalNetwork, err := c.GetNetworkByName(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to get internal network by name: %w", err)
	}

	if internalNetwork != nil {
		err = c.DeleteNetwork(ctx, internalNetwork.ID)
		if err != nil {
			return fmt.Errorf("failed to delete internal network: %w", err)
		}
	}

	router, err := c.GetRouterByName(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to get router by name: %w", err)
	}
	if router != nil {
		err = c.DeleteRouter(ctx, router.ID)
		if err != nil {
			return fmt.Errorf("failed to delete router: %w", err)
		}
	}

	return nil
}
