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

type cmdEnvtestRunOptionsType struct {
	k8sVersion       string
	kubeconfigOutput string
}

var cmdEnvtestRunOptions cmdEnvtestRunOptionsType

var cmdEnvtestRun = &cobra.Command{
	Use: "run",
	RunE: func(cmd *cobra.Command, args []string) error {
		if cmdEnvtestRunOptions.k8sVersion == "" {
			return fmt.Errorf("envtest k8s version is required, either set env var ENVTEST_K8S_VERSION or provide --version flag")
		}
		env := &envtest.Environment{
			BinaryAssetsDirectory: filepath.Join(configDir, "bin", "k8s",
				fmt.Sprintf("%s-%s-%s", cmdEnvtestRunOptions.k8sVersion, goruntime.GOOS, goruntime.GOARCH)),
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

		if err := os.WriteFile(cmdEnvtestRunOptions.kubeconfigOutput, b, 0644); err != nil {
			return fmt.Errorf("error writing kubeconfig to file %s: %w", cmdEnvtestRunOptions.kubeconfigOutput, err)
		}

		fmt.Printf("kubeconfig written to %s\n", cmdEnvtestRunOptions.kubeconfigOutput)

		fmt.Println("running...")

		<-rootCtx.Done()

		fmt.Println("stopping...")

		if err := env.Stop(); err != nil {
			return fmt.Errorf("error stopping envtest: %w", err)
		}

		fmt.Println("stopped")

		fmt.Println("deleting kubeconfig...")

		if err := os.Remove(cmdEnvtestRunOptions.kubeconfigOutput); err != nil {
			return fmt.Errorf("error deleting kubeconfig %s: %w", cmdEnvtestRunOptions.kubeconfigOutput, err)
		}

		fmt.Println("kubeconfig deleted")

		return nil
	},
}

func init() {
	cmdEnvtest.AddCommand(cmdEnvtestRun)
	cmdEnvtestRun.Flags().StringVarP(&cmdEnvtestRunOptions.k8sVersion, "version", "", os.Getenv("ENVTEST_K8S_VERSION"), "envtest version")
	cmdEnvtestRun.Flags().StringVarP(&cmdEnvtestRunOptions.kubeconfigOutput, "output", "o", "", "output kubeconfig file")
	_ = cmdEnvtestRun.MarkFlagRequired("output")
}
