package main

import (
	"fmt"

	e2ekeb "github.com/kyma-project/cloud-manager/e2e/keb"
	"github.com/kyma-project/cloud-manager/pkg/external/operatorv1beta2"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/types"
)

var cmdInstanceKyma = &cobra.Command{
	Use: "kyma",
	RunE: func(cmd *cobra.Command, args []string) error {
		keb, err := e2ekeb.Create(rootCtx, config)
		if err != nil {
			return fmt.Errorf("failed to create keb: %w", err)
		}

		clnt, err := keb.CreateInstanceClient(rootCtx, runtimeID)
		if err != nil {
			return err
		}

		kyma := &operatorv1beta2.Kyma{}
		err = clnt.Get(rootCtx, types.NamespacedName{
			Namespace: "kyma-system",
			Name:      "default",
		}, kyma)
		if err != nil {
			return fmt.Errorf("failed to get SKR kyma: %w", err)
		}

		txt, err := yaml.Marshal(kyma)
		if err != nil {
			return fmt.Errorf("failed to marshal SKR Kyma object: %w", err)
		}

		fmt.Println(string(txt))

		return nil
	},
}

func init() {
	cmdInstance.AddCommand(cmdInstanceKyma)
	cmdInstanceKyma.Flags().StringVarP(&runtimeID, "runtime-id", "r", "", "The runtime ID")
	_ = cmdInstanceModulesList.MarkFlagRequired("runtime-id")
}
