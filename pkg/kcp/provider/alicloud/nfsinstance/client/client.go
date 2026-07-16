package client

import (
	"context"
	"fmt"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	nas "github.com/alibabacloud-go/nas-20170626/v3/client"
	"github.com/alibabacloud-go/tea/tea"
)

// FileSystemInfo is a provider-agnostic view of an AliCloud NAS file system.
type FileSystemInfo struct {
	FileSystemId   string
	FileSystemType string
	ProtocolType   string
	StorageType    string
	Status         string
	ZoneId         string
	// MeteredSize is the used size in bytes; Capacity is the provisioned size in GiB.
	MeteredSize int64
	Capacity    int64
}

// MountTargetInfo is a provider-agnostic view of an AliCloud NAS mount target.
type MountTargetInfo struct {
	MountTargetDomain string
	NetworkType       string
	VpcId             string
	VSwitchId         string
	AccessGroup       string
	Status            string
}

// AccessGroupInfo is a provider-agnostic view of an AliCloud NAS permission group.
type AccessGroupInfo struct {
	AccessGroupName string
	AccessGroupType string
	FileSystemType  string
}

type Client interface {
	// File system
	DescribeFileSystem(ctx context.Context, fileSystemId string) (*FileSystemInfo, error)
	CreateFileSystem(ctx context.Context, protocolType, storageType, zoneId string) (string, error)
	DeleteFileSystem(ctx context.Context, fileSystemId string) error

	// Mount targets
	DescribeMountTargets(ctx context.Context, fileSystemId string) ([]MountTargetInfo, error)
	CreateMountTarget(ctx context.Context, fileSystemId, vpcId, vSwitchId, accessGroupName string) (string, error)
	DeleteMountTarget(ctx context.Context, fileSystemId, mountTargetDomain string) error

	// Access groups (permission groups)
	DescribeAccessGroups(ctx context.Context, accessGroupName string) ([]AccessGroupInfo, error)
	CreateAccessGroup(ctx context.Context, accessGroupName, description string) error
	DeleteAccessGroup(ctx context.Context, accessGroupName string) error

	// Access rules
	DescribeAccessRules(ctx context.Context, accessGroupName string) ([]string, error)
	CreateAccessRule(ctx context.Context, accessGroupName, sourceCidrIp string) error
}

type ClientProvider func(ctx context.Context, region, accessKeyId, accessKeySecret string) (Client, error)

func NewClientProvider() ClientProvider {
	return func(ctx context.Context, region, accessKeyId, accessKeySecret string) (Client, error) {
		config := &openapi.Config{
			AccessKeyId:     tea.String(accessKeyId),
			AccessKeySecret: tea.String(accessKeySecret),
			RegionId:        tea.String(region),
		}
		config.Endpoint = tea.String(fmt.Sprintf("nas.%s.aliyuncs.com", region))

		nasClient, err := nas.NewClient(config)
		if err != nil {
			return nil, fmt.Errorf("error creating alicloud nas client: %w", err)
		}

		return &alicloudClient{
			nasClient: nasClient,
			region:    region,
		}, nil
	}
}

// fileSystemType is fixed to general-purpose NAS. Extreme/CPFS are out of scope.
const fileSystemType = "standard"

type alicloudClient struct {
	nasClient *nas.Client
	region    string
}

func (c *alicloudClient) DescribeFileSystem(ctx context.Context, fileSystemId string) (*FileSystemInfo, error) {
	req := &nas.DescribeFileSystemsRequest{
		FileSystemId:   tea.String(fileSystemId),
		FileSystemType: tea.String(fileSystemType),
	}

	resp, err := c.nasClient.DescribeFileSystems(req)
	if err != nil {
		return nil, fmt.Errorf("error describing alicloud nas file system %s: %w", fileSystemId, err)
	}

	if resp.Body == nil || resp.Body.FileSystems == nil || len(resp.Body.FileSystems.FileSystem) == 0 {
		return nil, nil
	}

	fs := resp.Body.FileSystems.FileSystem[0]
	return &FileSystemInfo{
		FileSystemId:   tea.StringValue(fs.FileSystemId),
		FileSystemType: tea.StringValue(fs.FileSystemType),
		ProtocolType:   tea.StringValue(fs.ProtocolType),
		StorageType:    tea.StringValue(fs.StorageType),
		Status:         tea.StringValue(fs.Status),
		ZoneId:         tea.StringValue(fs.ZoneId),
		MeteredSize:    tea.Int64Value(fs.MeteredSize),
		Capacity:       tea.Int64Value(fs.Capacity),
	}, nil
}

