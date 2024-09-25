package client

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"k8s.io/utils/ptr"
)

type ResourceGroupClient interface {
	GetResourceGroup(ctx context.Context, name string) (*armresources.ResourceGroup, error)
	CreateResourceGroup(ctx context.Context, name string, location string, tags map[string]string) (*armresources.ResourceGroup, error)
	DeleteResourceGroup(ctx context.Context, name string) error
}

func NewResourceGroupClient(svc *armresources.ResourceGroupsClient) ResourceGroupClient {
	return &resourceGroupClient{svc: svc}
}

var _ ResourceGroupClient = &resourceGroupClient{}

type resourceGroupClient struct {
	svc *armresources.ResourceGroupsClient
}

func (c *resourceGroupClient) GetResourceGroup(ctx context.Context, name string) (*armresources.ResourceGroup, error) {
	resp, err := c.svc.Get(ctx, name, nil)
	if err != nil {
		return nil, err
	}
	return &resp.ResourceGroup, nil
}

func (c *resourceGroupClient) CreateResourceGroup(ctx context.Context, name string, location string, tags map[string]string) (*armresources.ResourceGroup, error) {
	var azureTags map[string]*string
	if tags != nil {
		azureTags = make(map[string]*string, len(tags))
		for k, v := range tags {
			azureTags[k] = ptr.To(v)
		}
	}
	resp, err := c.svc.CreateOrUpdate(ctx, name, armresources.ResourceGroup{
		Location: ptr.To(location),
		Tags:     azureTags,
	}, nil)
	if err != nil {
		return nil, err
	}
	return &resp.ResourceGroup, nil
}

func (c *resourceGroupClient) DeleteResourceGroup(ctx context.Context, name string) error {
	_, err := c.svc.BeginDelete(ctx, name, nil)
	return err
}
