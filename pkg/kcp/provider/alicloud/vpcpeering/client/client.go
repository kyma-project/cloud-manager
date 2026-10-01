package client

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	openapiv1 "github.com/alibabacloud-go/darabonba-openapi/client"
	openapiv2 "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/tea"
	vpc "github.com/alibabacloud-go/vpc-20160428/v6/client"
	vpcpeer "github.com/alibabacloud-go/vpcpeer-20220101/client"
	alicloudmetrics "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/metrics"
)

// VpcPeerInfo is a flattened view of a VpcPeerConnection, unwrapping nested SDK structs.
type VpcPeerInfo struct {
	InstanceId        string
	Status            string // Creating, Accepting, Activated, Deleting, Deleted
	RegionId          string
	AcceptingRegionId string
	LocalVpcId        string   // from Vpc.VpcId
	RemoteVpcId       string   // from AcceptingVpc.VpcId
	RemoteIpv4Cidrs   []string // from AcceptingVpc.Ipv4Cidrs
	LocalIpv4Cidrs    []string // from Vpc.Ipv4Cidrs
}

// RouteTableInfo is a flattened view of a route table entry.
type RouteTableInfo struct {
	RouteTableId string
	Tags         map[string]string
}

type Client interface {
	CreateVpcPeerConnection(ctx context.Context, vpcId, acceptingAliUid, acceptingRegion, acceptingVpcId, name string) (instanceId string, err error)
	AcceptVpcPeerConnection(ctx context.Context, instanceId string) error
	GetVpcPeerConnection(ctx context.Context, instanceId string) (*VpcPeerInfo, error)
	DeleteVpcPeerConnection(ctx context.Context, instanceId string) error
	ListVpcPeerConnections(ctx context.Context, vpcId, name string) ([]VpcPeerInfo, error)
	DescribeRouteTables(ctx context.Context, vpcId string) ([]RouteTableInfo, error)
	CreateRouteEntry(ctx context.Context, routeTableId, destCidr, instanceId string) error
	DeleteRouteEntry(ctx context.Context, routeTableId, destCidr, instanceId string) error
}

type ClientProvider func(ctx context.Context, region, accessKeyId, accessKeySecret string) (Client, error)

func NewClientProvider() ClientProvider {
	return func(ctx context.Context, region, accessKeyId, accessKeySecret string) (Client, error) {
		// vpcpeer-20220101 uses darabonba-openapi v1 and a central endpoint (no per-region URL).
		// NOTE: openapiv1.Config does not support a custom HttpClient, so vpcpeer-20220101
		// API calls (CreateVpcPeerConnection, AcceptVpcPeerConnection, etc.) are not tracked
		// by alicloudmetrics. vpc-20160428 route calls below are tracked.
		peerConfig := &openapiv1.Config{
			AccessKeyId:     tea.String(accessKeyId),
			AccessKeySecret: tea.String(accessKeySecret),
			RegionId:        tea.String(region),
		}
		peerClient, err := vpcpeer.NewClient(peerConfig)
		if err != nil {
			return nil, fmt.Errorf("error creating alicloud vpcpeer client: %w", err)
		}

		// vpc-20160428/v6 uses darabonba-openapi v2
		vpcConfig := &openapiv2.Config{
			AccessKeyId:     tea.String(accessKeyId),
			AccessKeySecret: tea.String(accessKeySecret),
			RegionId:        tea.String(region),
			Endpoint:        tea.String(fmt.Sprintf("vpc.%s.aliyuncs.com", region)),
			HttpClient:      alicloudmetrics.NewMetricsHTTPClient(region, alicloudmetrics.AccountIdFromContext(ctx)),
		}
		vpcClient, err := vpc.NewClient(vpcConfig)
		if err != nil {
			return nil, fmt.Errorf("error creating alicloud vpc client: %w", err)
		}

		return &alicloudVpcPeeringClient{
			peerClient: peerClient,
			vpcClient:  vpcClient,
			region:     region,
		}, nil
	}
}

var _ Client = (*alicloudVpcPeeringClient)(nil)

type alicloudVpcPeeringClient struct {
	peerClient *vpcpeer.Client
	vpcClient  *vpc.Client
	region     string
}

