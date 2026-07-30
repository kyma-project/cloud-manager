package mock

import (
	"context"

	rediscluster "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/rediscluster/client"
	redisinstance "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/redisinstance/client"
)

// redisInstanceClientView adapts redisStore to redisinstance.Client.
type redisInstanceClientView struct{ *redisStore }

var _ redisinstance.Client = (*redisInstanceClientView)(nil)

func (c *redisInstanceClientView) CreateInstance(ctx context.Context, opts redisinstance.CreateInstanceOptions) (string, error) {
	return c.createInstance(ctx, opts)
}

func (c *redisInstanceClientView) DescribeInstance(ctx context.Context, instanceId string) (*redisinstance.InstanceInfo, error) {
	return c.describeInstance(ctx, instanceId)
}

func (c *redisInstanceClientView) DescribeInstanceByName(ctx context.Context, name string) (*redisinstance.InstanceInfo, error) {
	return c.describeInstanceByName(ctx, name)
}

func (c *redisInstanceClientView) ModifyInstanceSpec(ctx context.Context, instanceId string, opts redisinstance.ModifyInstanceSpecOptions) error {
	return c.modifyInstanceSpec(ctx, instanceId, opts)
}

func (c *redisInstanceClientView) ModifyInstanceConfig(ctx context.Context, instanceId, config string) error {
	return c.modifyInstanceConfig(ctx, instanceId, config)
}

func (c *redisInstanceClientView) ModifySecurityIps(ctx context.Context, instanceId, securityIps string) error {
	return c.modifySecurityIps(ctx, instanceId, securityIps)
}

func (c *redisInstanceClientView) DescribeSecurityIps(ctx context.Context, instanceId string) (string, error) {
	return c.describeSecurityIps(ctx, instanceId)
}

func (c *redisInstanceClientView) DeleteInstance(ctx context.Context, instanceId string) error {
	return c.deleteInstance(ctx, instanceId)
}

func (c *redisInstanceClientView) ResetAccountPassword(ctx context.Context, instanceId, accountName, password string) error {
	return c.resetAccountPassword(ctx, instanceId, accountName, password)
}

func (c *redisInstanceClientView) DescribeInstanceSSL(ctx context.Context, instanceId string) (bool, error) {
	return c.describeInstanceSSL(ctx, instanceId)
}

func (c *redisInstanceClientView) ModifyInstanceSSL(ctx context.Context, instanceId string, enable bool) error {
	return c.modifyInstanceSSL(ctx, instanceId, enable)
}

// redisClusterClientView adapts redisStore to rediscluster.Client.
type redisClusterClientView struct{ *redisStore }

var _ rediscluster.Client = (*redisClusterClientView)(nil)

func (c *redisClusterClientView) CreateInstance(ctx context.Context, opts redisinstance.CreateInstanceOptions) (string, error) {
	return c.createInstance(ctx, opts)
}

func (c *redisClusterClientView) DescribeInstance(ctx context.Context, instanceId string) (*redisinstance.InstanceInfo, error) {
	return c.describeInstance(ctx, instanceId)
}

func (c *redisClusterClientView) DescribeInstanceByName(ctx context.Context, name string) (*redisinstance.InstanceInfo, error) {
	return c.describeInstanceByName(ctx, name)
}

func (c *redisClusterClientView) ModifyInstanceSpec(ctx context.Context, instanceId string, opts redisinstance.ModifyInstanceSpecOptions) error {
	return c.modifyInstanceSpec(ctx, instanceId, opts)
}

func (c *redisClusterClientView) ModifyInstanceConfig(ctx context.Context, instanceId, config string) error {
	return c.modifyInstanceConfig(ctx, instanceId, config)
}

func (c *redisClusterClientView) ModifySecurityIps(ctx context.Context, instanceId, securityIps string) error {
	return c.modifySecurityIps(ctx, instanceId, securityIps)
}

func (c *redisClusterClientView) DescribeSecurityIps(ctx context.Context, instanceId string) (string, error) {
	return c.describeSecurityIps(ctx, instanceId)
}

func (c *redisClusterClientView) DeleteInstance(ctx context.Context, instanceId string) error {
	return c.deleteInstance(ctx, instanceId)
}

func (c *redisClusterClientView) ResetAccountPassword(ctx context.Context, instanceId, accountName, password string) error {
	return c.resetAccountPassword(ctx, instanceId, accountName, password)
}

func (c *redisClusterClientView) DescribeInstanceSSL(ctx context.Context, instanceId string) (bool, error) {
	return c.describeInstanceSSL(ctx, instanceId)
}

func (c *redisClusterClientView) ModifyInstanceSSL(ctx context.Context, instanceId string, enable bool) error {
	return c.modifyInstanceSSL(ctx, instanceId, enable)
}

func (c *redisClusterClientView) AddShardingNode(ctx context.Context, instanceId string, targetShardCount int32) error {
	return c.addShardingNode(ctx, instanceId, targetShardCount)
}

func (c *redisClusterClientView) DeleteShardingNode(ctx context.Context, instanceId string, targetShardCount int32) error {
	return c.deleteShardingNode(ctx, instanceId, targetShardCount)
}
