package main

import (
	"fmt"
	"github.com/elliotchance/pie/v2"
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

		if runtimeID == "" {
			idArr, err := keb.List(rootCtx, e2ekeb.WithAlias(alias))
			if err != nil {
				return fmt.Errorf("failed to list runtimes: %w", err)
			}
			if len(idArr) == 0 {
				return fmt.Errorf("runtime with alias %q not found", alias)
			}
			if len(idArr) > 1 {
				return fmt.Errorf("multiple runtimes with alias %q found: %v", alias, pie.Map(idArr, func(x e2ekeb.InstanceDetails) string {
					return x.RuntimeID
				}))
			}
			runtimeID = idArr[0].RuntimeID
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
	cmdInstanceCredentialsDump.Flags().StringVarP(&alias, "alias", "a", "", "The runtime alias")
	cmdInstanceCredentialsDump.MarkFlagsMutuallyExclusive("runtime-id", "alias")
	cmdInstanceCredentialsDump.MarkFlagsOneRequired("runtime-id", "alias")
}