func (c *alicloudClient) CreateFileSystem(ctx context.Context, protocolType, storageType, zoneId string) (string, error) {
	req := &nas.CreateFileSystemRequest{
		FileSystemType: tea.String(fileSystemType),
		ProtocolType:   tea.String(protocolType),
		StorageType:    tea.String(storageType),
		ChargeType:     tea.String("PayAsYouGo"),
	}
	if zoneId != "" {
		req.ZoneId = tea.String(zoneId)
	}

	resp, err := c.nasClient.CreateFileSystem(req)
	if err != nil {
		return "", fmt.Errorf("error creating alicloud nas file system: %w", err)
	}
	if resp.Body == nil {
		return "", fmt.Errorf("error creating alicloud nas file system: empty response body")
	}

	return tea.StringValue(resp.Body.FileSystemId), nil
}

func (c *alicloudClient) DeleteFileSystem(ctx context.Context, fileSystemId string) error {
	req := &nas.DeleteFileSystemRequest{
		FileSystemId: tea.String(fileSystemId),
	}

	_, err := c.nasClient.DeleteFileSystem(req)
	if err != nil {
		return fmt.Errorf("error deleting alicloud nas file system %s: %w", fileSystemId, err)
	}

	return nil
}

func (c *alicloudClient) DescribeMountTargets(ctx context.Context, fileSystemId string) ([]MountTargetInfo, error) {
	req := &nas.DescribeMountTargetsRequest{
		FileSystemId: tea.String(fileSystemId),
	}

	resp, err := c.nasClient.DescribeMountTargets(req)
	if err != nil {
		return nil, fmt.Errorf("error describing alicloud nas mount targets for %s: %w", fileSystemId, err)
	}

	var result []MountTargetInfo
	if resp.Body != nil && resp.Body.MountTargets != nil {
		for _, mt := range resp.Body.MountTargets.MountTarget {
			result = append(result, MountTargetInfo{
				MountTargetDomain: tea.StringValue(mt.MountTargetDomain),
				NetworkType:       tea.StringValue(mt.NetworkType),
				VpcId:             tea.StringValue(mt.VpcId),
				VSwitchId:         tea.StringValue(mt.VswId),
				AccessGroup:       tea.StringValue(mt.AccessGroup),
				Status:            tea.StringValue(mt.Status),
			})
		}
	}

	return result, nil
}

func (c *alicloudClient) CreateMountTarget(ctx context.Context, fileSystemId, vpcId, vSwitchId, accessGroupName string) (string, error) {
	req := &nas.CreateMountTargetRequest{
		FileSystemId:    tea.String(fileSystemId),
		NetworkType:     tea.String("Vpc"),
		VpcId:           tea.String(vpcId),
		VSwitchId:       tea.String(vSwitchId),
		AccessGroupName: tea.String(accessGroupName),
	}

	resp, err := c.nasClient.CreateMountTarget(req)
	if err != nil {
		return "", fmt.Errorf("error creating alicloud nas mount target for %s: %w", fileSystemId, err)
	}
	if resp.Body == nil {
		return "", fmt.Errorf("error creating alicloud nas mount target: empty response body")
	}

	return tea.StringValue(resp.Body.MountTargetDomain), nil
}