func (c *alicloudVpcPeeringClient) CreateVpcPeerConnection(ctx context.Context, vpcId, acceptingAliUid, acceptingRegion, acceptingVpcId, name string) (string, error) {
	uid, err := strconv.ParseInt(acceptingAliUid, 10, 64)
	if err != nil {
		return "", fmt.Errorf("error parsing acceptingAliUid %q: %w", acceptingAliUid, err)
	}
	req := &vpcpeer.CreateVpcPeerConnectionRequest{
		RegionId:          tea.String(c.region),
		VpcId:             tea.String(vpcId),
		AcceptingAliUid:   tea.Int64(uid),
		AcceptingRegionId: tea.String(acceptingRegion),
		AcceptingVpcId:    tea.String(acceptingVpcId),
		Name:              tea.String(name),
	}
	resp, err := c.peerClient.CreateVpcPeerConnection(req)
	if err != nil {
		return "", fmt.Errorf("error creating alicloud vpc peer connection: %w", err)
	}
	return tea.StringValue(resp.Body.InstanceId), nil
}

func (c *alicloudVpcPeeringClient) AcceptVpcPeerConnection(ctx context.Context, instanceId string) error {
	req := &vpcpeer.AcceptVpcPeerConnectionRequest{
		InstanceId: tea.String(instanceId),
	}
	_, err := c.peerClient.AcceptVpcPeerConnection(req)
	if err != nil {
		return fmt.Errorf("error accepting alicloud vpc peer connection %s: %w", instanceId, err)
	}
	return nil
}

func (c *alicloudVpcPeeringClient) GetVpcPeerConnection(ctx context.Context, instanceId string) (*VpcPeerInfo, error) {
	req := &vpcpeer.GetVpcPeerConnectionAttributeRequest{
		InstanceId: tea.String(instanceId),
	}
	resp, err := c.peerClient.GetVpcPeerConnectionAttribute(req)
	if err != nil {
		return nil, fmt.Errorf("error getting alicloud vpc peer connection %s: %w", instanceId, err)
	}
	if resp.Body == nil {
		return nil, nil
	}
	b := resp.Body
	info := &VpcPeerInfo{
		InstanceId:        tea.StringValue(b.InstanceId),
		Status:            tea.StringValue(b.Status),
		RegionId:          tea.StringValue(b.RegionId),
		AcceptingRegionId: tea.StringValue(b.AcceptingRegionId),
	}
	if b.Vpc != nil {
		info.LocalVpcId = tea.StringValue(b.Vpc.VpcId)
		for _, cidr := range b.Vpc.Ipv4Cidrs {
			info.LocalIpv4Cidrs = append(info.LocalIpv4Cidrs, tea.StringValue(cidr))
		}
	}
	if b.AcceptingVpc != nil {
		info.RemoteVpcId = tea.StringValue(b.AcceptingVpc.VpcId)
		for _, cidr := range b.AcceptingVpc.Ipv4Cidrs {
			info.RemoteIpv4Cidrs = append(info.RemoteIpv4Cidrs, tea.StringValue(cidr))
		}
	}
	return info, nil
}

func (c *alicloudVpcPeeringClient) DeleteVpcPeerConnection(ctx context.Context, instanceId string) error {
	req := &vpcpeer.DeleteVpcPeerConnectionRequest{
		InstanceId: tea.String(instanceId),
	}
	_, err := c.peerClient.DeleteVpcPeerConnection(req)
	if err != nil {
		if IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("error deleting alicloud vpc peer connection %s: %w", instanceId, err)
	}
	return nil
}

func (c *alicloudVpcPeeringClient) ListVpcPeerConnections(ctx context.Context, vpcId, name string) ([]VpcPeerInfo, error) {
	var result []VpcPeerInfo
	var nextToken *string
	for {
		req := &vpcpeer.ListVpcPeerConnectionsRequest{
			RegionId:   tea.String(c.region),
			VpcId:      []*string{tea.String(vpcId)},
			Name:       tea.String(name),
			MaxResults: tea.Int32(100),
		}
		if nextToken != nil {
			req.NextToken = nextToken
		}
		resp, err := c.peerClient.ListVpcPeerConnections(req)
		if err != nil {
			return nil, fmt.Errorf("error listing alicloud vpc peer connections: %w", err)
		}
		if resp.Body == nil {
			break
		}
		for _, p := range resp.Body.VpcPeerConnects {
			info := VpcPeerInfo{
				InstanceId:        tea.StringValue(p.InstanceId),
				Status:            tea.StringValue(p.Status),
				RegionId:          tea.StringValue(p.RegionId),
				AcceptingRegionId: tea.StringValue(p.AcceptingRegionId),
			}
			if p.Vpc != nil {
				info.LocalVpcId = tea.StringValue(p.Vpc.VpcId)
				for _, cidr := range p.Vpc.Ipv4Cidrs {
					info.LocalIpv4Cidrs = append(info.LocalIpv4Cidrs, tea.StringValue(cidr))
				}
			}
			if p.AcceptingVpc != nil {
				info.RemoteVpcId = tea.StringValue(p.AcceptingVpc.VpcId)
				for _, cidr := range p.AcceptingVpc.Ipv4Cidrs {
					info.RemoteIpv4Cidrs = append(info.RemoteIpv4Cidrs, tea.StringValue(cidr))
				}
			}
			result = append(result, info)
		}
		if resp.Body.NextToken == nil || tea.StringValue(resp.Body.NextToken) == "" {
			break
		}
		nextToken = resp.Body.NextToken
	}
	return result, nil
}

