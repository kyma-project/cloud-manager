package main

import (
	"fmt"
	"time"

	"github.com/elliotchance/pie/v2"
	e2ekeb "github.com/kyma-project/cloud-manager/e2e/keb"
	e2elib "github.com/kyma-project/cloud-manager/e2e/lib"
	"github.com/spf13/cobra"
)

type cmdInstanceModulesInstallOptionsType struct {
	runtimeID  string
	alias      string
	moduleName string
	timeout    time.Duration
	delete     bool
}

var cmdInstanceModulesInstallOptions cmdInstanceModulesInstallOptionsType

var cmdInstanceModulesInstall = &cobra.Command{
	Use: "install",
	RunE: func(cmd *cobra.Command, args []string) error {
		keb, err := e2ekeb.Create(rootCtx, config)
		if err != nil {
			return fmt.Errorf("failed to create keb: %w", err)
		}

		if _, ok := config.Modules[cmdInstanceModulesInstallOptions.moduleName]; !ok {
			return fmt.Errorf("module %q is not defined in config.modules", cmdInstanceModulesInstallOptions.moduleName)
		}

		if cmdInstanceModulesInstallOptions.runtimeID == "" {
			idArr, err := keb.List(rootCtx, e2ekeb.WithAlias(cmdInstanceModulesInstallOptions.alias))
			if err != nil {
				return fmt.Errorf("failed to list runtimes: %w", err)
			}
			if len(idArr) == 0 {
				return fmt.Errorf("runtime with alias %q not found", cmdInstanceModulesInstallOptions.alias)
			}
			if len(idArr) > 1 {
				return fmt.Errorf("multiple runtimes with alias %q found: %v", cmdInstanceModulesInstallOptions.alias, pie.Map(idArr, func(x e2ekeb.InstanceDetails) string {
					return x.RuntimeID
				}))
			}
			cmdInstanceModulesInstallOptions.runtimeID = idArr[0].RuntimeID
		}

		clnt, err := keb.CreateInstanceClient(rootCtx, cmdInstanceModulesInstallOptions.runtimeID)
		if err != nil {
			return err
		}

		mm := e2elib.ModuleManifest(rootLogger, cmdInstanceModulesInstallOptions.moduleName)
		if err := mm.Load(); err != nil {
			return fmt.Errorf("failed to load module manifest: %w", err)
		}

		if cmdInstanceModulesInstallOptions.delete {
			if err := mm.DeleteDefaultCR(clnt, cmdInstanceModulesInstallOptions.timeout); err != nil {
				return fmt.Errorf("failed to delete default module CR: %w", err)
			}
			if err := mm.DeleteManifests(clnt, cmdInstanceModulesInstallOptions.timeout); err != nil {
				return fmt.Errorf("failed to delete module manifests: %w", err)
			}
		} else {
			if err := mm.InstallManifests(clnt, cmdInstanceModulesInstallOptions.timeout); err != nil {
				return fmt.Errorf("failed to install module manifests: %w", err)
			}
			if err := mm.InstallDefaultCR(clnt, cmdInstanceModulesInstallOptions.timeout); err != nil {
				return fmt.Errorf("failed to install default module CR: %w", err)
			}
		}

		return nil
	},
}

func init() {
	cmdInstanceModules.AddCommand(cmdInstanceModulesInstall)
	cmdInstanceModulesInstall.Flags().StringVarP(&cmdInstanceModulesInstallOptions.runtimeID, "runtime-id", "r", "", "The runtime ID")
	cmdInstanceModulesInstall.Flags().StringVarP(&cmdInstanceModulesInstallOptions.alias, "alias", "a", "", "The runtime alias")
	cmdInstanceModulesInstall.Flags().StringVarP(&cmdInstanceModulesInstallOptions.moduleName, "module", "m", "", "The module name")
	cmdInstanceModulesInstall.Flags().DurationVarP(&cmdInstanceModulesInstallOptions.timeout, "timeout", "t", 5*time.Minute, "Timeout")
	cmdInstanceModulesInstall.Flags().BoolVarP(&cmdInstanceModulesInstallOptions.delete, "delete", "d", false, "Delete the module after installation")
	_ = cmdInstanceModulesInstall.MarkFlagRequired("module")
	cmdInstanceModulesInstall.MarkFlagsMutuallyExclusive("runtime-id", "alias")
	cmdInstanceModulesInstall.MarkFlagsOneRequired("runtime-id", "alias")
}
