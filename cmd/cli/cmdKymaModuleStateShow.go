package main

import (
	"fmt"
	"github.com/kyma-project/cloud-manager/cmd/cli/helper"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func init() {
	cmdKymaModuleState.AddCommand(cmdKymaModuleStateShow)
}

var cmdKymaModuleStateShow = &cobra.Command{
	Use:     "show",
	Aliases: []string{"s"},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		mustAll(
			requiredKymaName(),
			requiredKymaModuleName(),
			defaultKcpNamespace(),
		)

		var kymaCR *unstructured.Unstructured
		if kymaCR, err = helper.LoadKymaCR(namespace, kymaName); err != nil {
			return
		}

		fmt.Printf("Module %s state is: %s\n", kymaModuleName, util.GetKymaModuleState(kymaCR, kymaModuleName))

		return nil
	},
}
