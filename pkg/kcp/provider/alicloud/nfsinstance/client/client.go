package client

import (
	"context"
	"errors"
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
			AccessKeyId:     new(accessKeyId),
			AccessKeySecret: new(accessKeySecret),
			RegionId:        new(region),
		}
		config.Endpoint = new(fmt.Sprintf("nas.%s.aliyuncs.com", region))

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

// isAccessGroupNotFound reports whether err is the AliCloud NAS "access group does not
// exist" error. DescribeAccessGroups / DescribeAccessRules return this (HTTP 404,
// Code=InvalidAccessGroup.NotFound) instead of an empty result when the group is absent,
// so callers that load-before-create must treat it as "not found", not a failure.
func isAccessGroupNotFound(err error) bool {
	return isSdkErrorWithCode(err, "InvalidAccessGroup.NotFound")
}

// isFileSystemNotFound reports whether err is the AliCloud NAS "file system does not exist"
// error. DescribeFileSystems / DescribeMountTargets return this (HTTP 404,
// Code=InvalidFileSystem.NotFound) instead of an empty result when the file system is
// absent, so load-before-create callers must treat it as "not found", not a failure.
func isFileSystemNotFound(err error) bool {
	return isSdkErrorWithCode(err, "InvalidFileSystem.NotFound")
}

// isSdkErrorWithCode reports whether err is an AliCloud SDK error with the given Code.
func isSdkErrorWithCode(err error, code string) bool {
	var sdkErr *tea.SDKError
	if errors.As(err, &sdkErr) {
		return tea.StringValue(sdkErr.Code) == code
	}
	return false
}

type alicloudClient struct {
	nasClient *nas.Client
	region    string
}

func (c *alicloudClient) DescribeFileSystem(ctx context.Context, fileSystemId string) (*FileSystemInfo, error) {
	req := &nas.DescribeFileSystemsRequest{
		FileSystemId:   new(fileSystemId),
		FileSystemType: new(fileSystemType),
	}

	resp, err := c.nasClient.DescribeFileSystems(req)
	if err != nil {
		if isFileSystemNotFound(err) {
			return nil, nil
		}
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
		FileSystemType: new(fileSystemType),
		ProtocolType:   new(protocolType),
		StorageType:    new(storageType),
		ChargeType:     new("PayAsYouGo"),
	}
	if zoneId != "" {
		req.ZoneId = new(zoneId)
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
		FileSystemId: new(fileSystemId),
	}

	_, err := c.nasClient.DeleteFileSystem(req)
	if err != nil {
		return fmt.Errorf("error deleting alicloud nas file system %s: %w", fileSystemId, err)
	}

	return nil
}

func (c *alicloudClient) DescribeMountTargets(ctx context.Context, fileSystemId string) ([]MountTargetInfo, error) {
	req := &nas.DescribeMountTargetsRequest{
		FileSystemId: new(fileSystemId),
	}

	resp, err := c.nasClient.DescribeMountTargets(req)
	if err != nil {
		if isFileSystemNotFound(err) {
			return nil, nil
		}
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
		FileSystemId:    new(fileSystemId),
		NetworkType:     new("Vpc"),
		VpcId:           new(vpcId),
		VSwitchId:       new(vSwitchId),
		AccessGroupName: new(accessGroupName),
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
		FileSystemId:      new(fileSystemId),
		MountTargetDomain: new(mountTargetDomain),
	}

	_, err := c.nasClient.DeleteMountTarget(req)
	if err != nil {
		return fmt.Errorf("error deleting alicloud nas mount target %s: %w", mountTargetDomain, err)
	}

	return nil
}

func (c *alicloudClient) DescribeAccessGroups(ctx context.Context, accessGroupName string) ([]AccessGroupInfo, error) {
	req := &nas.DescribeAccessGroupsRequest{
		AccessGroupName: new(accessGroupName),
		FileSystemType:  new(fileSystemType),
	}

	resp, err := c.nasClient.DescribeAccessGroups(req)
	if err != nil {
		if isAccessGroupNotFound(err) {
			return nil, nil
		}
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
		AccessGroupName: new(accessGroupName),
		AccessGroupType: new("Vpc"),
		FileSystemType:  new(fileSystemType),
		Description:     new(description),
	}

	_, err := c.nasClient.CreateAccessGroup(req)
	if err != nil {
		return fmt.Errorf("error creating alicloud nas access group %s: %w", accessGroupName, err)
	}

	return nil
}

func (c *alicloudClient) DeleteAccessGroup(ctx context.Context, accessGroupName string) error {
	req := &nas.DeleteAccessGroupRequest{
		AccessGroupName: new(accessGroupName),
		FileSystemType:  new(fileSystemType),
	}

	_, err := c.nasClient.DeleteAccessGroup(req)
	if err != nil {
		return fmt.Errorf("error deleting alicloud nas access group %s: %w", accessGroupName, err)
	}

	return nil
}

func (c *alicloudClient) DescribeAccessRules(ctx context.Context, accessGroupName string) ([]string, error) {
	req := &nas.DescribeAccessRulesRequest{
		AccessGroupName: new(accessGroupName),
		FileSystemType:  new(fileSystemType),
	}

	resp, err := c.nasClient.DescribeAccessRules(req)
	if err != nil {
		if isAccessGroupNotFound(err) {
			return nil, nil
		}
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
		AccessGroupName: new(accessGroupName),
		FileSystemType:  new(fileSystemType),
		SourceCidrIp:    new(sourceCidrIp),
		RWAccessType:    new("RDWR"),
		UserAccessType:  new("no_squash"),
		Priority:        tea.Int32(1),
	}

	_, err := c.nasClient.CreateAccessRule(req)
	if err != nil {
		return fmt.Errorf("error creating alicloud nas access rule for %s: %w", accessGroupName, err)
	}

	return nil
}
