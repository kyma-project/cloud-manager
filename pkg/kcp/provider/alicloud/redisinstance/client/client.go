// Package client wraps the AliCloud r-kvstore SDK for use by the KCP AliCloud
// Redis reconcilers. It exposes a narrow, provider-neutral Client interface
// covering the standard (non-sharded) HA instance lifecycle.
//
// The interface intentionally mirrors the operations described in issue #2012:
//
//   - CreateInstance          → r-kvstore CreateInstance
//   - DescribeInstance        → r-kvstore DescribeInstanceAttribute
//   - ModifyInstanceSpec      → r-kvstore ModifyInstanceSpec
//   - DeleteInstance          → r-kvstore DeleteInstance
//   - ResetAccountPassword    → r-kvstore ResetAccountPassword
//
// Cluster-specific operations (AddShardingNode/DeleteShardingNode) live in the
// sibling rediscluster/client package which embeds this interface.
package client

import (
	"context"
	"errors"
	"fmt"
	"strings"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	rkvstore "github.com/alibabacloud-go/r-kvstore-20150101/v7/client"
	"github.com/alibabacloud-go/tea/tea"
)

// Instance status values used by the AliCloud r-kvstore API. Only a subset is
// consumed by the reconcilers; the constants are documented in issue #2012.
const (
	InstanceStatusCreating     = "Creating"
	InstanceStatusNormal       = "Normal"
	InstanceStatusChanging     = "Changing"
	InstanceStatusSSLModifying = "SSLModifying"
	InstanceStatusInactive     = "Inactive"
	InstanceStatusReleased     = "Released"
)

// NetworkType is always VPC for cloud-manager (public endpoints are not
// supported in v1 per design decision).
const NetworkType = "VPC"

// ChargeType is always PostPaid for cloud-manager (PrePaid deletion is not
// supported via API, deferred).
const ChargeType = "PostPaid"

// DefaultPort is the r-kvstore default listening port.
const DefaultPort int64 = 6379

// InstanceInfo is a provider-neutral snapshot of the state returned by
// DescribeInstanceAttribute.
type InstanceInfo struct {
	InstanceId       string
	InstanceStatus   string
	InstanceClass    string
	ArchitectureType string
	NodeType         string
	NetworkType      string
	VpcId            string
	VSwitchId        string
	EngineVersion    string
	Capacity         int64
	Port             int64
	ConnectionDomain string
	ChargeType       string
	ShardCount       int32
	ReadOnlyCount    int32
	Config           string
}

// CreateInstanceOptions are the arguments accepted by CreateInstance.
//
// ZoneId is intentionally omitted from this struct: the AliCloud API derives
// it automatically from VSwitchId (design decision 7 in issue #2012).
type CreateInstanceOptions struct {
	InstanceName  string
	InstanceClass string
	EngineVersion string
	VpcId         string
	VSwitchId     string
	Password      string
	Port          int64
	// ShardCount is only meaningful for cluster classes (redis.shard.*.ce).
	// Leave zero for standard instances.
	ShardCount int32
	// ReadOnlyCount encodes the S/P tier for standard instances (0/1).
	ReadOnlyCount int32
	// Token is an idempotency token; if set, repeated calls with the same
	// token return the same instance without creating a duplicate.
	Token string
}

// ModifyInstanceSpecOptions covers the ModifyInstanceSpec request surface
// used by the cloud-manager reconcilers. Fields set to zero value are omitted
// from the request. ReadOnlyCount uses a pointer so that zero (transition P→S)
// can be distinguished from "not provided".
type ModifyInstanceSpecOptions struct {
	InstanceClass string
	ShardCount    int32
	// ReadOnlyCount is nil when the caller does not want to change the replica
	// count. Use tea.Int32(0) to explicitly set zero (P→S transition).
	ReadOnlyCount *int32
}

