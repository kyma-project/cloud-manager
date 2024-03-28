package main

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/cmd/cli/helper"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var (
	moduleState string
)

func init() {
	cmdKymaModuleStateUpdate.PersistentFlags().StringVarP(&moduleState, "state", "s", "", "Module state, one of Ready|Error|Processing|Deleting|Warning")
	_ = cmdKymaModuleState.MarkPersistentFlagRequired("state")
	cmdKymaModuleState.AddCommand(cmdKymaModuleStateUpdate)
}

var cmdKymaModuleStateUpdate = &cobra.Command{
	Use:     "update",
	Aliases: []string{"u"},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		mustAll(
			requiredKymaName(),
			requiredKymaModuleName(),
			requiredKymaModuleState(),
			defaultKcpNamespace(),
		)

		c := helper.NewKcpClient()

		var kymaCR *unstructured.Unstructured
		if kymaCR, err = helper.LoadKymaCRWithClient(c, namespace, kymaName); err != nil {
			return
		}

		err = util.SetKymaModuleStateToStatus(kymaCR, kymaModuleName, util.KymaModuleState(moduleState))
		if err != nil {
			return fmt.Errorf("error setting kyma module %s state to %s: %w", kymaModuleName, moduleState, err)
		}
		err = c.Status().Update(context.Background(), kymaCR)
		if err != nil {
			return fmt.Errorf("error updating Kyma CR: %w", err)
		}
		fmt.Printf("Kyma module %s state set to %s\n", kymaModuleName, moduleState)

		return nil

	},
}
