package security

import (
	"context"
	"strings"
	"testing"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/external/infrastructuremanagerv1"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type testRuntimeState struct {
	composed.State
}

func (s *testRuntimeState) ObjAsRuntime() *infrastructuremanagerv1.Runtime {
	return s.Obj().(*infrastructuremanagerv1.Runtime)
}

func (s *testRuntimeState) Subscription() *cloudcontrolv1beta1.Subscription {
	return nil
}

func (s *testRuntimeState) VpcNetwork() *cloudcontrolv1beta1.VpcNetwork {
	return nil
}

func (s *testRuntimeState) SecurityServiceEnabledOnSubscription() bool {
	return false
}

func (s *testRuntimeState) SecurityDataSourceEnabledOnRuntime() bool {
	return false
}

func (s *testRuntimeState) SecurityServiceEnabledOnSubscriptionPredicate(_ context.Context, st composed.State) bool {
	return s.SecurityServiceEnabledOnSubscription()
}

func (s *testRuntimeState) SecurityDataSourceEnabledOnRuntimePredicate(_ context.Context, st composed.State) bool {
	return s.SecurityDataSourceEnabledOnRuntime()
}

func TestState(t *testing.T) {

	newStateWithRuntime := func(shootName string) *State {
		rt := &infrastructuremanagerv1.Runtime{
			Spec: infrastructuremanagerv1.RuntimeSpec{
				Shoot: infrastructuremanagerv1.RuntimeShoot{
					Name: shootName,
				},
			},
		}
		st := &State{
			State: &testRuntimeState{
				State: composed.NewStateFactory(nil).NewState(client.ObjectKeyFromObject(rt), rt),
			},
		}
		return st
	}

	t.Run("storageAccountBaseName", func(t *testing.T) {
		testCases := []struct {
			shoot    string
			expected string
		}{
			{"t-abd456", "kymasectabd456"},
			{"abc123", "kymasecabc123"},
			{"p-x-2-y-3-z-4", "kymasecpx2y3z4"},
			{"abcdefgh1", "kymasecabcdefgh1"},
			{"abcdefgh12", "kymasecabcdefgh12"},
			{"abcdefgh123", "kymasecabcdefgh123"},
			{"abcdefgh1234", "kymasecabcdefgh1234"},
			{"abcdefgh12345", "kymasecabcdefgh1234"},
			{"abcdefgh123456", "kymasecabcdefgh1234"},
			{"abcdefgh12345657", "kymasecabcdefgh1234"},
			{"abcdefgh12345678", "kymasecabcdefgh1234"},
			{"abcdefgh123456789", "kymasecabcdefgh1234"},
			{"abcdefgh1234567890", "kymasecabcdefgh1234"},
		}

		for _, tc := range testCases {
			t.Run(tc.shoot, func(t *testing.T) {
				st := newStateWithRuntime(tc.shoot)
				actual := st.storageAccountBaseName()
				assert.Equal(t, tc.expected, actual)
			})

		}
	})

	t.Run("storageAccountNameAttempt", func(t *testing.T) {
		testCases := []struct {
			shoot  string
			prefix string
		}{
			{"t-abd456", "kymasectabd456"},
			{"abc123", "kymasecabc123"},
			{"p-x-2-y-3-z-4", "kymasecpx2y3z4"},
			{"abcdefgh1", "kymasecabcdefgh1"},
			{"abcdefgh12", "kymasecabcdefgh12"},
			{"abcdefgh123", "kymasecabcdefgh123"},
			{"abcdefgh1234", "kymasecabcdefgh1234"},
			{"abcdefgh12345", "kymasecabcdefgh1234"},
			{"abcdefgh123456", "kymasecabcdefgh1234"},
			{"abcdefgh12345657", "kymasecabcdefgh1234"},
			{"abcdefgh12345678", "kymasecabcdefgh1234"},
			{"abcdefgh123456789", "kymasecabcdefgh1234"},
			{"abcdefgh1234567890", "kymasecabcdefgh1234"},
		}

		for _, tc := range testCases {
			t.Run(tc.shoot, func(t *testing.T) {
				st := newStateWithRuntime(tc.shoot)

				// first attempt doesn't have random suffix and equals to base name
				actual := st.storageAccountNameAttempt(0)
				assert.True(t, strings.HasPrefix(actual, tc.prefix))
				assert.LessOrEqual(t, len(actual), maxStorageAccountNameLength-5)
				assert.Equal(t, st.storageAccountBaseName(), actual)

				// second attempt has random suffix, and base name is the prefix
				actual = st.storageAccountNameAttempt(1)
				assert.True(t, strings.HasPrefix(actual, tc.prefix))
				assert.LessOrEqual(t, len(actual), maxStorageAccountNameLength)
			})

		}
	})

}