// Client is the AliCloud r-kvstore standard-instance client contract.
type Client interface {
	CreateInstance(ctx context.Context, opts CreateInstanceOptions) (instanceId string, err error)
	DescribeInstance(ctx context.Context, instanceId string) (*InstanceInfo, error)
	// DescribeInstanceByName searches for an instance by exact name within the
	// current region. Returns (nil, nil) when no matching instance is found.
	// Used to recover from crash-after-create-before-status-write scenarios.
	DescribeInstanceByName(ctx context.Context, name string) (*InstanceInfo, error)
	ModifyInstanceSpec(ctx context.Context, instanceId string, opts ModifyInstanceSpecOptions) error
	// ModifyInstanceConfig applies runtime configuration parameters encoded as
	// a JSON string (e.g. `{"maxmemory-policy":"noeviction"}`).
	ModifyInstanceConfig(ctx context.Context, instanceId, config string) error
	// DescribeSecurityIps returns the comma-separated IP list for the "default"
	// security group on an instance. Returns an empty string when the group is
	// absent (e.g. newly created instance that only has 127.0.0.1).
	DescribeSecurityIps(ctx context.Context, instanceId string) (string, error)
	// ModifySecurityIps replaces the default security IP group with securityIps
	// (comma-separated CIDR list). AliCloud Redis instances are created with
	// security IPs set to 127.0.0.1, blocking all external access. This must be
	// called after CreateInstance to allow VPC traffic to reach the instance.
	ModifySecurityIps(ctx context.Context, instanceId, securityIps string) error
	DeleteInstance(ctx context.Context, instanceId string) error
	ResetAccountPassword(ctx context.Context, instanceId, accountName, password string) error
	// DescribeInstanceSSL returns whether SSL/TLS encryption is enabled on the
	// instance. The AliCloud API uses string values "Enable"/"Disable" internally;
	// this method normalises to bool.
	DescribeInstanceSSL(ctx context.Context, instanceId string) (sslEnabled bool, err error)
	// ModifyInstanceSSL enables or disables SSL/TLS encryption on the instance.
	// The underlying AliCloud ModifyInstanceSSL call starts an async operation;
	// the caller should wait for the instance to return to Normal status before
	// reading connection details.
	ModifyInstanceSSL(ctx context.Context, instanceId string, enable bool) error
}

// IsPermanentError returns true for AliCloud SDK errors that will never
// succeed on retry - specifically 4xx responses other than 429 (rate limit).
// Callers should use StopAndForget for permanent errors instead of requeueing.
func IsPermanentError(err error) bool {
	var sdkErr *tea.SDKError
	if errors.As(err, &sdkErr) {
		if sdkErr.StatusCode == nil {
			return false
		}
		code := tea.IntValue(sdkErr.StatusCode)
		return code >= 400 && code < 500 && code != 429
	}
	return false
}

// IsVSwitchZoneErr returns true when AliCloud rejects a CreateInstance call
// because the selected vSwitch's zone does not support the requested instance
// class. The caller should retry with a different vSwitch from another zone.
func IsVSwitchZoneErr(err error) bool {
	var sdkErr *tea.SDKError
	if errors.As(err, &sdkErr) && sdkErr.Code != nil {
		return tea.StringValue(sdkErr.Code) == "InvalidvSwitchId"
	}
	return false
}

// IsPasswordErr returns true when AliCloud rejects the CreateInstance password
// as malformed (InvalidPassword.Malformed). Callers should clear the stored
// authString so the next reconcile generates a fresh, compliant password.
func IsPasswordErr(err error) bool {
	var sdkErr *tea.SDKError
	if errors.As(err, &sdkErr) && sdkErr.Code != nil {
		return tea.StringValue(sdkErr.Code) == "InvalidPassword.Malformed"
	}
	return false
}

// IsNotFoundErr returns true when AliCloud reports the instance ID does not
// exist (404 InvalidInstanceId.NotFound). Callers should treat this as
// "instance already gone" rather than a transient error.
func IsNotFoundErr(err error) bool {
	var sdkErr *tea.SDKError
	if errors.As(err, &sdkErr) && sdkErr.Code != nil {
		return tea.StringValue(sdkErr.Code) == "InvalidInstanceId.NotFound"
	}
	return false
}

// IsProxyClusterClass returns true for proxy-based sharded cluster classes
// (redis.logic.sharding.* and redis.amber.logic.sharding.*). These classes
// encode replica count in the class name; ReadOnlyCount is always 0 and
// cannot be tuned independently via ModifyInstanceSpec.
func IsProxyClusterClass(instanceClass string) bool {
	return strings.HasPrefix(instanceClass, "redis.logic.sharding.") ||
		strings.HasPrefix(instanceClass, "redis.amber.logic.sharding.")
}

