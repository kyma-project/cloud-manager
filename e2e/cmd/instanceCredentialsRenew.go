package main

import (
	"fmt"

	e2ekeb "github.com/kyma-project/cloud-manager/e2e/keb"
	"github.com/spf13/cobra"
)

type cmdInstanceCredentialsRenewOptionsType struct {
	runtimeID string
}

var cmdInstanceCredentialsRenewOptions cmdInstanceCredentialsRenewOptionsType

var cmdInstanceCredentialsRenew = &cobra.Command{
	Use: "renew",
	RunE: func(cmd *cobra.Command, args []string) error {
		keb, err := e2ekeb.Create(rootCtx, config)
		if err != nil {
			return fmt.Errorf("failed to create keb: %w", err)
		}

		err = keb.RenewInstanceKubeconfig(rootCtx, cmdInstanceCredentialsRenewOptions.runtimeID)
		if err != nil {
			return err
		}

		fmt.Println("")
		fmt.Println("New credential created")

		return nil
	},
}

func init() {
	cmdInstanceCredentials.AddCommand(cmdInstanceCredentialsRenew)
	cmdInstanceCredentialsRenew.Flags().StringVarP(&cmdInstanceCredentialsRenewOptions.runtimeID, "runtime-id", "r", "", "The runtime ID")
	_ = cmdInstanceCredentialsRenew.MarkFlagRequired("runtime-id")
}
