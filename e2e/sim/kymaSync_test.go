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
			kcpNotInSkrSpec:  nil,
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
			kcpNotInSkrSpec: nil,
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
			kcpNotInSkrSpec: nil,
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
			kcpNotInSkrSpec: nil,
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
			kcpNotInSkrSpec: nil,
		},
		{
			// Mid-removal processing: bbb is being removed (in status, not in spec). KCP status
			// mirrors SKR status and still lists bbb=Processing. Divergence is expected here —
			// the gate correctly requeues while the genuine removal is in progress.
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
			kcpNotInSkrSpec: []string{"bbb"},
		},
		{
			// Resolved removal: bbb processed→Ready causes it to be stripped from both statuses.
			// KCP status ends clean ([aaa]). No divergence → no requeue.
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
			kcpNotInSkrSpec: nil,
		},
		{
			// Stale-read shape: SKR spec no longer lists the module (removed) but a stale cache
			// read still shows it in SKR status. Once resolved (Processed Ready), the module is
			// gone from both statuses. KCP status ends empty → no divergence → no requeue.
			// This is the exact shape that stranded KCP in the flake, now confirmed to converge.
			title:   "sole module removed from spec, resolved and dropped from both statuses",
			skrSpec: nil,
			skrStatus: []operatorv1beta2.ModuleStatus{
				{Name: "cloud-manager", State: operatorshared.StateProcessing},
			},
			kcpSpec: nil,
			kcpStatus: []operatorv1beta2.ModuleStatus{
				{Name: "cloud-manager", State: operatorshared.StateProcessing},
			},
			processed:        map[string]operatorshared.State{"cloud-manager": operatorshared.StateReady},
			changedSkrStatus: true,
			changedKcpSpec:   false,
			changedKcpStatus: true,
			removedModules:   []string{"cloud-manager"},
			expectedStatus:   nil,
			kcpNotInSkrSpec:  nil,
		},
		{
			// Stale re-mirror, unresolved: SKR spec is empty (module removed) but the stale
			// cache still shows cloud-manager=Processing in SKR status, which gets mirrored
			// onto KCP status. No Processed() call this cycle → module stays in KCP status.
			// Divergence is non-empty → the gate requeues. This is the exact strand shape;
			// asserts the gate would prevent KCP from being permanently stranded.
			title:   "stale re-mirror unresolved: module absent from spec but present in KCP status",
			skrSpec: nil,
			skrStatus: []operatorv1beta2.ModuleStatus{
				{Name: "cloud-manager", State: operatorshared.StateProcessing},
			},
			kcpSpec: nil,
			kcpStatus: []operatorv1beta2.ModuleStatus{
				{Name: "cloud-manager", State: operatorshared.StateProcessing},
			},
			processed:        nil,
			changedSkrStatus: true,
			changedKcpSpec:   false,
			changedKcpStatus: true,
			removedModules:   []string{"cloud-manager"},
			expectedStatus: []operatorv1beta2.ModuleStatus{
				{Name: "cloud-manager", State: operatorshared.StateProcessing},
			},
			kcpNotInSkrSpec: []string{"cloud-manager"},
		},
		{
			// Steady-state active module: spec and status both list aaa=Ready. No divergence.
			title:   "steady-state active module: no divergence",
			skrSpec: []string{"aaa"},
			skrStatus: []operatorv1beta2.ModuleStatus{
				{Name: "aaa", State: operatorshared.StateReady},
			},
			kcpSpec: []string{"aaa"},
			kcpStatus: []operatorv1beta2.ModuleStatus{
				{Name: "aaa", State: operatorshared.StateReady},
			},
			processed:        nil,
			changedSkrStatus: false,
			changedKcpSpec:   false,
			changedKcpStatus: false,
			removedModules:   nil,
			expectedStatus: []operatorv1beta2.ModuleStatus{
				{Name: "aaa", State: operatorshared.StateReady},
			},
			kcpNotInSkrSpec: nil,
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

			// check KcpModulesNotInSkrSpec (convergence gate)
			Expect(outcome.KcpModulesNotInSkrSpec()).To(ConsistOf(util.ToAnySlice(tc.kcpNotInSkrSpec)...),
				"KCP-status-not-in-SKR-spec divergence set")

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
	kcpNotInSkrSpec  []string
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
