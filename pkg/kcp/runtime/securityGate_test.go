package runtime

import (
	"context"
	"testing"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	runtimetypes "github.com/kyma-project/cloud-manager/pkg/kcp/runtime/types"
	"github.com/stretchr/testify/assert"
)

type securityGateGivenTestCase struct {
	title                  string
	actualSubscriptionsOn  []string
	actualSubscriptionsOff []string
	actualRuntimesOn       []string
	actualRuntimesOff      []string
	currentSubscriptionId  string
	currentRuntimeId       string
	desiredSubscription    bool
	desiredRuntime         bool
	expected               bool
}

type testSecurityGateState struct {
	composed.State
	ds runtimetypes.SecurityDesiredState
}

var _ composed.State = (*testSecurityGateState)(nil)

func (s *testSecurityGateState) SecurityDesiredState() runtimetypes.SecurityDesiredState {
	return s.ds
}

func TestSecurityGate(t *testing.T) {

	// shouldRun test cases =============================================================

	t.Run("shouldRun", func(t *testing.T) {
		const sub1 = "sub1"
		const rt1 = "rt1"
		const rt2 = "rt2"

		testCases := []securityGateGivenTestCase{
			// given cold state
			{
				title:               "given cold state when true|true then true",
				desiredSubscription: true,
				desiredRuntime:      true,
				expected:            true,
			},
			{
				title:               "given cold state when true|false then true",
				desiredSubscription: true,
				desiredRuntime:      false,
				expected:            true,
			},
			{
				title:               "given cold state when false|true then true",
				desiredSubscription: false,
				desiredRuntime:      true,
				expected:            true,
			},
			{
				title:               "given cold state when false|false then true",
				desiredSubscription: false,
				desiredRuntime:      false,
				expected:            true,
			},
			// given last run true|true
			{
				title:                 "given last run true|true when true|true then false",
				actualSubscriptionsOn: []string{sub1},
				actualRuntimesOn:      []string{rt1},
				desiredSubscription:   true,
				desiredRuntime:        true,
				expected:              false,
			},
			{
				title:                 "given last run true|true when true|false then true",
				actualSubscriptionsOn: []string{sub1},
				actualRuntimesOn:      []string{rt1},
				desiredSubscription:   true,
				desiredRuntime:        false,
				expected:              true,
			},
			{
				title:                 "given last run true|true when false|true then true",
				actualSubscriptionsOn: []string{sub1},
				actualRuntimesOn:      []string{rt1},
				desiredSubscription:   false,
				desiredRuntime:        true,
				expected:              true,
			},
			{
				title:                 "given last run true|true when false|false then true",
				actualSubscriptionsOn: []string{sub1},
				actualRuntimesOn:      []string{rt1},
				desiredSubscription:   false,
				desiredRuntime:        false,
				expected:              true,
			},
			// given last run false|false
			{
				title:                  "given last run false|false when true|true then true",
				actualSubscriptionsOff: []string{sub1},
				actualRuntimesOff:      []string{rt1},
				desiredSubscription:    true,
				desiredRuntime:         true,
				expected:               true,
			},
			{
				title:                  "given last run false|false when true|false then true",
				actualSubscriptionsOff: []string{sub1},
				actualRuntimesOff:      []string{rt1},
				desiredSubscription:    true,
				desiredRuntime:         false,
				expected:               true,
			},
			{
				title:                  "given last run false|false when false|true then true",
				actualSubscriptionsOff: []string{sub1},
				actualRuntimesOff:      []string{rt1},
				desiredSubscription:    false,
				desiredRuntime:         true,
				expected:               true,
			},
			{
				title:                  "given last run false|false when false|false then false",
				actualSubscriptionsOff: []string{sub1},
				actualRuntimesOff:      []string{rt1},
				desiredSubscription:    false,
				desiredRuntime:         false,
				expected:               false,
			},
			// given sub1 false and two runtimes false|false
			{
				title:                  "given sub1 false and two runtimes false|false when rt1 true|true then true",
				actualSubscriptionsOff: []string{sub1},
				actualRuntimesOff:      []string{rt1, rt2},
				currentRuntimeId:       rt1,
				desiredSubscription:    true,
				desiredRuntime:         true,
				expected:               true,
			},
			{
				title:                  "given sub1 false and two runtimes false|false when rt2 true|true then true",
				actualSubscriptionsOff: []string{sub1},
				actualRuntimesOff:      []string{rt1, rt2},
				currentRuntimeId:       rt2,
				desiredSubscription:    true,
				desiredRuntime:         true,
				expected:               true,
			},
		}

		for _, testCase := range testCases {
			t.Run(testCase.title, func(t *testing.T) {
				gate := &securityGate{
					actualStateRuntimes:      map[string]bool{},
					actualStateSubscriptions: map[string]bool{},
				}
				for _, sub := range testCase.actualSubscriptionsOn {
					gate.actualStateSubscriptions[sub] = true
				}
				for _, sub := range testCase.actualSubscriptionsOff {
					gate.actualStateSubscriptions[sub] = false
				}
				for _, rt := range testCase.actualRuntimesOn {
					gate.actualStateRuntimes[rt] = true
				}
				for _, rt := range testCase.actualRuntimesOff {
					gate.actualStateRuntimes[rt] = false
				}
				subscriptionId := testCase.currentSubscriptionId
				if subscriptionId == "" {
					subscriptionId = sub1
				}
				runtimeId := testCase.currentRuntimeId
				if runtimeId == "" {
					runtimeId = rt1
				}
				ds := &securityDesiredState{
					runtimeId:             runtimeId,
					subscriptionId:        subscriptionId,
					enabledOnRuntime:      testCase.desiredRuntime,
					enabledOnSubscription: testCase.desiredSubscription,
				}
				actual := gate.shouldRun(ds)
				assert.Equal(t, testCase.expected, actual)

				// check the predicate
				st := &testSecurityGateState{
					ds: ds,
				}
				actualFromPredicate := gate.ShouldRunPredicate(context.TODO(), st)
				assert.Equal(t, actual, actualFromPredicate, "predicate should return same value as shouldRun()")

				// when marked as success then it should NOT run
				gate.markSuccess(ds)
				actual = gate.shouldRun(ds)
				assert.False(t, actual, "second run with same desired state after markSuccess must result with false")
			})
		}
	})

}
