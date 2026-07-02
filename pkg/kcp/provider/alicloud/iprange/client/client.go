package client

import (
	"context"
	"fmt"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/tea"
	vpc "github.com/alibabacloud-go/vpc-20160428/v6/client"
)

type VSwitchInfo struct {
	VSwitchId   string
	VSwitchName string
	CidrBlock   string
	VpcId       string
	ZoneId      string
	Status      string
}

type Client interface {
	CreateVSwitch(ctx context.Context, vpcId, zoneId, cidrBlock, name string) (string, error)
	DescribeVSwitch(ctx context.Context, vSwitchId string) (*VSwitchInfo, error)
	DescribeVSwitchesByName(ctx context.Context, vpcId, name string) ([]VSwitchInfo, error)
	DeleteVSwitch(ctx context.Context, vSwitchId string) error
	DescribeVpcs(ctx context.Context, name string) ([]VpcInfo, error)
	DescribeZones(ctx context.Context) ([]string, error)
}

type VpcInfo struct {
	VpcId     string
	VpcName   string
	CidrBlock string
	Status    string
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

func (c *alicloudClient) CreateVSwitch(ctx context.Context, vpcId, zoneId, cidrBlock, name string) (string, error) {
	req := &vpc.CreateVSwitchRequest{
		RegionId:    new(c.region),
		VpcId:       new(vpcId),
		ZoneId:      new(zoneId),
		CidrBlock:   new(cidrBlock),
		VSwitchName: new(name),
	}

	resp, err := c.vpcClient.CreateVSwitch(req)
	if err != nil {
		return "", fmt.Errorf("error creating alicloud vswitch: %w", err)
	}

	return tea.StringValue(resp.Body.VSwitchId), nil
}

func (c *alicloudClient) DescribeVSwitch(ctx context.Context, vSwitchId string) (*VSwitchInfo, error) {
	req := &vpc.DescribeVSwitchesRequest{
		RegionId:  new(c.region),
		VSwitchId: new(vSwitchId),
	}

	resp, err := c.vpcClient.DescribeVSwitches(req)
	if err != nil {
		return nil, fmt.Errorf("error describing alicloud vswitch %s: %w", vSwitchId, err)
	}

	if resp.Body == nil || resp.Body.VSwitches == nil || len(resp.Body.VSwitches.VSwitch) == 0 {
		return nil, nil
	}

	v := resp.Body.VSwitches.VSwitch[0]
	return &VSwitchInfo{
		VSwitchId:   tea.StringValue(v.VSwitchId),
		VSwitchName: tea.StringValue(v.VSwitchName),
		CidrBlock:   tea.StringValue(v.CidrBlock),
		VpcId:       tea.StringValue(v.VpcId),
		ZoneId:      tea.StringValue(v.ZoneId),
		Status:      tea.StringValue(v.Status),
	}, nil
}

func (c *alicloudClient) DescribeVSwitchesByName(ctx context.Context, vpcId, name string) ([]VSwitchInfo, error) {
	req := &vpc.DescribeVSwitchesRequest{
		RegionId:    new(c.region),
		VpcId:       new(vpcId),
		VSwitchName: new(name),
	}

	resp, err := c.vpcClient.DescribeVSwitches(req)
	if err != nil {
		return nil, fmt.Errorf("error describing alicloud vswitches by name: %w", err)
	}

	var result []VSwitchInfo
	if resp.Body != nil && resp.Body.VSwitches != nil {
		for _, v := range resp.Body.VSwitches.VSwitch {
			result = append(result, VSwitchInfo{
				VSwitchId:   tea.StringValue(v.VSwitchId),
				VSwitchName: tea.StringValue(v.VSwitchName),
				CidrBlock:   tea.StringValue(v.CidrBlock),
				VpcId:       tea.StringValue(v.VpcId),
				ZoneId:      tea.StringValue(v.ZoneId),
				Status:      tea.StringValue(v.Status),
			})
		}
	}

	return result, nil
}

func (c *alicloudClient) DeleteVSwitch(ctx context.Context, vSwitchId string) error {
	req := &vpc.DeleteVSwitchRequest{
		RegionId:  new(c.region),
		VSwitchId: new(vSwitchId),
	}

	_, err := c.vpcClient.DeleteVSwitch(req)
	if err != nil {
		return fmt.Errorf("error deleting alicloud vswitch %s: %w", vSwitchId, err)
	}

	return nil
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

func (c *alicloudClient) DescribeZones(ctx context.Context) ([]string, error) {
	req := &vpc.DescribeZonesRequest{
		RegionId: new(c.region),
	}

	resp, err := c.vpcClient.DescribeZones(req)
	if err != nil {
		return nil, fmt.Errorf("error describing alicloud zones: %w", err)
	}

	var result []string
	if resp.Body != nil && resp.Body.Zones != nil {
		for _, z := range resp.Body.Zones.Zone {
			result = append(result, tea.StringValue(z.ZoneId))
		}
	}

	return result, nil
}
