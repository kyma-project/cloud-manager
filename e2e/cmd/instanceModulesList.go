package main

import (
	"fmt"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	e2ekeb "github.com/kyma-project/cloud-manager/e2e/keb"
	"github.com/kyma-project/cloud-manager/pkg/external/operatorv1beta2"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ModuleInfo struct {
	Name    string
	Spec    bool
	State   string
	Message string
	CR      bool
	CrState string
}

var cmdInstanceModulesList = &cobra.Command{
	Use: "list",
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

		tbl := table.New("Module", "Kyma Spec", "Kyma State", "Kyma Message", "Module CR", "CR State").WithPadding(4)
		data := map[string]*ModuleInfo{}

		dashIt := func(s string) string {
			if s == "" {
				return "-"
			}
			return s
		}

		for _, m := range kyma.Status.Modules {
			data[m.Name] = &ModuleInfo{
				Name:    m.Name,
				Spec:    false,
				State:   dashIt(string(m.State)),
				Message: dashIt(m.Message),
				CR:      false,
				CrState: "",
			}
		}
		for _, m := range kyma.Spec.Modules {
			mi, ok := data[m.Name]
			if !ok {
				data[m.Name] = &ModuleInfo{
					Name:    m.Name,
					Spec:    true,
					State:   dashIt(""),
					Message: dashIt(""),
					CR:      false,
					CrState: dashIt(""),
				}
			} else {
				mi.Spec = true
			}
		}

		for _, mi := range data {
			if mi.Name == "cloud-manager" {
				cr := &cloudresourcesv1beta1.CloudResources{}
				err = clnt.Get(rootCtx, types.NamespacedName{
					Namespace: "kyma-system",
					Name:      "default",
				}, cr)
				if util.IgnoreNoMatch(client.IgnoreNotFound(err)) != nil {
					continue
				}
				if err != nil {
					return fmt.Errorf("failed to get CloudResources: %w", err)
				}
				mi.CR = true
				mi.CrState = dashIt(string(cr.Status.State))
			}
		}

		for _, mi := range data {
			tbl.AddRow(mi.Name, mi.Spec, mi.State, mi.Message, mi.CR, mi.CrState)
		}

		tbl.Print()
		fmt.Println("")

		return nil
	},
}

func init() {
	cmdInstanceModules.AddCommand(cmdInstanceModulesList)
	cmdInstanceModulesList.Flags().StringVarP(&runtimeID, "runtime-id", "r", "", "The runtime ID")
	_ = cmdInstanceModulesList.MarkFlagRequired("runtime-id")
}
