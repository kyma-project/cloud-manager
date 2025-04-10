package main

import (
	"context"
	"fmt"
	gardenertypes "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/kyma-project/cloud-manager/cmd/cli/helper"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/types"
)

func init() {
	cmdGardenShoot.AddCommand(cmdGardenShootList)
}

var cmdGardenShootList = &cobra.Command{
	Use:     "list",
	Aliases: []string{"l"},
	RunE: func(cmd *cobra.Command, args []string) error {
		c := helper.NewGardenClient()
		list := &gardenertypes.ShootList{}
		err := c.List(context.Background(), list)
		if err != nil {
			return fmt.Errorf("error listing shoots: %w", err)
		}

		sb := &gardenertypes.SecretBinding{}
		fmt.Println("Namespace \t\t Shoot \t\t Provider \t Secret")
		for _, shoot := range list.Items {
			var provider string
			var secret string
			if shoot.Spec.SecretBindingName != nil {
				err = c.Get(context.Background(), types.NamespacedName{
					Namespace: shoot.Namespace,
					Name:      *shoot.Spec.SecretBindingName,
				}, sb)
				if err != nil {
					return fmt.Errorf("error loading SecretBinding %s for Shoot %s/%s", *shoot.Spec.SecretBindingName, shoot.Namespace, shoot.Name)
				}
				if sb.Provider != nil {
					provider = sb.Provider.Type
				}
				secret = sb.SecretRef.Name
			}

			fmt.Printf("%s \t %s \t %s \t\t %s\n", shoot.Namespace, shoot.Name, provider, secret)

		}

		return nil
	},
}
