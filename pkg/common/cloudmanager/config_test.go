package cloudmanager

import (
	"testing"
	"time"

	"github.com/kyma-project/cloud-manager/pkg/config"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/rest"
)

func TestConfigLoading(t *testing.T) {

	testCases := []struct {
		name   string
		env    map[string]string
		data   map[string]any
		assert func(t *testing.T, cfg *ConfigType)
	}{
		{
			name: "all default values",
			assert: func(t *testing.T, cfg *ConfigType) {
				assert.Equal(t, 0, cfg.Manager.Controller.MaxConcurrentReconciles, "maxConcurrentReconciles should be 0 by default")
				assert.Nil(t, cfg.Manager.Controller.GroupKindConcurrency, "groupKindConcurrency should be nil by default")
				assert.Equal(t, 0*time.Second, cfg.Manager.Controller.ReconciliationTimeout, "reconciliationTimeout should be 0 by default")
				assert.Equal(t, "", cfg.Manager.Pprof.Address, "pprof address should be empty by default")
				assert.InDelta(t, 50.0, cfg.RestConfig.QPS, 0.0001, "rest config QPS should be default 50.0")
				assert.Equal(t, 100, cfg.RestConfig.Burst, "rest config Burst should be default 100")
			},
		},
		{
			name: "pprof enabled from env",
			env: map[string]string{
				"ENABLE_PPROF": "true",
			},
			assert: func(t *testing.T, cfg *ConfigType) {
				assert.Equal(t, true, cfg.Manager.Pprof.Enabled, "pprof should be enabled when set from env ENABLE_PPROF")
				assert.Equal(t, ":8082", cfg.Manager.Pprof.Address, "pprof address should be default when only enabled from env and custom address not set")
			},
		},
		{
			name: "pprof enabled from env to custom address",
			env: map[string]string{
				"ENABLE_PPROF":       "true",
				"PPROF_BIND_ADDRESS": ":3456",
			},
			assert: func(t *testing.T, cfg *ConfigType) {
				assert.Equal(t, true, cfg.Manager.Pprof.Enabled, "pprof should be enabled when set from env ENABLE_PPROF")
				assert.Equal(t, ":3456", cfg.Manager.Pprof.Address, "pprof address should be custom when set from env PPROF_BIND_ADDRESS")
			},
		},
		{
			name: "manager controller maxConcurrentReconciles set from file",
			data: map[string]any{
				"manager": map[string]any{
					"controller": map[string]any{
						"maxConcurrentReconciles": 5,
					},
				},
			},
			assert: func(t *testing.T, cfg *ConfigType) {
				assert.Equal(t, 5, cfg.Manager.Controller.MaxConcurrentReconciles, "maxConcurrentReconciles should be set to file value 5")
			},
		},
		{
			name: "manager controller groupKindConcurrency set from file",
			data: map[string]any{
				"manager": map[string]any{
					"controller": map[string]any{
						"groupKindConcurrency": map[string]any{
							"apps/Deployment": 3,
							"core/Pod":        2,
						},
					},
				},
			},
			assert: func(t *testing.T, cfg *ConfigType) {
				assert.Equal(t, 3, cfg.Manager.Controller.GroupKindConcurrency["apps/Deployment"], "groupKindConcurrency for apps/Deployment should be set to 3")
				assert.Equal(t, 2, cfg.Manager.Controller.GroupKindConcurrency["core/Pod"], "groupKindConcurrency for core/Pod should be set to 2")
			},
		},
		{
			name: "manager controller reconciliationTimeout set from file in seconds",
			data: map[string]any{
				"manager": map[string]any{
					"controller": map[string]any{
						"reconciliationTimeout": "43s",
					},
				},
			},
			assert: func(t *testing.T, cfg *ConfigType) {
				assert.Equal(t, 43*time.Second, cfg.Manager.Controller.ReconciliationTimeout, "reconciliationTimeout should be set to 43s")
			},
		},
		{
			name: "manager controller reconciliationTimeout set from file in minutes",
			data: map[string]any{
				"manager": map[string]any{
					"controller": map[string]any{
						"reconciliationTimeout": "31m",
					},
				},
			},
			assert: func(t *testing.T, cfg *ConfigType) {
				assert.Equal(t, 31*time.Minute, cfg.Manager.Controller.ReconciliationTimeout, "reconciliationTimeout should be set to 31m")
			},
		},
		{
			name: "rest config QPS and Burst set from file",
			data: map[string]any{
				"restConfig": map[string]any{
					"qps":   100.0,
					"burst": 200,
				},
			},
			assert: func(t *testing.T, cfg *ConfigType) {
				assert.InDelta(t, 100.0, cfg.RestConfig.QPS, 0.0001, "rest config QPS should be set to 100.0")
				assert.Equal(t, 200, cfg.RestConfig.Burst, "rest config Burst should be set to 200")
			},
		},
		{
			name: "rest config QPS set from file as int",
			data: map[string]any{
				"restConfig": map[string]any{
					"qps": 100,
				},
			},
			assert: func(t *testing.T, cfg *ConfigType) {
				assert.InDelta(t, 100.0, cfg.RestConfig.QPS, 0.0001, "rest config QPS should be set to 100")
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			cmconfig := &ConfigType{}
			_, done, err := config.MockSetupUtil(testCase.env, "cloudmanager.yaml", testCase.data, func(cfg config.Config) {
				InitConfig(cfg, cmconfig)
			})
			defer done()

			assert.NoError(t, err, "error setting up mocked config")

			testCase.assert(t, cmconfig)
		})

	}

}