// IsReadOnlyCountUnsupported returns true for instance classes where the
// AliCloud API silently ignores ReadOnlyCount - the field is absent from
// DescribeInstanceAttribute and ModifyInstanceSpec with ReadOnlyCount has no
// effect. Callers must skip ReadOnlyCount drift checks for these classes to
// avoid an infinite modify loop.
//
// Affected families (confirmed via live API testing):
//   - tair.rdb.*: DRAM-based HA, ReadOnlyCount absent
//
// Note: redis.master.*.cloud (cloud-disk) and redis.master.*.default (local-disk)
// both support ReadOnlyCount correctly — they are NOT in this list.
func IsReadOnlyCountUnsupported(instanceClass string) bool {
	return strings.HasPrefix(instanceClass, "tair.rdb.")
}

// ClientProvider is the standard cloud-manager credential/region-scoped
// constructor signature used across all AliCloud client packages.
type ClientProvider func(ctx context.Context, region, accessKeyId, accessKeySecret string) (Client, error)

// NewClientProvider returns a ClientProvider that constructs a real SDK-backed
// AliCloud r-kvstore client.
func NewClientProvider() ClientProvider {
	return func(ctx context.Context, region, accessKeyId, accessKeySecret string) (Client, error) {
		config := &openapi.Config{
			AccessKeyId:     new(accessKeyId),
			AccessKeySecret: new(accessKeySecret),
			RegionId:        new(region),
		}
		config.Endpoint = new(fmt.Sprintf("r-kvstore.%s.aliyuncs.com", region))

		rc, err := rkvstore.NewClient(config)
		if err != nil {
			return nil, fmt.Errorf("error creating alicloud r-kvstore client: %w", err)
		}

		return &alicloudRedisClient{
			c:      rc,
			region: region,
		}, nil
	}
}

// NewClientFromSDK wraps an already-constructed *rkvstore.Client. Used by the
// cluster client package to share a single SDK connection for both standard
// instance and cluster-only operations.
func NewClientFromSDK(rc *rkvstore.Client, region string) Client {
	return &alicloudRedisClient{c: rc, region: region}
}

type alicloudRedisClient struct {
	c      *rkvstore.Client
	region string
}

// CreateInstance provisions a new r-kvstore instance. The instance
// architecture is determined by opts.InstanceClass:
//   - redis.master.*.cloud              → standard HA (cloud-disk, engines 5.0/6.0/7.0)
//   - redis.master.*.default            → standard HA (local-disk, engines 5.0/6.0 only)
//   - tair.rdb.*                        → DRAM-based HA
//   - redis.logic.sharding.*            → proxy-based cluster (ShardCount encoded in class name)
//   - redis.shard.*.ce                  → non-proxy cloud-native cluster (separate ShardCount)
func (c *alicloudRedisClient) CreateInstance(ctx context.Context, opts CreateInstanceOptions) (string, error) {
	req := &rkvstore.CreateInstanceRequest{
		RegionId:      new(c.region),
		InstanceName:  new(opts.InstanceName),
		InstanceClass: new(opts.InstanceClass),
		EngineVersion: new(opts.EngineVersion),
		VpcId:         new(opts.VpcId),
		VSwitchId:     new(opts.VSwitchId),
		NetworkType:   new(NetworkType),
		ChargeType:    new(ChargeType),
	}
	if opts.Password != "" {
		req.Password = new(opts.Password)
	}
	if opts.Port > 0 {
		portStr := fmt.Sprintf("%d", opts.Port)
		req.Port = new(portStr)
	}
	if opts.ShardCount > 0 {
		req.ShardCount = new(opts.ShardCount)
	}
	// AliCloud defaults ReadOnlyCount to 0 for redis.master.* classes, so
	// omitting the field for S-tier (ReadOnlyCount==0) is safe and correct.
	if opts.ReadOnlyCount > 0 {
		req.ReadOnlyCount = new(opts.ReadOnlyCount)
	}
	if opts.Token != "" {
		req.Token = new(opts.Token)
	}

	resp, err := c.c.CreateInstance(req)
	if err != nil {
		return "", fmt.Errorf("error creating alicloud r-kvstore instance: %w", err)
	}
	if resp == nil || resp.Body == nil {
		return "", fmt.Errorf("empty response from alicloud r-kvstore CreateInstance")
	}
	return tea.StringValue(resp.Body.InstanceId), nil
}

