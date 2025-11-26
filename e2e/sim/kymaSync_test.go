package sim

import (
	"testing"

	"github.com/elliotchance/pie/v2"
	"github.com/kyma-project/cloud-manager/pkg/external/operatorshared"
	"github.com/kyma-project/cloud-manager/pkg/external/operatorv1beta2"
	"github.com/stretchr/testify/assert"
)

func Test_KymaSync(t *testing.T) {
	testData := []struct {
		title            string
		skrSpec          []string
		skrStatus        []string
		kcpSpec          []string
		kcpStatus        []string
		changedSkrStatus bool
		changedKcpSpec   bool
		changedKcpStatus bool
		removedModules   []string
	}{
		{
			title:            "no sync when empty",
			skrSpec:          nil,
			skrStatus:        nil,
			kcpSpec:          nil,
			kcpStatus:        nil,
			changedSkrStatus: false,
			changedKcpSpec:   false,
			changedKcpStatus: false,
		},
		{
			title:            "first module added",
			skrSpec:          []string{"aaa"},
			skrStatus:        nil,
			kcpSpec:          nil,
			kcpStatus:        nil,
			changedSkrStatus: true,
			changedKcpSpec:   true,
			changedKcpStatus: true,
		},
		{
			title:            "second module added",
			skrSpec:          []string{"aaa", "bbb"},
			skrStatus:        []string{"aaa"},
			kcpSpec:          []string{"aaa"},
			kcpStatus:        []string{"aaa"},
			changedSkrStatus: true,
			changedKcpSpec:   true,
			changedKcpStatus: true,
		},
		{
			title:            "second module removed",
			skrSpec:          []string{"aaa"},
			skrStatus:        []string{"bbb", "aaa"},
			kcpSpec:          []string{"bbb", "aaa"},
			kcpStatus:        []string{"bbb", "aaa"},
			changedSkrStatus: true,
			changedKcpSpec:   true,
			changedKcpStatus: true,
			removedModules:   []string{"bbb"},
		},
		{
			title:            "last module removed",
			skrSpec:          nil,
			skrStatus:        []string{"aaa"},
			kcpSpec:          []string{"aaa"},
			kcpStatus:        []string{"aaa"},
			changedSkrStatus: true,
			changedKcpSpec:   true,
			changedKcpStatus: true,
			removedModules:   []string{"aaa"},
		},
	}

	for _, data := range testData {
		t.Run(data.title, func(t *testing.T) {
			skr := &operatorv1beta2.Kyma{
				Spec: operatorv1beta2.KymaSpec{
					Channel: operatorv1beta2.DefaultChannel,
					Modules: pie.Map(data.skrSpec, func(moduleName string) operatorv1beta2.Module {
						return operatorv1beta2.Module{
							Name:    moduleName,
							Channel: operatorv1beta2.DefaultChannel,
						}
					}),
				},
				Status: operatorv1beta2.KymaStatus{
					Modules: pie.Map(data.skrStatus, func(moduleName string) operatorv1beta2.ModuleStatus {
						return operatorv1beta2.ModuleStatus{
							Name:  moduleName,
							State: operatorshared.StateReady,
						}
					}),
				},
			}

			kcp := &operatorv1beta2.Kyma{
				Spec: operatorv1beta2.KymaSpec{
					Channel: operatorv1beta2.DefaultChannel,
					Modules: pie.Map(data.kcpSpec, func(moduleName string) operatorv1beta2.Module {
						return operatorv1beta2.Module{
							Name:    moduleName,
							Channel: operatorv1beta2.DefaultChannel,
						}
					}),
				},
				Status: operatorv1beta2.KymaStatus{
					Modules: pie.Map(data.kcpStatus, func(moduleName string) operatorv1beta2.ModuleStatus {
						return operatorv1beta2.ModuleStatus{
							Name:  moduleName,
							State: operatorshared.StateReady,
						}
					}),
				},
			}

			outcome := (KymaSync{SKR: skr, KCP: kcp}).Sync()

			skrSpecModules := moduleNames(skr)
			skrStatusModules := moduleStatusNames(skr)
			kcpSpecModules := moduleNames(kcp)
			kcpStatusModules := moduleStatusNames(kcp)

			assert.ElementsMatchf(t, skrSpecModules, skrStatusModules, "SKR spec vs SKR status diff")
			assert.ElementsMatchf(t, skrSpecModules, kcpSpecModules, "SKR spec vs KCP spec diff")
			assert.ElementsMatchf(t, skrSpecModules, kcpStatusModules, "SKR spec vs KCP status diff")
			assert.False(t, outcome.SKR.SpecChanged, "SKR spec changed is supposed always to be false")
			assert.Equal(t, data.changedSkrStatus, outcome.SKR.StatusChanged, "SKR status changed")
			assert.Equal(t, data.changedKcpSpec, outcome.KCP.SpecChanged, "KCP spec changed")
			assert.Equal(t, data.changedKcpStatus, outcome.KCP.StatusChanged, "KCP status changed")

			for _, moduleName := range data.removedModules {
				assert.True(t, outcome.IsRemoved(moduleName))
			}
			for _, moduleName := range data.skrSpec {
				assert.True(t, outcome.IsActive(moduleName))
			}
		})
	}
}
