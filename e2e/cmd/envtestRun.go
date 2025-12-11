package main

import (
	"fmt"
	"os"
	"path/filepath"
	goruntime "runtime"

	"github.com/kyma-project/cloud-manager/pkg/testinfra"
	"github.com/spf13/cobra"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

var envtestK8sVersion string
var envtestKubeconfigOutput string

var cmdEnvtestRun = &cobra.Command{
	Use: "run",
	RunE: func(cmd *cobra.Command, args []string) error {
		if envtestK8sVersion == "" {
			return fmt.Errorf("envtest k8s version is required, either set env var ENVTEST_K8S_VERSION or provide --version flag")
		}
		env := &envtest.Environment{
			BinaryAssetsDirectory: filepath.Join(configDir, "bin", "k8s",
				fmt.Sprintf("%s-%s-%s", envtestK8sVersion, goruntime.GOOS, goruntime.GOARCH)),
		}
		restConfig, err := env.Start()
		if err != nil {
			return fmt.Errorf("failed to start envtest: %w", err)
		}

		fmt.Println("envtest started")

		b, err := testinfra.KubeconfigToBytes(testinfra.RestConfigToKubeconfig(restConfig))
		if err != nil {
			return fmt.Errorf("error creating kubeconfig bytes: %w", err)
		}

		if err := os.WriteFile(envtestKubeconfigOutput, b, 0644); err != nil {
			return fmt.Errorf("error writing kubeconfig to file %s: %w", envtestKubeconfigOutput, err)
		}

		fmt.Printf("kubeconfig written to %s\n", envtestKubeconfigOutput)

		fmt.Println("running...")

		<-rootCtx.Done()

		fmt.Println("stopping...")

		if err := env.Stop(); err != nil {
			return fmt.Errorf("error stopping envtest: %w", err)
		}

		fmt.Println("stopped")

		fmt.Println("deleting kubeconfig...")

		if err := os.Remove(envtestKubeconfigOutput); err != nil {
			return fmt.Errorf("error deleting kubeconfig %s: %w", envtestKubeconfigOutput, err)
		}

		fmt.Println("kubeconfig deleted")

		return nil
	},
}

func init() {
	cmdEnvtest.AddCommand(cmdEnvtestRun)
	cmdEnvtestRun.Flags().StringVarP(&envtestK8sVersion, "version", "v", os.Getenv("ENVTEST_K8S_VERSION"), "envtest version")
	cmdEnvtestRun.Flags().StringVarP(&envtestKubeconfigOutput, "output", "o", "", "output kubeconfig file")
	_ = cmdEnvtestRun.MarkFlagRequired("output")
}
