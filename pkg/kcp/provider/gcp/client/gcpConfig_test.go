package client

import (
	"testing"
	"time"

	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/config"
	"github.com/stretchr/testify/assert"
)

func Test_GcpConfig(t *testing.T) {

	t.Run("with env set", func(t *testing.T) {
		GcpConfig = &GcpConfigStruct{}

		env := abstractions.NewMockedEnvironment(map[string]string{
			"GCP_RETRY_WAIT_DURATION":     "123s",
			"GCP_OPERATION_WAIT_DURATION": "124s",
			"GCP_API_TIMEOUT_DURATION":    "125s",
			"GCP_CAPACITY_CHECK_INTERVAL": "126s",
		})

		cfg := config.NewConfig(env)
		InitConfig(cfg)
		cfg.Read()

		assert.Equal(t, "123s", GcpConfig.RetryWaitTime)
		assert.Equal(t, 123*time.Second, GcpConfig.GcpRetryWaitTime)

		assert.Equal(t, "124s", GcpConfig.OperationWaitTime)
		assert.Equal(t, 124*time.Second, GcpConfig.GcpOperationWaitTime)

		assert.Equal(t, "125s", GcpConfig.ApiTimeout)
		assert.Equal(t, 125*time.Second, GcpConfig.GcpApiTimeout)

		assert.Equal(t, "126s", GcpConfig.CapacityCheckInterval)
		assert.Equal(t, 126*time.Second, GcpConfig.GcpCapacityCheckInterval)
	})

	t.Run("empty", func(t *testing.T) {
		// must instantiate a new one to isolate from previous tests
		GcpConfig = &GcpConfigStruct{}

		env := abstractions.NewMockedEnvironment(map[string]string{})

		cfg := config.NewConfig(env)
		InitConfig(cfg)
		cfg.Read()

		assert.Equal(t, "5s", GcpConfig.RetryWaitTime)
		assert.Equal(t, 5*time.Second, GcpConfig.GcpRetryWaitTime)

		assert.Equal(t, "5s", GcpConfig.OperationWaitTime)
		assert.Equal(t, 5*time.Second, GcpConfig.GcpOperationWaitTime)

		assert.Equal(t, "8s", GcpConfig.ApiTimeout)
		assert.Equal(t, 8*time.Second, GcpConfig.GcpApiTimeout)

		assert.Equal(t, "1h", GcpConfig.CapacityCheckInterval)
		assert.Equal(t, time.Hour, GcpConfig.GcpCapacityCheckInterval)
	})
}