func (c *alicloudClient) DeleteMountTarget(ctx context.Context, fileSystemId, mountTargetDomain string) error {
	req := &nas.DeleteMountTargetRequest{
		FileSystemId:      tea.String(fileSystemId),
		MountTargetDomain: tea.String(mountTargetDomain),
	}

	_, err := c.nasClient.DeleteMountTarget(req)
	if err != nil {
		return fmt.Errorf("error deleting alicloud nas mount target %s: %w", mountTargetDomain, err)
	}

	return nil
}

func (c *alicloudClient) DescribeAccessGroups(ctx context.Context, accessGroupName string) ([]AccessGroupInfo, error) {
	req := &nas.DescribeAccessGroupsRequest{
		AccessGroupName: tea.String(accessGroupName),
		FileSystemType:  tea.String(fileSystemType),
	}

	resp, err := c.nasClient.DescribeAccessGroups(req)
	if err != nil {
		return nil, fmt.Errorf("error describing alicloud nas access groups %s: %w", accessGroupName, err)
	}

	var result []AccessGroupInfo
	if resp.Body != nil && resp.Body.AccessGroups != nil {
		for _, ag := range resp.Body.AccessGroups.AccessGroup {
			result = append(result, AccessGroupInfo{
				AccessGroupName: tea.StringValue(ag.AccessGroupName),
				AccessGroupType: tea.StringValue(ag.AccessGroupType),
				FileSystemType:  tea.StringValue(ag.FileSystemType),
			})
		}
	}

	return result, nil
}

func (c *alicloudClient) CreateAccessGroup(ctx context.Context, accessGroupName, description string) error {
	req := &nas.CreateAccessGroupRequest{
		AccessGroupName: tea.String(accessGroupName),
		AccessGroupType: tea.String("Vpc"),
		FileSystemType:  tea.String(fileSystemType),
		Description:     tea.String(description),
	}

	_, err := c.nasClient.CreateAccessGroup(req)
	if err != nil {
		return fmt.Errorf("error creating alicloud nas access group %s: %w", accessGroupName, err)
	}

	return nil
}

func (c *alicloudClient) DeleteAccessGroup(ctx context.Context, accessGroupName string) error {
	req := &nas.DeleteAccessGroupRequest{
		AccessGroupName: tea.String(accessGroupName),
		FileSystemType:  tea.String(fileSystemType),
	}

	_, err := c.nasClient.DeleteAccessGroup(req)
	if err != nil {
		return fmt.Errorf("error deleting alicloud nas access group %s: %w", accessGroupName, err)
	}

	return nil
}

func (c *alicloudClient) DescribeAccessRules(ctx context.Context, accessGroupName string) ([]string, error) {
	req := &nas.DescribeAccessRulesRequest{
		AccessGroupName: tea.String(accessGroupName),
		FileSystemType:  tea.String(fileSystemType),
	}

	resp, err := c.nasClient.DescribeAccessRules(req)
	if err != nil {
		return nil, fmt.Errorf("error describing alicloud nas access rules for %s: %w", accessGroupName, err)
	}

	var result []string
	if resp.Body != nil && resp.Body.AccessRules != nil {
		for _, ar := range resp.Body.AccessRules.AccessRule {
			result = append(result, tea.StringValue(ar.SourceCidrIp))
		}
	}

	return result, nil
}

func (c *alicloudClient) CreateAccessRule(ctx context.Context, accessGroupName, sourceCidrIp string) error {
	req := &nas.CreateAccessRuleRequest{
		AccessGroupName: tea.String(accessGroupName),
		FileSystemType:  tea.String(fileSystemType),
		SourceCidrIp:    tea.String(sourceCidrIp),
		RWAccessType:    tea.String("RDWR"),
		UserAccessType:  tea.String("no_squash"),
		Priority:        tea.Int32(1),
	}

	_, err := c.nasClient.CreateAccessRule(req)
	if err != nil {
		return fmt.Errorf("error creating alicloud nas access rule for %s: %w", accessGroupName, err)
	}

	return nil
}
