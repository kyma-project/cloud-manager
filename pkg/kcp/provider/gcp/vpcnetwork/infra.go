package vpcnetwork

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/elliotchance/pie/v2"
	gcpmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/meta"
	gcpvpcnetworkclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/vpcnetwork/client"
)

type CreateInfraOption func(*createInfraOptions)

func WithName(name string) CreateInfraOption {
	return func(o *createInfraOptions) {
		o.name = name
	}
}

func WithGcpProjectId(projectId string) CreateInfraOption {
	return func(o *createInfraOptions) {
		o.gcpProjectId = projectId
	}
}

func WithClient(c gcpvpcnetworkclient.Client) CreateInfraOption {
	return func(o *createInfraOptions) {
		o.client = c
	}
}

func WithRegion(region string) CreateInfraOption {
	return func(o *createInfraOptions) {
		o.region = region
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
	name         string
	gcpProjectId string
	region       string
	client       gcpvpcnetworkclient.Client
	timeout      time.Duration
	interval     time.Duration
}

func (o *createInfraOptions) validate() error {
	var result error
	if o.name == "" {
		result = errors.Join(result, fmt.Errorf("name is required"))
	}
	if o.gcpProjectId == "" {
		result = errors.Join(result, fmt.Errorf("gcpProjectId is required"))
	}
	if o.client == nil {
		result = errors.Join(result, fmt.Errorf("client is required"))
	}
	if o.region == "" {
		result = errors.Join(result, fmt.Errorf("region is required"))
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

type CreateInfraOutput struct {
	Created bool
	Network *computepb.Network
	Router  *computepb.Router
}

func CreateInfra(ctx context.Context, opts ...CreateInfraOption) (*CreateInfraOutput, error) {
	created := false
	o := &createInfraOptions{}
	for _, opt := range opts {
		opt(o)
	}
	if err := o.validate(); err != nil {
		return nil, err
	}

	// network

	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, o.timeout)
	defer cancel()

	net, err := o.client.GetNetwork(ctx, &computepb.GetNetworkRequest{
		Project: o.gcpProjectId,
		Network: o.name,
	})
	if err != nil && !gcpmeta.IsNotFound(err) {
		return nil, fmt.Errorf("error getting gcp network in create: %w", err)
	}
	if gcpmeta.IsNotFound(err) {
		op, err := o.client.InsertNetwork(ctx, &computepb.InsertNetworkRequest{
			Project: o.gcpProjectId,
			NetworkResource: &computepb.Network{
				Name:                  new(o.name),
				AutoCreateSubnetworks: new(false),
			},
		})
		if err != nil {
			return nil, fmt.Errorf("error inserting gcp network: %w", err)
		}
		// ~10s
		if err := op.Wait(ctx); err != nil {
			return nil, fmt.Errorf("error waiting insertion of gcp network: %w", err)
		}

		created = true

		net, err = o.client.GetNetwork(ctx, &computepb.GetNetworkRequest{
			Project: o.gcpProjectId,
			Network: o.name,
		})
		if err != nil {
			return nil, fmt.Errorf("error getting inserted gcp network: %w", err)
		}
	}

	// cloud router

	router, err := o.client.GetRouter(ctx, &computepb.GetRouterRequest{
		Project: o.gcpProjectId,
		Region:  o.region,
		Router:  o.name,
	})
	if err != nil && !gcpmeta.IsNotFound(err) {
		return nil, fmt.Errorf("error getting gcp router for create: %w", err)
	}
	if gcpmeta.IsNotFound(err) {
		op, err := o.client.InsertRouter(ctx, &computepb.InsertRouterRequest{
			Project: o.gcpProjectId,
			Region:  o.region,
			RouterResource: &computepb.Router{
				Name:    new(o.name),
				Network: new(net.GetSelfLink()),
			},
		})
		if err != nil {
			return nil, fmt.Errorf("error inserting gcp router: %w", err)
		}
		if err := op.Wait(ctx); err != nil {
			return nil, fmt.Errorf("error waiting insertion of gcp router: %w", err)
		}

		created = true

		router, err = o.client.GetRouter(ctx, &computepb.GetRouterRequest{
			Project: o.gcpProjectId,
			Region:  o.region,
			Router:  o.name,
		})
		if err != nil {
			return nil, fmt.Errorf("error getting inserted gcp router: %w", err)
		}
	}

	return &CreateInfraOutput{
		Created: created,
		Network: net,
		Router:  router,
	}, nil
}

func DeleteInfra(ctx context.Context, opts ...CreateInfraOption) error {
	o := &createInfraOptions{}
	for _, opt := range opts {
		opt(o)
	}

	if err := o.validate(); err != nil {
		return err
	}

	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, o.timeout)
	defer cancel()

	router, err := o.client.GetRouter(ctx, &computepb.GetRouterRequest{
		Project: o.gcpProjectId,
		Region:  o.region,
		Router:  o.name,
	})
	if err != nil && !gcpmeta.IsNotFound(err) {
		return fmt.Errorf("error getting gcp router for delete: %w", err)
	}
	if err == nil {
		// router exists

		if len(router.Nats) > 0 {
			var subnetsUsingRouter []string
			for _, nat := range router.Nats {
				for _, sub := range nat.Subnetworks {
					subnetsUsingRouter = append(subnetsUsingRouter, sub.GetName())
				}
			}
			return fmt.Errorf("gcp router %s can not be deleted, it has NATs in subnets: %v", router.GetSelfLink(), subnetsUsingRouter)
		}

		op, err := o.client.DeleteRouter(ctx, &computepb.DeleteRouterRequest{
			Project: o.gcpProjectId,
			Region:  o.region,
			Router:  router.GetName(),
		})
		if err != nil {
			return fmt.Errorf("error deleting gcp router: %w", err)
		}
		if err := op.Wait(ctx); err != nil {
			return fmt.Errorf("error waiting gcp router is deleted: %w", err)
		}
	}

	net, err := o.client.GetNetwork(ctx, &computepb.GetNetworkRequest{
		Project: o.gcpProjectId,
		Network: o.name,
	})
	if err != nil && !gcpmeta.IsNotFound(err) {
		return fmt.Errorf("error getting gcp network in delete: %w", err)
	}
	if err == nil {
		// network exists

		if len(net.Subnetworks) > 0 {
			return fmt.Errorf("gcp network %s can not be deleted, it has subnetworks: %v", net.GetSelfLink(), net.Subnetworks)
		}
		if len(net.Peerings) > 0 {
			return fmt.Errorf("gcp network %s can not be deleted, it has peerings to networks: %v", net.GetSelfLink(), pie.Map(net.Peerings, func(peering *computepb.NetworkPeering) string {
				return peering.GetNetwork()
			}))
		}

		op, err := o.client.DeleteNetwork(ctx, &computepb.DeleteNetworkRequest{
			Project: o.gcpProjectId,
			Network: net.GetName(),
		})
		if err != nil {
			return fmt.Errorf("error deleting gcp network: %w", err)
		}
		// ~45s
		if err := op.Wait(ctx); err != nil {
			return fmt.Errorf("error waiting gcp network is deleted: %w", err)
		}
	}

	return nil
}
