package feature

import (
	"context"
)

const (
	alicloudRedisInstanceFlagName = "alicloudRedisInstance"
	alicloudRedisClusterFlagName  = "alicloudRedisCluster"
)

var AlicloudRedisInstance = &alicloudRedisInfo{flagName: alicloudRedisInstanceFlagName}
var AlicloudRedisCluster = &alicloudRedisInfo{flagName: alicloudRedisClusterFlagName}

type alicloudRedisInfo struct {
	flagName string
}

func (f *alicloudRedisInfo) Value(ctx context.Context) bool {
	return provider.BoolVariation(ctx, f.flagName, false)
}
