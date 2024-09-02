package cloudclient

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"time"
)

type ClientProvider[T any] func(ctx context.Context, saJsonKeyPath string) (T, error)

const GcpRetryWaitTime = time.Second * 3
const GcpOperationWaitTime = time.Second * 5
const GcpApiTimeout = time.Second * 3

type GcpConfig struct {
	GcpRetryWaitTime     time.Duration
	GcpOperationWaitTime time.Duration
	GcpApiTimeout        time.Duration
}

func GetGcpConfig(env abstractions.Environment) *GcpConfig {
	return &GcpConfig{
		GcpRetryWaitTime:     GetConfigDuration(env, "GCP_RETRY_WAIT_DURATION", GcpRetryWaitTime),
		GcpOperationWaitTime: GetConfigDuration(env, "GCP_OPERATION_WAIT_DURATION", GcpOperationWaitTime),
		GcpApiTimeout:        GetConfigDuration(env, "GCP_API_TIMEOUT_DURATION", GcpApiTimeout),
	}
}

func GetConfigDuration(env abstractions.Environment, key string, defaultValue time.Duration) time.Duration {
	duration, err := time.ParseDuration(env.Get(key))
	if err != nil {
		return defaultValue
	}
	return duration
}