func TestConfigType_GetManagerControllerConfig(t *testing.T) {

	t.Run("case 1 default values", func(t *testing.T) {
		cfg := &ConfigType{}
		cc := cfg.GetManagerControllerConfig()
		assert.Equal(t, 0, cc.MaxConcurrentReconciles)
		assert.Nil(t, cc.GroupKindConcurrency)
		assert.Equal(t, 0*time.Second, cc.ReconciliationTimeout)
	})

	t.Run("case 2 full spec", func(t *testing.T) {
		cfg := &ConfigType{
			Manager: ManagerConfig{
				Controller: ManagerControllerConfig{
					MaxConcurrentReconciles: 5,
					GroupKindConcurrency: map[string]int{
						"apps/Deployment": 3,
					},
					ReconciliationTimeout: 8 * time.Minute,
				},
			},
		}
		cc := cfg.GetManagerControllerConfig()
		assert.Equal(t, 5, cc.MaxConcurrentReconciles)
		assert.NotNil(t, cc.GroupKindConcurrency)
		assert.Equal(t, 3, cc.GroupKindConcurrency["apps/Deployment"])
		assert.Equal(t, 8*time.Minute, cc.ReconciliationTimeout)
	})

}

func TestConfigType_GetManagerPprofBindAddress(t *testing.T) {

	t.Run("case 1 pprof disabled by default", func(t *testing.T) {
		cfg := &ConfigType{}
		addr := cfg.GetManagerPprofBindAddress()
		assert.Equal(t, "", addr)
	})

	t.Run("case 2 pprof enabled", func(t *testing.T) {
		cfg := &ConfigType{
			Manager: ManagerConfig{
				Pprof: PprofConfig{
					Enabled: true,
					Address: ":8083",
				},
			},
		}
		addr := cfg.GetManagerPprofBindAddress()
		assert.Equal(t, ":8083", addr)
	})

}

func TestConfigType_ApplyToRestConfig(t *testing.T) {

	t.Run("case 1 default values", func(t *testing.T) {
		cfg := &ConfigType{
			RestConfig: RestConfig{},
		}
		rc := &rest.Config{
			QPS:   5.0,
			Burst: 10,
		}
		cfg.ApplyToRestConfig(rc)

		assert.InDelta(t, 5.0, rc.QPS, 0.0001)
		assert.Equal(t, 10, rc.Burst)
	})

	t.Run("case 2 custom values for QPS and Burst", func(t *testing.T) {
		cfg := &ConfigType{
			RestConfig: RestConfig{
				QPS:   99.0,
				Burst: 199,
			},
		}
		rc := &rest.Config{
			QPS:   5.0,
			Burst: 10,
		}
		cfg.ApplyToRestConfig(rc)

		assert.InDelta(t, 99.0, rc.QPS, 0.0001)
		assert.Equal(t, 199, rc.Burst)
	})

	t.Run("case 3 custom value for QPS", func(t *testing.T) {
		cfg := &ConfigType{
			RestConfig: RestConfig{
				QPS: 133.3,
			},
		}
		rc := &rest.Config{
			QPS:   5.0,
			Burst: 10,
		}
		cfg.ApplyToRestConfig(rc)

		assert.InDelta(t, 133.3, rc.QPS, 0.0001)
		assert.Equal(t, 10, rc.Burst) // default value for Burst
	})

	t.Run("case 4 custom values for Burst", func(t *testing.T) {
		cfg := &ConfigType{
			RestConfig: RestConfig{
				Burst: 321,
			},
		}
		rc := &rest.Config{
			QPS:   5.0,
			Burst: 10,
		}
		cfg.ApplyToRestConfig(rc)

		assert.InDelta(t, 5.0, rc.QPS, 0.0001) // default value for QPS
		assert.Equal(t, 321, rc.Burst)
	})
}
