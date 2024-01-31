package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"os"
	"path/filepath"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func init() {
	moduleStateCmd.Flags().StringVarP(&kymaName, "kymaName", "k", "", "Kyma CR name")
	moduleStateCmd.Flags().StringVarP(&namespace, "namespace", "n", "kcp-system", "Kyma CR namespace")
	moduleStateCmd.Flags().StringVarP(&moduleName, "module", "m", "cloud-manager", "Module name")
	moduleStateCmd.Flags().StringVarP(&moduleState, "state", "s", "", "Module state, one of Ready|Error|Processing|Deleting|Warning")
	moduleStateCmd.Flags().BoolVarP(&removeModule, "remove", "r", false, "Remove module")
	moduleStateCmd.Flags().BoolVarP(&listModule, "list", "l", false, "List module state")
	_ = moduleStateCmd.MarkFlagRequired("kymaName")
	rootCmd.AddCommand(moduleStateCmd)
}

var (
	kymaName     string
	namespace    string
	moduleName   string = "cloud-manager"
	moduleState  string
	removeModule bool
	listModule   bool
)

var moduleStateCmd = &cobra.Command{
	Use:     "module-state",
	Short:   "Change module state",
	Aliases: []string{"ms"},
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(moduleName) == 0 {
			return errors.New("module name is required")
		}

		kubeconfig := os.Getenv("KUBECONFIG")
		if len(kubeconfig) == 0 {
			home := homedir.HomeDir()
			if home == "" {
				return errors.New("Unable to locate kubeconfig, use KUBECONFIG env var")
			}
			kubeconfig = filepath.Join(home, ".kube", "config")
		}
		cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return fmt.Errorf("error loading kubeconfig: %w", err)
		}

		c, err := client.New(cfg, client.Options{})
		if err != nil {
			return fmt.Errorf("error creating client: %w", err)
		}

		kymaCR := util.NewKymaUnstructured()
		err = c.Get(context.Background(), types.NamespacedName{
			Namespace: namespace,
			Name:      kymaName,
		}, kymaCR)
		if err != nil {
			return fmt.Errorf("error loading Kyma CR: %w", err)
		}

		if removeModule {
			err = util.RemoveKymaModuleState(kymaCR, moduleName)
			if err != nil {
				return fmt.Errorf("error removing kyma module %s from state: %w", moduleName, err)
			}
			err = c.Status().Update(context.Background(), kymaCR)
			if err != nil {
				return fmt.Errorf("error updating Kyma CR: %w", err)
			}
			fmt.Printf("Kyma module %s removed from state\n", moduleName)
			return nil
		}

		if listModule || len(moduleState) == 0 {
			fmt.Printf("Module %s state is: %s\n", moduleName, util.GetKymaModuleState(kymaCR, moduleName))
			return nil
		}

		err = util.SetKymaModuleState(kymaCR, moduleName, util.KymaModuleState(moduleState))
		if err != nil {
			return fmt.Errorf("error setting kyma module %s state to %s: %w", moduleName, moduleState, err)
		}
		err = c.Status().Update(context.Background(), kymaCR)
		if err != nil {
			return fmt.Errorf("error updating Kyma CR: %w", err)
		}
		fmt.Printf("Kyma module %s state set to %s\n", moduleName, moduleState)
		return nil
	},
}
