package sim

import (
	"fmt"

	"github.com/elliotchance/pie/v2"
	"github.com/kyma-project/cloud-manager/pkg/external/operatorshared"
	"github.com/kyma-project/cloud-manager/pkg/external/operatorv1beta2"
	"github.com/kyma-project/cloud-manager/pkg/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Feature: SKR/KCP Kyma sync and outcome", func() {

	testData := []syncTestCase{
		{
			title:            "no sync when empty",
			skrSpec:          nil,
			skrStatus:        nil,
			kcpSpec:          nil,
			kcpStatus:        nil,
			processed:        nil,
			changedSkrStatus: false,
			changedKcpSpec:   false,
			changedKcpStatus: false,
			removedModules:   nil,
			expectedStatus:   nil,
		},
		{
			title:            "first module added and processing",
			skrSpec:          []string{"aaa"},
			skrStatus:        nil,
			kcpSpec:          nil,
			kcpStatus:        nil,
			processed:        nil,
			changedSkrStatus: true,
			changedKcpSpec:   true,
			changedKcpStatus: true,
			removedModules:   nil,
			expectedStatus: []operatorv1beta2.ModuleStatus{
				{Name: "aaa", State: operatorshared.StateProcessing},
			},
		},
		{
			title:   "first module added and processed",
			skrSpec: []string{"aaa"},
			skrStatus: []operatorv1beta2.ModuleStatus{
				{Name: "aaa", State: operatorshared.StateProcessing},
			},
			kcpSpec: []string{"aaa"},
			kcpStatus: []operatorv1beta2.ModuleStatus{
				{Name: "aaa", State: operatorshared.StateProcessing},
			},
			processed:        map[string]operatorshared.State{"aaa": operatorshared.StateReady},
			changedSkrStatus: true,
			changedKcpSpec:   false,
			changedKcpStatus: true,
			removedModules:   nil,
			expectedStatus: []operatorv1beta2.ModuleStatus{
				{Name: "aaa", State: operatorshared.StateReady},
			},
		},
		{
			title:   "first module processed, second module added and processing",
			skrSpec: []string{"aaa", "bbb"},
			skrStatus: []operatorv1beta2.ModuleStatus{
				{Name: "aaa", State: operatorshared.StateReady},
			},
			kcpSpec: []string{"aaa"},
			kcpStatus: []operatorv1beta2.ModuleStatus{
				{Name: "aaa", State: operatorshared.StateReady},
			},
			processed:        nil,
			changedSkrStatus: true,
			changedKcpSpec:   true,
			changedKcpStatus: true,
			removedModules:   nil,
			expectedStatus: []operatorv1beta2.ModuleStatus{
				{Name: "aaa", State: operatorshared.StateReady},
				{Name: "bbb", State: operatorshared.StateProcessing},
			},
		},
		{
			title:   "first module processed, second module added and processed",
			skrSpec: []string{"aaa", "bbb"},
			skrStatus: []operatorv1beta2.ModuleStatus{
				{Name: "aaa", State: operatorshared.StateReady},
				{Name: "bbb", State: operatorshared.StateProcessing},
			},
			kcpSpec: []string{"aaa", "bbb"},
			kcpStatus: []operatorv1beta2.ModuleStatus{
				{Name: "aaa", State: operatorshared.StateReady},
				{Name: "bbb", State: operatorshared.StateProcessing},
			},
			processed:        map[string]operatorshared.State{"bbb": operatorshared.StateReady},
			changedSkrStatus: true,
			changedKcpSpec:   false,
			changedKcpStatus: true,
			removedModules:   nil,
			expectedStatus: []operatorv1beta2.ModuleStatus{
				{Name: "aaa", State: operatorshared.StateReady},
				{Name: "bbb", State: operatorshared.StateReady},
			},
		},
		{
			title:   "first module is ready, remove second ready module and processing",
			skrSpec: []string{"aaa"},
			skrStatus: []operatorv1beta2.ModuleStatus{
				{Name: "aaa", State: operatorshared.StateReady},
				{Name: "bbb", State: operatorshared.StateReady},
			},
			kcpSpec: []string{"aaa", "bbb"},
			kcpStatus: []operatorv1beta2.ModuleStatus{
				{Name: "aaa", State: operatorshared.StateReady},
				{Name: "bbb", State: operatorshared.StateReady},
			},
			processed:        nil,
			changedSkrStatus: true,
			changedKcpSpec:   true,
			changedKcpStatus: true,
			removedModules:   []string{"bbb"},
			expectedStatus: []operatorv1beta2.ModuleStatus{
				{Name: "aaa", State: operatorshared.StateReady},
				{Name: "bbb", State: operatorshared.StateProcessing},
			},
		},
		{
			title:   "first module is ready, removed second module and processed",
			skrSpec: []string{"aaa"},
			skrStatus: []operatorv1beta2.ModuleStatus{
				{Name: "aaa", State: operatorshared.StateReady},
				{Name: "bbb", State: operatorshared.StateProcessing},
			},
			kcpSpec: []string{"aaa"},
			kcpStatus: []operatorv1beta2.ModuleStatus{
				{Name: "aaa", State: operatorshared.StateReady},
				{Name: "bbb", State: operatorshared.StateProcessing},
			},
			processed:        map[string]operatorshared.State{"bbb": operatorshared.StateReady},
			changedSkrStatus: true,
			changedKcpSpec:   false,
			changedKcpStatus: true,
			removedModules:   []string{"bbb"},
			expectedStatus: []operatorv1beta2.ModuleStatus{
				{Name: "aaa", State: operatorshared.StateReady},
			},
		},
	}

	for _, tc := range testData {

		It(tc.title, func() {

			skr, kcp, outcome := tc.products()

			// checks

			Expect(outcome.SKR.SpecChanged).To(BeFalse(), "SKR spec changed is supposed always to be false")
			Expect(outcome.SKR.StatusChanged).To(Equal(tc.changedSkrStatus), "SKR status changed")
			Expect(outcome.KCP.SpecChanged).To(Equal(tc.changedKcpSpec), "KCP spec changed")
			Expect(outcome.KCP.StatusChanged).To(Equal(tc.changedKcpStatus), "KCP status changed")

			skrSpecModules := moduleNames(skr.Spec.Modules)
			kcpSpecModules := moduleNames(kcp.Spec.Modules)

			// KCP
			Expect(kcpSpecModules).To(ConsistOf(util.ToAnySlice(skrSpecModules)...), "KCP spec should equal to SKR spec")

			// SKR status
			skrMsm := skr.GetModuleStatusMap()
			for _, expectedModule := range tc.expectedStatus {
				actualModule, exists := skrMsm[expectedModule.Name]
				Expect(exists).To(BeTrue())
				Expect(actualModule.State).To(Equal(expectedModule.State), fmt.Sprintf("expected SKR module %s to be in state %s, but it is in %s", expectedModule.Name, expectedModule.State, actualModule.State))
			}
			for actualModuleName, actualModule := range skrMsm {
				isExpected := false
				for _, expectedModule := range tc.expectedStatus {
					if actualModuleName == expectedModule.Name {
						isExpected = true
						break
					}
				}
				Expect(isExpected).To(BeTrue(), fmt.Sprintf("unexpected module %s with state %s in SKR status", actualModuleName, actualModule.State))
			}

			// KCP status
			kcpMsm := kcp.GetModuleStatusMap()
			for _, expectedModule := range tc.expectedStatus {
				actualModule, exists := kcpMsm[expectedModule.Name]
				Expect(exists).To(BeTrue())
				Expect(actualModule.State).To(Equal(expectedModule.State), fmt.Sprintf("expected KCP module %s to be in state %s, but it is in %s", expectedModule.Name, expectedModule.State, actualModule.State))
			}
			for actualModuleName, actualModule := range kcpMsm {
				isExpected := false
				for _, expectedModule := range tc.expectedStatus {
					if actualModuleName == expectedModule.Name {
						isExpected = true
						break
					}
				}
				Expect(isExpected).To(BeTrue(), fmt.Sprintf("unexpected module %s with state %s in KCP status", actualModuleName, actualModule.State))
			}

			// check IsRemoved
			for _, moduleName := range tc.removedModules {
				Expect(outcome.IsRemoved(moduleName)).To(BeTrue(), fmt.Sprintf("Module %s should be removed", moduleName))
			}
			// check IsActive
			for _, moduleName := range tc.skrSpec {
				Expect(outcome.IsActive(moduleName)).To(BeTrue(), fmt.Sprintf("Module %s should be active", moduleName))
			}

		})

	}
})

