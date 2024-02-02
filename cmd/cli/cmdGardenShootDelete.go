package main

import (
	"context"
	"fmt"
	gardenerTypes "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/kyma-project/cloud-manager/cmd/cli/helper"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	cmdGardenShoot.AddCommand(cmdGardenShootDelete)
}

var cmdGardenShootDelete = &cobra.Command{
	Use:     "delete",
	Aliases: []string{"d"},
	RunE: func(cmd *cobra.Command, args []string) error {
		mustAll(
			requiredShootName(),
			defaultGardenNamespace(),
		)

		c := helper.NewGardenClient()

		shoot := &gardenerTypes.Shoot{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      shootName,
			},
		}
		err := c.Delete(context.Background(), shoot)
		if err != nil {
			return fmt.Errorf("error deleting Shoot: %w", err)
		}
		fmt.Println("Deleted Shoot")

		sb := &gardenerTypes.SecretBinding{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      shootName,
			},
		}
		err = c.Delete(context.Background(), sb)
		if err != nil {
			return fmt.Errorf("error deleting SecretBinding: %w", err)
		}
		fmt.Println("Deleted SecretBinding")

		s := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      shootName,
			},
		}
		err = c.Delete(context.Background(), s)
		if err != nil {
			return fmt.Errorf("error deleting Secret: %w", err)
		}
		fmt.Println("Deleted Secret")

		return nil
	},
}