// DescribeInstance loads the current state of an instance. Returns (nil, nil)
// when the instance is not found (caller should treat as "does not exist").
func (c *alicloudRedisClient) DescribeInstance(ctx context.Context, instanceId string) (*InstanceInfo, error) {
	req := &rkvstore.DescribeInstanceAttributeRequest{
		InstanceId: new(instanceId),
	}
	resp, err := c.c.DescribeInstanceAttribute(req)
	if err != nil {
		return nil, fmt.Errorf("error describing alicloud r-kvstore instance %s: %w", instanceId, err)
	}
	if resp == nil || resp.Body == nil || resp.Body.Instances == nil {
		return nil, fmt.Errorf("empty response from alicloud r-kvstore DescribeInstanceAttribute for %s", instanceId)
	}
	if len(resp.Body.Instances.DBInstanceAttribute) == 0 {
		return nil, nil
	}
	a := resp.Body.Instances.DBInstanceAttribute[0]
	return &InstanceInfo{
		InstanceId:       tea.StringValue(a.InstanceId),
		InstanceStatus:   tea.StringValue(a.InstanceStatus),
		InstanceClass:    tea.StringValue(a.InstanceClass),
		ArchitectureType: tea.StringValue(a.ArchitectureType),
		NodeType:         tea.StringValue(a.NodeType),
		NetworkType:      tea.StringValue(a.NetworkType),
		VpcId:            tea.StringValue(a.VpcId),
		VSwitchId:        tea.StringValue(a.VSwitchId),
		EngineVersion:    tea.StringValue(a.EngineVersion),
		Capacity:         tea.Int64Value(a.Capacity),
		Port:             tea.Int64Value(a.Port),
		ConnectionDomain: tea.StringValue(a.ConnectionDomain),
		ChargeType:       tea.StringValue(a.ChargeType),
		ShardCount:       tea.Int32Value(a.ShardCount),
		ReadOnlyCount:    tea.Int32Value(a.ReadOnlyCount),
		Config:           tea.StringValue(a.Config),
	}, nil
}

// DescribeInstanceByName searches for an instance by exact name. Returns
// (nil, nil) when no matching instance is found.
func (c *alicloudRedisClient) DescribeInstanceByName(ctx context.Context, name string) (*InstanceInfo, error) {
	const pageSize = int32(50)
	pageNum := int32(1)
	for {
		req := &rkvstore.DescribeInstancesRequest{
			RegionId:   new(c.region),
			SearchKey:  new(name),
			PageSize:   tea.Int32(pageSize),
			PageNumber: new(pageNum),
		}
		resp, err := c.c.DescribeInstances(req)
		if err != nil {
			return nil, fmt.Errorf("error describing alicloud r-kvstore instances by name %s: %w", name, err)
		}
		if resp == nil || resp.Body == nil || resp.Body.Instances == nil {
			return nil, nil
		}
		for _, inst := range resp.Body.Instances.KVStoreInstance {
			if tea.StringValue(inst.InstanceName) == name {
				return c.DescribeInstance(ctx, tea.StringValue(inst.InstanceId))
			}
		}
		if resp.Body.TotalCount == nil || pageNum*pageSize >= tea.Int32Value(resp.Body.TotalCount) {
			break
		}
		pageNum++
	}
	return nil, nil
}

// ModifyInstanceSpec scales an existing instance. Per issue #2012 design
// decision 8, callers must not combine InstanceClass changes with ShardCount
// changes in the same request; the cluster reconciler splits these into two
// separate calls.
func (c *alicloudRedisClient) ModifyInstanceSpec(ctx context.Context, instanceId string, opts ModifyInstanceSpecOptions) error {
	req := &rkvstore.ModifyInstanceSpecRequest{
		InstanceId: new(instanceId),
	}
	if opts.InstanceClass != "" {
		req.InstanceClass = new(opts.InstanceClass)
	}
	if opts.ShardCount > 0 {
		req.ShardCount = new(opts.ShardCount)
	}
	if opts.ReadOnlyCount != nil {
		req.ReadOnlyCount = opts.ReadOnlyCount
	}

	if _, err := c.c.ModifyInstanceSpec(req); err != nil {
		return fmt.Errorf("error modifying alicloud r-kvstore instance %s: %w", instanceId, err)
	}
	return nil
}

