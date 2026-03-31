package main

import (
	"context"
	"fmt"
	"os"

	"github.com/go-logr/logr"
	e2econfig "github.com/kyma-project/cloud-manager/e2e/config"
	commonconfig "github.com/kyma-project/cloud-manager/pkg/common/config"
	"github.com/spf13/cobra"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

// var world e2e.WorldIntf
var rootCtx context.Context
var rootLogger logr.Logger

var configDir string
var config *e2econfig.ConfigType

var cmdRoot = &cobra.Command{
	Use:   "e2e",
	Short: "E2E command tool",
	PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
		opts := zap.Options{}
		opts.Development = true
		rootLogger = zap.New(zap.UseFlagOptions(&opts))
		ctrl.SetLogger(rootLogger)

		if configDir != "" {
			_ = os.Setenv("CONFIG_DIR", configDir)
		}
		config = e2econfig.LoadConfig()

		_ = commonconfig.CreateNewConfigAndLoad()

		return nil
	},
}

func init() {
	cmdRoot.PersistentFlags().StringVar(&configDir, "config-dir", "", "Path to the directory containing e2econfig.yaml file")
	cmdRoot.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")
}

func main() {
	ctx, cancel := context.WithCancel(ctrl.SetupSignalHandler())
	defer cancel()
	rootCtx = ctx

	if err := cmdRoot.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
