package sim

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/elliotchance/pie/v2"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/external/operatorshared"
	"github.com/kyma-project/cloud-manager/pkg/external/operatorv1beta2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type KymaSync struct {
	KCP *operatorv1beta2.Kyma
	SKR *operatorv1beta2.Kyma
}

type SyncOutcome struct {
	KymaKCP *operatorv1beta2.Kyma
	KymaSKR *operatorv1beta2.Kyma

	KCP KymaOutcome
	SKR KymaOutcome

	// modulesToRemove contains a difference between status and spec of the SKR Kyma - it's a list
	// of modules that exist in status, but do not exist in spec
	modulesToRemove []string

	// activeModules contains all modules in the spec of the SKR Kyma
	activeModules []string
}

type KymaOutcome struct {
	SpecChanged   bool
	StatusChanged bool
}

// AllRemovedModules returns list of modules that exist in status, but do not exist in the spec of the SKR Kyma
func (o *SyncOutcome) AllRemovedModules() []string {
	return append([]string{}, o.modulesToRemove...)
}

// IsRemoved returns true if given module is in the removed modules - exists in status, but not in spec
func (o *SyncOutcome) IsRemoved(moduleName string) bool {
	return pie.Contains(o.modulesToRemove, moduleName)
}

// IsActive returns true is given moduleName is listed in the spec of SKR Kyma
func (o *SyncOutcome) IsActive(moduleName string) bool {
	return pie.Contains(o.activeModules, moduleName) && !o.IsRemoved(moduleName)
}

func (o *SyncOutcome) SetSkrKymaReadyStatus() {
	if o.KymaSKR.Status.State != operatorshared.StateReady {
		o.KymaSKR.Status.State = operatorshared.StateReady
		o.SKR.StatusChanged = true
	}
	if len(o.KymaSKR.Status.Conditions) > 0 {
		o.KymaSKR.Status.Conditions = []metav1.Condition{}
		o.SKR.StatusChanged = true
	}
}

func (o *SyncOutcome) PatchObjects(ctx context.Context, skrClient, kcpClient client.Client) error {
	logger := composed.LoggerFromCtx(ctx)
	var result []error
	if o.SKR.SpecChanged {
		logger.Info("Updating SKR Kyma spec.modules")
		if err := skrClient.Update(ctx, o.KymaSKR); err != nil {
			result = append(result, fmt.Errorf("error updating SKR Kyma spec: %w", err))
		}
	}
	if o.SKR.StatusChanged {
		logger.Info("Patching SKR Kyma status.modules")
		if err := composed.PatchObjStatus(ctx, o.KymaSKR, skrClient); err != nil {
			result = append(result, fmt.Errorf("error patching SKR Kyma status: %w", err))
		}
	}
	if o.KCP.SpecChanged {
		logger.Info("Updating KCP Kyma spec.modules")
		if err := kcpClient.Update(ctx, o.KymaKCP); err != nil {
			result = append(result, fmt.Errorf("error updating KCP Kyma spec: %w", err))
		}
	}
	if o.KCP.StatusChanged {
		logger.Info("Patching KCP Kyma status.modules")
		if err := composed.PatchObjStatus(ctx, o.KymaKCP, kcpClient); err != nil {
			result = append(result, fmt.Errorf("error patching KCP Kyma status: %w", err))
		}
	}
	return errors.Join(result...)
}

func mapModuleNameToString(m operatorv1beta2.Module) string {
	return m.Name
}

func mapModuleStatusNameToString(m operatorv1beta2.ModuleStatus) string {
	return m.Name
}

func moduleNames(arr []operatorv1beta2.Module) []string {
	return pie.Unique(pie.Map(arr, mapModuleNameToString))
}

func moduleStatusNames(arr []operatorv1beta2.ModuleStatus) []string {
	return pie.Unique(pie.Map(arr, mapModuleStatusNameToString))
}

func (s *KymaSync) Sync() SyncOutcome {
	result := SyncOutcome{
		KymaSKR: s.SKR,
		KymaKCP: s.KCP,
	}

	skrSpec := moduleNames(s.SKR.Spec.Modules)
	skrStatus := moduleStatusNames(s.SKR.Status.Modules)
	kcpSpec := moduleNames(s.KCP.Spec.Modules)

	skrAddToStatus, skrRemoveFromStatus := pie.Diff(skrStatus, skrSpec)
	s.addRemoveStatus(s.SKR, skrAddToStatus, skrRemoveFromStatus)
	result.SKR.StatusChanged = len(skrAddToStatus) > 0 || len(skrRemoveFromStatus) > 0
	result.modulesToRemove = skrRemoveFromStatus
	result.activeModules = skrSpec

	kcpAddToSpec, kcpRemoveFromSpec := pie.Diff(kcpSpec, skrSpec)
	result.KCP.SpecChanged = len(kcpAddToSpec) > 0 || len(kcpRemoveFromSpec) > 0
	s.KCP.Spec = s.SKR.Spec

	s.KCP.Status = s.SKR.Status
	result.KCP.StatusChanged = result.SKR.StatusChanged

	return result
}

func (s *KymaSync) addRemoveStatus(k *operatorv1beta2.Kyma, add, remove []string) {
	msm := k.GetModuleStatusMap()

	for _, moduleName := range add {
		if _, exists := msm[moduleName]; !exists {
			k.Status.Modules = append(k.Status.Modules, operatorv1beta2.ModuleStatus{
				Name:    moduleName,
				Channel: k.Spec.Channel,
				State:   operatorshared.StateProcessing,
			})
		}
	}

	for _, moduleName := range remove {
		m, exists := msm[moduleName]
		if !exists {
			k.Status.Modules = append(k.Status.Modules, operatorv1beta2.ModuleStatus{
				Name:    moduleName,
				Channel: k.Spec.Channel,
				State:   operatorshared.StateProcessing,
			})
		} else {
			m.State = operatorshared.StateProcessing
		}
	}
}

func (o *SyncOutcome) AutoProcessAllBut(skipModuleNames ...string) {
	var moduleNamesToProcess []string
	for moduleName := range o.KymaSKR.GetModuleStatusMap() {
		if !slices.Contains(skipModuleNames, moduleName) {
			moduleNamesToProcess = append(moduleNamesToProcess, moduleName)
		}
	}
	for _, moduleName := range moduleNamesToProcess {
		o.Processed(moduleName, operatorshared.StateReady, "Success")
	}
}

func (o *SyncOutcome) Processed(moduleName string, state operatorshared.State, msg string) {
	changed := false
	for i, m := range o.KymaSKR.Status.Modules {
		if m.Name == moduleName {
			if m.State != state {
				o.KymaSKR.Status.Modules[i].State = state
				if len(msg) > 0 {
					o.KymaSKR.Status.Modules[i].Message = msg
				}
				o.SKR.StatusChanged = true
				changed = true
			}
			break
		}
	}
	// remove if it has to be removed and in ready status
	if o.IsRemoved(moduleName) && state == operatorshared.StateReady {
		o.KymaSKR.Status.Modules = pie.FilterNot(o.KymaSKR.Status.Modules, func(moduleStatus operatorv1beta2.ModuleStatus) bool {
			return moduleStatus.Name == moduleName
		})
		changed = true
	}
	if changed {
		o.KymaKCP.Status = o.KymaSKR.Status
		o.KCP.StatusChanged = true
	}
}