// DeleteInstance releases a PostPaid r-kvstore instance. PrePaid instances
// cannot be deleted via API (design decision 5) and are not supported in v1.
func (c *alicloudRedisClient) DeleteInstance(ctx context.Context, instanceId string) error {
	req := &rkvstore.DeleteInstanceRequest{
		InstanceId: new(instanceId),
	}
	if _, err := c.c.DeleteInstance(req); err != nil {
		return fmt.Errorf("error deleting alicloud r-kvstore instance %s: %w", instanceId, err)
	}
	return nil
}

// ResetAccountPassword resets the password of a named account on an existing
// instance. Used only in edge-cases where the pre-creation password is lost
// (e.g. AuthSecret regeneration). Cloud-manager normally sets the initial
// password at CreateInstance time (design decision 6).
func (c *alicloudRedisClient) ResetAccountPassword(ctx context.Context, instanceId, accountName, password string) error {
	req := &rkvstore.ResetAccountPasswordRequest{
		InstanceId:      new(instanceId),
		AccountName:     new(accountName),
		AccountPassword: new(password),
	}
	if _, err := c.c.ResetAccountPassword(req); err != nil {
		return fmt.Errorf("error resetting password for alicloud r-kvstore instance %s account %s: %w", instanceId, accountName, err)
	}
	return nil
}

// ModifyInstanceConfig applies runtime configuration parameters to an existing
// instance. config must be a JSON object string as required by the AliCloud API.
func (c *alicloudRedisClient) ModifyInstanceConfig(ctx context.Context, instanceId, config string) error {
	req := &rkvstore.ModifyInstanceConfigRequest{
		InstanceId: new(instanceId),
		Config:     new(config),
	}
	if _, err := c.c.ModifyInstanceConfig(req); err != nil {
		return fmt.Errorf("error modifying alicloud r-kvstore instance %s config: %w", instanceId, err)
	}
	return nil
}

func (c *alicloudRedisClient) ModifySecurityIps(ctx context.Context, instanceId, securityIps string) error {
	req := &rkvstore.ModifySecurityIpsRequest{
		InstanceId:          new(instanceId),
		SecurityIps:         new(securityIps),
		SecurityIpGroupName: new("default"),
		ModifyMode:          new("Cover"),
	}
	if _, err := c.c.ModifySecurityIps(req); err != nil {
		return fmt.Errorf("error modifying alicloud r-kvstore instance %s security IPs: %w", instanceId, err)
	}
	return nil
}

func (c *alicloudRedisClient) DescribeSecurityIps(ctx context.Context, instanceId string) (string, error) {
	req := &rkvstore.DescribeSecurityIpsRequest{
		InstanceId: new(instanceId),
	}
	resp, err := c.c.DescribeSecurityIps(req)
	if err != nil {
		return "", fmt.Errorf("error describing alicloud r-kvstore instance %s security IPs: %w", instanceId, err)
	}
	if resp == nil || resp.Body == nil || resp.Body.SecurityIpGroups == nil {
		return "", nil
	}
	for _, g := range resp.Body.SecurityIpGroups.SecurityIpGroup {
		if tea.StringValue(g.SecurityIpGroupName) == "default" {
			return tea.StringValue(g.SecurityIpList), nil
		}
	}
	return "", nil
}

func (c *alicloudRedisClient) DescribeInstanceSSL(ctx context.Context, instanceId string) (bool, error) {
	req := &rkvstore.DescribeInstanceSSLRequest{
		InstanceId: new(instanceId),
	}
	resp, err := c.c.DescribeInstanceSSL(req)
	if err != nil {
		return false, fmt.Errorf("error describing alicloud r-kvstore instance %s SSL: %w", instanceId, err)
	}
	if resp == nil || resp.Body == nil {
		return false, nil
	}
	return tea.StringValue(resp.Body.SSLEnabled) == "Enable", nil
}

func (c *alicloudRedisClient) ModifyInstanceSSL(ctx context.Context, instanceId string, enable bool) error {
	sslEnabled := "Disable"
	if enable {
		sslEnabled = "Enable"
	}
	req := &rkvstore.ModifyInstanceSSLRequest{
		InstanceId: new(instanceId),
		SSLEnabled: new(sslEnabled),
	}
	if _, err := c.c.ModifyInstanceSSL(req); err != nil {
		return fmt.Errorf("error modifying alicloud r-kvstore instance %s SSL: %w", instanceId, err)
	}
	return nil
}
