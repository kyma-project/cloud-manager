package client

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	"k8s.io/utils/ptr"
)

type SecurityGroupsClient interface {
	GetSecurityGroup(ctx context.Context, resourceGroupName, networkSecurityGroupName string) (*armnetwork.SecurityGroup, error)
	CreateSecurityGroup(ctx context.Context, resourceGroupName, networkSecurityGroupName, location string, securityRules []*armnetwork.SecurityRule, tags map[string]string) error
	DeleteSecurityGroup(ctx context.Context, resourceGroupName, networkSecurityGroupName string) error
}

func NewSecurityGroupsClient(svc *armnetwork.SecurityGroupsClient) SecurityGroupsClient {
	return &securityGroupsClient{svc: svc}
}

type securityGroupsClient struct {
	svc *armnetwork.SecurityGroupsClient
}

func (c *securityGroupsClient) GetSecurityGroup(ctx context.Context, resourceGroupName, networkSecurityGroupName string) (*armnetwork.SecurityGroup, error) {
	resp, err := c.svc.Get(ctx, resourceGroupName, networkSecurityGroupName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.SecurityGroup, nil
}

func (c *securityGroupsClient) CreateSecurityGroup(ctx context.Context, resourceGroupName, networkSecurityGroupName, location string, securityRules []*armnetwork.SecurityRule, tags map[string]string) error {
	var azureTags map[string]*string
	if tags != nil {
		azureTags = make(map[string]*string, len(tags))
		for k, v := range tags {
			azureTags[k] = ptr.To(v)
		}
	}
	sg := armnetwork.SecurityGroup{
		Location: ptr.To(location),
		Tags:     azureTags,
	}
	if securityRules != nil {
		sg.Properties = &armnetwork.SecurityGroupPropertiesFormat{
			SecurityRules: securityRules,
		}
	}
	_, err := c.svc.BeginCreateOrUpdate(ctx, resourceGroupName, networkSecurityGroupName, sg, nil)
	return err
}

func (c *securityGroupsClient) DeleteSecurityGroup(ctx context.Context, resourceGroupName, networkSecurityGroupName string) error {
	_, err := c.svc.BeginDelete(ctx, resourceGroupName, networkSecurityGroupName, nil)
	return err
}
