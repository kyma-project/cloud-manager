package main

import (
	"fmt"

	"github.com/elliotchance/pie/v2"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	e2ekeb "github.com/kyma-project/cloud-manager/e2e/keb"
	"github.com/kyma-project/cloud-manager/pkg/external/operatorv1beta2"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ModuleInfo struct {
	Alias     string
	RuntimeID string
	Name      string
	Spec      *bool
	Status    *bool
	State     string
	Message   string
	CrState   string
}

func (mi ModuleInfo) ToTableRows() []any {
	bts := func(b *bool) string {
		if b == nil {
			return ""
		}
		if *b {
			return "true"
		}
		return "false"
	}
	return []any{
		mi.Alias,
		mi.RuntimeID,
		mi.Name,
		bts(mi.Spec),
		bts(mi.Status),
		mi.State,
		mi.Message,
		mi.CrState,
	}
}

type cmdInstanceModulesListOptionsType struct {
	runtimeID string
	alias     string
	modules   []string
}

var cmdInstanceModulesListOptions cmdInstanceModulesListOptionsType

var cmdInstanceModulesList = &cobra.Command{
	Use: "list",
	RunE: func(cmd *cobra.Command, args []string) error {
		keb, err := e2ekeb.Create(rootCtx, config)
		if err != nil {
			return fmt.Errorf("failed to create keb: %w", err)
		}

		var opts []e2ekeb.ListOption
		if cmdInstanceModulesListOptions.runtimeID != "" {
			opts = append(opts, e2ekeb.WithRuntime(cmdInstanceModulesListOptions.runtimeID))
		}
		if cmdInstanceModulesListOptions.alias != "" {
			opts = append(opts, e2ekeb.WithAlias(cmdInstanceModulesListOptions.alias))
		}

		arr, err := keb.List(rootCtx, opts...)
		if err != nil {
			return fmt.Errorf("failed to list instances: %w", err)
		}

		tbl := table.New("Alias", "Runtime", "Module", "Spec", "Status", "State", "Message", "CR State").WithPadding(4)

		for _, id := range arr {

			kcpKyma := &operatorv1beta2.Kyma{}
			err = keb.KcpClient().Get(rootCtx, types.NamespacedName{
				Namespace: keb.Config().KcpNamespace,
				Name:      id.RuntimeID,
			}, kcpKyma)
			if err != nil {
				tbl.AddRow(id.Alias, id.RuntimeID, "", "", "", "", err.Error())
				continue
			}

			rtModules := map[string]*ModuleInfo{}

			for _, m := range kcpKyma.Status.Modules {
				if len(cmdInstanceModulesListOptions.modules) > 0 && !pie.Contains(cmdInstanceModulesListOptions.modules, m.Name) {
					continue
				}
				mi := &ModuleInfo{
					Alias:     id.Alias,
					RuntimeID: id.RuntimeID,
					Name:      m.Name,
					Status:    ptr.To(true),
					State:     string(m.State),
					Message:   m.Message,
				}
				rtModules[m.Name] = mi
			}
			for _, m := range kcpKyma.Spec.Modules {
				if len(cmdInstanceModulesListOptions.modules) > 0 && !pie.Contains(cmdInstanceModulesListOptions.modules, m.Name) {
					continue
				}
				mi, ok := rtModules[m.Name]
				if !ok {
					rtModules[m.Name] = &ModuleInfo{
						Alias:     id.Alias,
						RuntimeID: id.RuntimeID,
						Name:      m.Name,
						Spec:      ptr.To(true),
					}
				} else {
					mi.Spec = ptr.To(true)
				}
			}

			for _, mi := range rtModules {
				if mi.Name == "cloud-manager" {
					clnt, err := keb.CreateInstanceClient(rootCtx, id.RuntimeID)
					if err != nil {
						mi.CrState = err.Error()
						break
					}
					cr := &cloudresourcesv1beta1.CloudResources{}
					err = clnt.Get(rootCtx, types.NamespacedName{
						Namespace: "kyma-system",
						Name:      "default",
					}, cr)
					if util.IgnoreNoMatch(client.IgnoreNotFound(err)) != nil {
						continue
					}
					if err != nil {
						mi.CrState = err.Error()
						break
					}
					mi.CrState = string(cr.Status.State)
					break
				}
			}

			for _, mi := range rtModules {
				tbl.AddRow(mi.ToTableRows()...)
			}
			if len(rtModules) == 0 {
				tbl.AddRow(id.Alias, id.RuntimeID, "-")
			}

		} // for runtime

		tbl.Print()
		fmt.Println("")

		return nil
	},
}

func init() {
	cmdInstanceModules.AddCommand(cmdInstanceModulesList)
	cmdInstanceModulesList.Flags().StringVarP(&cmdInstanceModulesListOptions.runtimeID, "runtime-id", "r", "", "The runtime ID")
	cmdInstanceModulesList.Flags().StringVarP(&cmdInstanceModulesListOptions.alias, "alias", "a", "", "The runtime alias")
	cmdInstanceModulesList.Flags().StringArrayVarP(&cmdInstanceModulesListOptions.modules, "module-name", "m", nil, "The module name")
	cmdInstanceModulesList.MarkFlagsMutuallyExclusive("runtime-id", "alias")
}
