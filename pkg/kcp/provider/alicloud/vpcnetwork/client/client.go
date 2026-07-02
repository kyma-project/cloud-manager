package client

import (
	"context"
	"fmt"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/tea"
	vpc "github.com/alibabacloud-go/vpc-20160428/v6/client"
)

type VpcInfo struct {
	VpcId     string
	VpcName   string
	CidrBlock string
	Status    string
}

type Client interface {
	CreateVpc(ctx context.Context, name string, cidrBlock string) (*VpcInfo, error)
	DescribeVpcs(ctx context.Context, name string) ([]VpcInfo, error)
	DeleteVpc(ctx context.Context, vpcId string) error
}

type ClientProvider func(ctx context.Context, region, accessKeyId, accessKeySecret string) (Client, error)

func NewClientProvider() ClientProvider {
	return func(ctx context.Context, region, accessKeyId, accessKeySecret string) (Client, error) {
		config := &openapi.Config{
			AccessKeyId:     new(accessKeyId),
			AccessKeySecret: new(accessKeySecret),
			RegionId:        new(region),
		}
		config.Endpoint = new(fmt.Sprintf("vpc.%s.aliyuncs.com", region))

		vpcClient, err := vpc.NewClient(config)
		if err != nil {
			return nil, fmt.Errorf("error creating alicloud vpc client: %w", err)
		}

		return &alicloudClient{
			vpcClient: vpcClient,
			region:    region,
		}, nil
	}
}

type alicloudClient struct {
	vpcClient *vpc.Client
	region    string
}

func (c *alicloudClient) CreateVpc(ctx context.Context, name string, cidrBlock string) (*VpcInfo, error) {
	req := &vpc.CreateVpcRequest{
		RegionId:  new(c.region),
		VpcName:   new(name),
		CidrBlock: new(cidrBlock),
	}

	resp, err := c.vpcClient.CreateVpc(req)
	if err != nil {
		return nil, fmt.Errorf("error creating alicloud vpc: %w", err)
	}

	return &VpcInfo{
		VpcId:     tea.StringValue(resp.Body.VpcId),
		VpcName:   name,
		CidrBlock: cidrBlock,
		Status:    "Pending",
	}, nil
}

func (c *alicloudClient) DescribeVpcs(ctx context.Context, name string) ([]VpcInfo, error) {
	req := &vpc.DescribeVpcsRequest{
		RegionId: new(c.region),
		VpcName:  new(name),
	}

	resp, err := c.vpcClient.DescribeVpcs(req)
	if err != nil {
		return nil, fmt.Errorf("error describing alicloud vpcs: %w", err)
	}

	var result []VpcInfo
	if resp.Body != nil && resp.Body.Vpcs != nil {
		for _, v := range resp.Body.Vpcs.Vpc {
			result = append(result, VpcInfo{
				VpcId:     tea.StringValue(v.VpcId),
				VpcName:   tea.StringValue(v.VpcName),
				CidrBlock: tea.StringValue(v.CidrBlock),
				Status:    tea.StringValue(v.Status),
			})
		}
	}

	return result, nil
}

func (c *alicloudClient) DeleteVpc(ctx context.Context, vpcId string) error {
	req := &vpc.DeleteVpcRequest{
		RegionId: new(c.region),
		VpcId:    new(vpcId),
	}

	_, err := c.vpcClient.DeleteVpc(req)
	if err != nil {
		return fmt.Errorf("error deleting alicloud vpc %s: %w", vpcId, err)
	}

	return nil
}
