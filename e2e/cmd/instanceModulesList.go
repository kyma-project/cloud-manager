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
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ModuleInfo struct {
	Alias     string
	RuntimeID string
	Name      string
	Spec      bool
	State     string
	Message   string
	CR        bool
	CrState   string
}

var cmdInstanceModulesList = &cobra.Command{
	Use: "list",
	RunE: func(cmd *cobra.Command, args []string) error {
		keb, err := e2ekeb.Create(rootCtx, config)
		if err != nil {
			return fmt.Errorf("failed to create keb: %w", err)
		}

		var runtimeIdList []string
		if len(runtimes) > 0 {
			runtimeIdList = append([]string{}, runtimes...)
		} else {
			idArr, err := keb.List(rootCtx)
			if err != nil {
				return fmt.Errorf("failed to list keb instances: %w", err)
			}
			for _, id := range idArr {
				if aliases == nil || pie.Contains(aliases, id.Alias) {
					runtimeIdList = append(runtimeIdList, id.RuntimeID)
				}
			}
		}

		tbl := table.New("ALias", "RuntimeID", "Module", "Kyma Spec", "Kyma State", "Kyma Message", "Module CR", "CR State").WithPadding(4)
		var data []*ModuleInfo

		for _, rtID := range runtimeIdList {

			id, err := keb.GetInstance(rootCtx, rtID)
			if err != nil {
				return fmt.Errorf("failed to get instance %q: %w", rtID, err)
			}

			clnt, err := keb.CreateInstanceClient(rootCtx, rtID)
			if err != nil {
				return err
			}

			skrKyma := &operatorv1beta2.Kyma{}
			err = clnt.Get(rootCtx, types.NamespacedName{
				Namespace: "kyma-system",
				Name:      "default",
			}, skrKyma)
			if err != nil {
				return fmt.Errorf("failed to get SKR kyma: %w", err)
			}

			dashIt := func(s string) string {
				if s == "" {
					return "-"
				}
				return s
			}

			localData := map[string]*ModuleInfo{}

			for _, m := range skrKyma.Status.Modules {
				if len(modules) > 0 && !pie.Contains(modules, m.Name) {
					continue
				}
				localData[m.Name] = &ModuleInfo{
					Alias:     id.Alias,
					RuntimeID: id.RuntimeID,
					Name:      m.Name,
					Spec:      false,
					State:     dashIt(string(m.State)),
					Message:   dashIt(m.Message),
					CR:        false,
					CrState:   "",
				}
			}
			for _, m := range skrKyma.Spec.Modules {
				if len(modules) > 0 && !pie.Contains(modules, m.Name) {
					continue
				}
				mi, ok := localData[m.Name]
				if !ok {
					localData[m.Name] = &ModuleInfo{
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

			for _, mi := range localData {
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

			for _, mi := range localData {
				data = append(data, mi)
			}

		} // for runtime

		for _, mi := range data {
			tbl.AddRow(mi.Alias, mi.RuntimeID, mi.Name, mi.Spec, mi.State, mi.Message, mi.CR, mi.CrState)
		}

		tbl.Print()
		fmt.Println("")

		return nil
	},
}

func init() {
	cmdInstanceModules.AddCommand(cmdInstanceModulesList)
	cmdInstanceModulesList.Flags().StringSliceVarP(&runtimes, "runtime-id", "r", nil, "The runtime ID")
	cmdInstanceModulesList.Flags().StringSliceVarP(&aliases, "alias", "a", nil, "The runtime alias")
	cmdInstanceModulesList.Flags().StringSliceVarP(&modules, "module-name", "m", nil, "The module name")
	cmdInstanceModulesList.MarkFlagsMutuallyExclusive("runtime-id", "alias")
}
