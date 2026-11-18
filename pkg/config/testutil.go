package config

import (
	"fmt"
	"os"

	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"gopkg.in/yaml.v3"
)

func copyFile(src, dest string) error {
	buf, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	err = os.WriteFile(dest, buf, 0644)
	if err != nil {
		return err
	}
	return nil
}

// MockSetupUtil creates a Config instance from the given environment variables and if filename fn is not empty it creates
// a temporary directory and writes the given data as a yaml file with the given filename fn in that directory.
// It then sets the base directory of the Config instance to that temporary directory. If initcb is not nil, it calls
// initcb with the Config instance and then reads the configuration from the sources. It returns the Config instance and
// any error that occurred during the process. It returns done function that should be called to clean up the temporary
// directory if it was created.
// Example usage:
//
//	env := map[string]string{...}
//	data := map[string]interface{}{...}
//	cfg, done, err := MockSetupUtil(env, "config.yaml", data, func(cfg Config) {
//		InitConfig(cfg)
//	})
//	if err != nil {
//		// handle error
//	}
//	// test my config
//	defer done()
func MockSetupUtil(envvars map[string]string, fn string, data map[string]any, initcb func(Config)) (Config, func(), error) {
	env := abstractions.NewMockedEnvironment(envvars)
	cfg := NewConfig(env)
	done := func() {}

	if fn != "" {
		dir, err := os.MkdirTemp("", fmt.Sprintf("cloud-manager-config-mock-%s", fn))
		if err != nil {
			return nil, done, fmt.Errorf("error creating tmp dir: %w", err)
		}
		done = func() {
			_ = os.RemoveAll(dir)
		}
		b, err := yaml.Marshal(data)
		if err != nil {
			_ = os.RemoveAll(dir)
			return nil, done, fmt.Errorf("error marshalling data to yaml: %w", err)
		}
		err = os.WriteFile(fmt.Sprintf("%s/%s", dir, fn), b, 0644)
		if err != nil {
			_ = os.RemoveAll(dir)
			return nil, done, fmt.Errorf("error writing yaml file: %w", err)
		}

		cfg.BaseDir(dir)
	}

	if initcb != nil {
		initcb(cfg)
		cfg.Read()
	}

	return cfg, done, nil
}
