package main

import (
	"context"
	"fmt"

	"github.com/elliotchance/pie/v2"
	e2ekeb "github.com/kyma-project/cloud-manager/e2e/keb"
	e2elib "github.com/kyma-project/cloud-manager/e2e/lib"
	"github.com/kyma-project/cloud-manager/pkg/external/infrastructuremanagerv1"
	"github.com/kyma-project/cloud-manager/pkg/external/operatorv1beta2"
	"github.com/spf13/cobra"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var cmdInstanceIgnore = &cobra.Command{
	Use:     "ignore",
	Aliases: []string{"i"},
	RunE: func(cmd *cobra.Command, args []string) error {

		keb, err := e2ekeb.Create(rootCtx, config)
		if err != nil {
			return fmt.Errorf("failed to create keb: %w", err)
		}

		setIgnoreLabel := func(ctx context.Context, obj client.Object, shouldIgnore bool) error {
			if shouldIgnore {
				if _, ok := obj.GetLabels()[e2elib.DoNotReconcile]; ok {
					return nil
				}
				if obj.GetLabels() == nil {
					obj.SetLabels(make(map[string]string))
				}
				obj.GetLabels()[e2elib.DoNotReconcile] = "true"
			} else {
				if _, ok := obj.GetLabels()[e2elib.DoNotReconcile]; !ok {
					return nil
				}
				delete(obj.GetLabels(), e2elib.DoNotReconcile)
			}
			return keb.KcpClient().Update(ctx, obj)
		}

		rtList := &infrastructuremanagerv1.RuntimeList{}
		err = keb.KcpClient().List(rootCtx, rtList)
		if err != nil {
			return fmt.Errorf("failed to list runtimes: %w", err)
		}

		for _, rt := range rtList.Items {
			rtAlias := rt.Labels[e2elib.AliasLabel]

			if listOnly {
				txt := "ok"
				if _, ok := rt.GetLabels()[e2elib.DoNotReconcile]; ok {
					txt = "ignore"
				}
				fmt.Printf("%s   %s   %s\n", rt.Name, rtAlias, txt)
				continue
			}

			if rt.Labels == nil {
				rt.Labels = map[string]string{}
			}
			shouldIgnore := ignoreAll
			if pie.Contains(ignoreAliases, rtAlias) || pie.Contains(ignoreAliases, rt.Name) {
				shouldIgnore = true
			}

			if shouldIgnore {
				fmt.Printf("Ignoring  %s   %s\n", rt.Name, rtAlias)
			} else {
				fmt.Printf("Enabling  %s   %s\n", rt.Name, rtAlias)
			}

			if err := setIgnoreLabel(rootCtx, &rt, shouldIgnore); err != nil {
				return fmt.Errorf("error handling Runtime %q %q", rtAlias, rt.Name)
			}

			gc := &infrastructuremanagerv1.GardenerCluster{}
			if err := keb.KcpClient().Get(rootCtx, client.ObjectKeyFromObject(&rt), gc); err != nil {
				return fmt.Errorf("failed to get GardenCluster: %w", err)
			}
			if err := setIgnoreLabel(rootCtx, gc, shouldIgnore); err != nil {
				return fmt.Errorf("error handling GardenCluster %q %q", rtAlias, rt.Name)
			}

			kymaKcp := &operatorv1beta2.Kyma{}
			if err := keb.KcpClient().Get(rootCtx, client.ObjectKeyFromObject(&rt), kymaKcp); err != nil {
				return fmt.Errorf("failed to get KCP Kyma: %w", err)
			}
			if err := setIgnoreLabel(rootCtx, kymaKcp, shouldIgnore); err != nil {
				return fmt.Errorf("error handling KCP Kyma %q %q", rtAlias, rt.Name)
			}
			kymaKcp.Spec.Modules = nil
			if err := keb.KcpClient().Update(rootCtx, kymaKcp); err != nil {
				return fmt.Errorf("failed to remove modules from spec of KCP Kyma %q %q: %w", rtAlias, rt.Name, err)
			}
			kymaKcp.Status.Modules = nil
			if err := keb.KcpClient().Status().Update(rootCtx, kymaKcp); err != nil {
				return fmt.Errorf("failed to remove modules from status of KCP Kyma %q %q: %w", rtAlias, rt.Name, err)
			}
		}

		return nil
	},
}

func init() {
	cmdInstance.AddCommand(cmdInstanceIgnore)
	cmdInstanceIgnore.Flags().BoolVarP(&listOnly, "list-only", "l", false, "List only without any change")
	cmdInstanceIgnore.Flags().BoolVarP(&ignoreAll, "ignore-all", "a", false, "disable reconciliation on all SKR instances")
	cmdInstanceIgnore.Flags().BoolVarP(&ignoreNone, "ignore-none", "n", false, "enable reconciliation on all SKR instances")
	cmdInstanceIgnore.Flags().StringSliceVarP(&ignoreAliases, "ignore", "i", nil, "disable reconciliation SKR instance alias, enable all others")
	cmdInstanceIgnore.MarkFlagsMutuallyExclusive("list-only", "ignore-all", "ignore-none", "ignore")
	cmdInstanceIgnore.MarkFlagsOneRequired("list-only", "ignore-all", "ignore", "ignore-none")
}