type syncTestCase struct {
	title            string
	skrSpec          []string
	skrStatus        []operatorv1beta2.ModuleStatus
	kcpSpec          []string
	kcpStatus        []operatorv1beta2.ModuleStatus
	processed        map[string]operatorshared.State
	changedSkrStatus bool
	changedKcpSpec   bool
	changedKcpStatus bool
	removedModules   []string
	expectedStatus   []operatorv1beta2.ModuleStatus
}

func (tc syncTestCase) skrKyma() *operatorv1beta2.Kyma {
	return &operatorv1beta2.Kyma{
		Spec: operatorv1beta2.KymaSpec{
			Channel: operatorv1beta2.DefaultChannel,
			Modules: pie.Map(tc.skrSpec, func(moduleName string) operatorv1beta2.Module {
				return operatorv1beta2.Module{
					Name:    moduleName,
					Channel: operatorv1beta2.DefaultChannel,
				}
			}),
		},
		Status: operatorv1beta2.KymaStatus{
			Modules: tc.skrStatus,
		},
	}
}

func (tc syncTestCase) kcpKyma() *operatorv1beta2.Kyma {
	return &operatorv1beta2.Kyma{
		Spec: operatorv1beta2.KymaSpec{
			Channel: operatorv1beta2.DefaultChannel,
			Modules: pie.Map(tc.kcpSpec, func(moduleName string) operatorv1beta2.Module {
				return operatorv1beta2.Module{
					Name:    moduleName,
					Channel: operatorv1beta2.DefaultChannel,
				}
			}),
		},
		Status: operatorv1beta2.KymaStatus{
			Modules: tc.kcpStatus,
		},
	}
}

func (tc syncTestCase) products() (*operatorv1beta2.Kyma, *operatorv1beta2.Kyma, SyncOutcome) {
	skr := tc.skrKyma()

	kcp := tc.kcpKyma()

	outcome := (&KymaSync{SKR: skr, KCP: kcp}).Sync()

	for moduleName, state := range tc.processed {
		outcome.Processed(moduleName, state, "")
	}

	return skr, kcp, outcome
}
