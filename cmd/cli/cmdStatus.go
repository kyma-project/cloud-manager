package main

import (
	"errors"
	"os"

	"github.com/goccy/go-yaml"
	"github.com/spf13/cobra"
)

var (
	apiVersion string
	kind       string
	name       string
	file       string
)

func init() {
	cmdStatus.PersistentFlags().StringVarP(&apiVersion, "api-version", "v", "cloud-control.kyma-project.io/v1beta1", "API version")
	cmdStatus.PersistentFlags().StringVarP(&kind, "kind", "k", "RedisInstance", "Kind of instance")
	cmdStatus.PersistentFlags().StringVarP(&name, "name", "", "", "Name of instance")
	cmdStatus.PersistentFlags().StringVarP(&file, "file", "f", "", "Patch file path")
	cmdRoot.AddCommand(cmdStatus)
}

var cmdStatus = &cobra.Command{
	Use:     "status",
	Aliases: []string{"s", "st"},
}

func validateStatusFlags() error {
	if apiVersion == "" {
		return errors.New("apiVersion flag is required")
	}
	if kind == "" {
		return errors.New("kind flag is required")
	}
	if name == "" {
		return errors.New("name flag is required")
	}
	if file == "" {
		return errors.New("file flag is required")
	}

	if namespace == "" {
		namespace = "default"
	}

	return nil
}

func loadStatusFile() (map[string]interface{}, error) {
	buf, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	data := map[string]interface{}{}
	if err := yaml.Unmarshal(buf, &data); err != nil {
		return nil, err
	}
	return data, nil
}
