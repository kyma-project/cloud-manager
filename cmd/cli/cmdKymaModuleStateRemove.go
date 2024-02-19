package main

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/cmd/cli/helper"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func init() {
	cmdKymaModuleState.AddCommand(cmdKymaModuleStateRemove)
}

var cmdKymaModuleStateRemove = &cobra.Command{
	Use:     "remove",
	Aliases: []string{"r"},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		mustAll(
			requiredKymaName(),
			requiredKymaModuleName(),
			defaultKcpNamespace(),
		)

		c := helper.NewKcpClient()

		var kymaCR *unstructured.Unstructured
		if kymaCR, err = helper.LoadKymaCRWithClient(c, namespace, kymaName); err != nil {
			return
		}

		err = util.RemoveKymaModuleStateFromStatus(kymaCR, kymaModuleName)
		if err != nil {
			return fmt.Errorf("error removing kyma module %s from state: %w", kymaModuleName, err)
		}
		err = c.Status().Update(context.Background(), kymaCR)
		if err != nil {
			return fmt.Errorf("error updating Kyma CR: %w", err)
		}
		fmt.Printf("Kyma module %s removed from state\n", kymaModuleName)

		return nil
	},
}
