package client

import (
	"cloud.google.com/go/compute/apiv1/computepb"
	"context"
	"fmt"
	"k8s.io/utils/ptr"

	compute "cloud.google.com/go/compute/apiv1"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
)

type Client interface {
}

func NewClientProvider(gcpClients *gcpclient.GcpClients) gcpclient.GcpClientProvider[Client] {
	return func() Client {
		return &client{routersClient: gcpClients.ComputeRouters}
	}
}

type client struct {
	routersClient *compute.RoutersClient
}

func (c *client) GetVpcRouters(ctx context.Context, project string, region string, vpcName string) ([]*computepb.Router, error) {
	it := c.routersClient.List(ctx, &computepb.ListRoutersRequest{
		Filter:  ptr.To(fmt.Sprintf("network = %s", vpcName)),
		Project: project,
		Region:  region,
	}).All()
	var results []*computepb.Router
	for r, err := range it {
		if err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, nil
}