func (c *alicloudVpcPeeringClient) DescribeRouteTables(ctx context.Context, vpcId string) ([]RouteTableInfo, error) {
	var result []RouteTableInfo
	pageNum := int32(1)
	const pageSize = int32(50)
	for {
		req := &vpc.DescribeRouteTableListRequest{
			RegionId:   tea.String(c.region),
			VpcId:      tea.String(vpcId),
			PageNumber: tea.Int32(pageNum),
			PageSize:   tea.Int32(pageSize),
		}
		resp, err := c.vpcClient.DescribeRouteTableList(req)
		if err != nil {
			return nil, fmt.Errorf("error describing alicloud route tables for vpc %s: %w", vpcId, err)
		}
		if resp.Body == nil || resp.Body.RouterTableList == nil {
			break
		}
		for _, t := range resp.Body.RouterTableList.RouterTableListType {
			info := RouteTableInfo{
				RouteTableId: tea.StringValue(t.RouteTableId),
				Tags:         map[string]string{},
			}
			if t.Tags != nil {
				for _, tag := range t.Tags.Tag {
					if tag.Key != nil && tag.Value != nil {
						info.Tags[tea.StringValue(tag.Key)] = tea.StringValue(tag.Value)
					}
				}
			}
			result = append(result, info)
		}
		totalCount := tea.Int32Value(resp.Body.TotalCount)
		if int32(len(result)) >= totalCount || int32(len(resp.Body.RouterTableList.RouterTableListType)) < pageSize {
			break
		}
		pageNum++
	}
	return result, nil
}

func (c *alicloudVpcPeeringClient) CreateRouteEntry(ctx context.Context, routeTableId, destCidr, instanceId string) error {
	req := &vpc.CreateRouteEntryRequest{
		RegionId:             tea.String(c.region),
		RouteTableId:         tea.String(routeTableId),
		DestinationCidrBlock: tea.String(destCidr),
		NextHopType:          tea.String("VpcPeer"),
		NextHopId:            tea.String(instanceId),
	}
	_, err := c.vpcClient.CreateRouteEntry(req)
	if err != nil {
		if IsAlreadyExists(err) {
			return nil
		}
		return fmt.Errorf("error creating alicloud route entry in %s for %s: %w", routeTableId, destCidr, err)
	}
	return nil
}

func (c *alicloudVpcPeeringClient) DeleteRouteEntry(ctx context.Context, routeTableId, destCidr, instanceId string) error {
	req := &vpc.DeleteRouteEntryRequest{
		RegionId:             tea.String(c.region),
		RouteTableId:         tea.String(routeTableId),
		DestinationCidrBlock: tea.String(destCidr),
		NextHopId:            tea.String(instanceId),
	}
	_, err := c.vpcClient.DeleteRouteEntry(req)
	if err != nil {
		if IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("error deleting alicloud route entry in %s for %s: %w", routeTableId, destCidr, err)
	}
	return nil
}

func IsNotFound(err error) bool {
	var sdkErr *tea.SDKError
	if errors.As(err, &sdkErr) && sdkErr.Code != nil {
		code := tea.StringValue(sdkErr.Code)
		return code == "ResourceNotFound.InstanceId" ||
			code == "InvalidInstanceId.NotFound" ||
			code == "InvalidRouteEntry.NotFound" ||
			code == "RouterEntry.NotFound"
	}
	return false
}

func IsAlreadyExists(err error) bool {
	var sdkErr *tea.SDKError
	if errors.As(err, &sdkErr) && sdkErr.Code != nil {
		code := tea.StringValue(sdkErr.Code)
		return code == "RouterEntryConflict.Duplicated" ||
			code == "EntryAlreadyExist"
	}
	return false
}

func IsRetryable(err error) bool {
	var sdkErr *tea.SDKError
	if errors.As(err, &sdkErr) && sdkErr.Code != nil {
		code := tea.StringValue(sdkErr.Code)
		return code == "IncorrectStatus" ||
			code == "TaskConflict" ||
			code == "OperationConflict" ||
			code == "ServiceUnavailable" ||
			code == "Throttling"
	}
	return false
}
