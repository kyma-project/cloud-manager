package sim

import (
	"github.com/elliotchance/pie/v2"
	"github.com/kyma-project/cloud-manager/pkg/external/operatorshared"
	"github.com/kyma-project/cloud-manager/pkg/external/operatorv1beta2"
)

type KymaSync struct {
	KCP *operatorv1beta2.Kyma
	SKR *operatorv1beta2.Kyma
}

type SyncOutcome struct {
	KCP KymaOutcome
	SKR KymaOutcome

	modulesToRemove []string
	activeModules   []string
}

type KymaOutcome struct {
	SpecChanged   bool
	StatusChanged bool
}

func (o *SyncOutcome) AllRemovedModules() []string {
	return append([]string{}, o.modulesToRemove...)
}

func (o *SyncOutcome) IsRemoved(moduleName string) bool {
	return pie.Contains(o.modulesToRemove, moduleName)
}

func (o *SyncOutcome) IsActive(moduleName string) bool {
	return pie.Contains(o.activeModules, moduleName) && !o.IsRemoved(moduleName)
}

func mapModuleNameToString(m operatorv1beta2.Module) string {
	return m.Name
}

func mapModuleStatusNameToString(m operatorv1beta2.ModuleStatus) string {
	return m.Name
}

func moduleNames(k *operatorv1beta2.Kyma) []string {
	return pie.Unique(pie.Map(k.Spec.Modules, mapModuleNameToString))
}

func moduleStatusNames(k *operatorv1beta2.Kyma) []string {
	return pie.Unique(pie.Map(k.Status.Modules, mapModuleStatusNameToString))
}

func (s KymaSync) Sync() SyncOutcome {
	result := SyncOutcome{}
	skrSpec := moduleNames(s.SKR)
	skrStatus := moduleStatusNames(s.SKR)
	kcpSpec := moduleNames(s.KCP)
	kcpStatus := moduleStatusNames(s.KCP)

	skrAddToStatus, skrRemoveFromStatus := pie.Diff(skrStatus, skrSpec)
	s.addRemoveStatus(s.SKR, skrAddToStatus, skrRemoveFromStatus)
	result.SKR.StatusChanged = len(skrAddToStatus) > 0 || len(skrRemoveFromStatus) > 0
	result.modulesToRemove = skrRemoveFromStatus
	result.activeModules = skrSpec

	kcpAddToSpec, kcpRemoveFromSpec := pie.Diff(kcpSpec, skrSpec)
	s.addRemoveSpec(s.KCP, kcpAddToSpec, kcpRemoveFromSpec)
	result.KCP.SpecChanged = len(kcpAddToSpec) > 0 || len(kcpRemoveFromSpec) > 0

	kcpAddToStatus, kcpRemoveFromStatus := pie.Diff(kcpStatus, skrSpec)
	s.addRemoveStatus(s.KCP, kcpAddToStatus, kcpRemoveFromStatus)
	result.KCP.StatusChanged = len(kcpAddToStatus) > 0 || len(kcpRemoveFromStatus) > 0

	return result
}

func (s *KymaSync) addRemoveStatus(k *operatorv1beta2.Kyma, add, remove []string) {
	for _, moduleName := range add {
		k.Status.Modules = append(k.Status.Modules, operatorv1beta2.ModuleStatus{
			Name:    moduleName,
			Channel: k.Spec.Channel,
			State:   operatorshared.StateReady,
		})
	}
	k.Status.Modules = pie.FilterNot(k.Status.Modules, func(m operatorv1beta2.ModuleStatus) bool {
		return pie.Contains(remove, m.Name)
	})
}

func (s *KymaSync) addRemoveSpec(k *operatorv1beta2.Kyma, add, remove []string) {
	for _, moduleName := range add {
		k.Spec.Modules = append(k.Spec.Modules, operatorv1beta2.Module{
			Name:    moduleName,
			Channel: k.Spec.Channel,
		})
	}
	k.Spec.Modules = pie.FilterNot(k.Spec.Modules, func(m operatorv1beta2.Module) bool {
		return pie.Contains(remove, m.Name)
	})
}
