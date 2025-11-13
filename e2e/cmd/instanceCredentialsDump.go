package main

import (
	"fmt"
	"time"

	e2ekeb "github.com/kyma-project/cloud-manager/e2e/keb"
	"github.com/spf13/cobra"
)

var cmdInstanceCredentialsDump = &cobra.Command{
	Use: "dump",
	RunE: func(cmd *cobra.Command, args []string) error {
		keb, err := e2ekeb.Create(rootCtx, config)
		if err != nil {
			return fmt.Errorf("failed to create keb: %w", err)
		}

		data, expiresAt, err := keb.GetInstanceKubeconfig(rootCtx, runtimeID)
		if err != nil {
			return err
		}

		fmt.Println("")
		fmt.Printf("# expires at %s\n", expiresAt.Format(time.RFC3339))
		fmt.Println(string(data))
		fmt.Println("")

		return nil
	},
}

func init() {
	cmdInstanceCredentials.AddCommand(cmdInstanceCredentialsDump)
	cmdInstanceCredentialsDump.Flags().StringVarP(&runtimeID, "runtime-id", "r", "", "The runtime ID")
	_ = cmdInstanceCredentialsDump.MarkFlagRequired("runtime-id")
}
