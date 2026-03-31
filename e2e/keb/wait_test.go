package keb

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/kyma-project/cloud-manager/pkg/external/infrastructuremanagerv1"
	"github.com/stretchr/testify/assert"
)

func TestWait(t *testing.T) {

	t.Run("wait", func(t *testing.T) {

		testCases := []struct {
			title         string
			changes       []IdChange
			errMsg        string
			listCallCount int
		}{
			{
				"pending...completed",
				[]IdChange{
					NewIdChange(0, ChState(infrastructuremanagerv1.RuntimeStatePending)),
					NewIdChange(10, ChState(infrastructuremanagerv1.RuntimeStateReady), ChProvisioned(true)),
				},
				"",
				10,
			},
			{
				"no_status...pending....completed",
				[]IdChange{
					NewIdChange(10, ChState(infrastructuremanagerv1.RuntimeStatePending)),
					NewIdChange(20, ChState(infrastructuremanagerv1.RuntimeStateReady), ChProvisioned(true)),
				},
				"",
				20,
			},
			{
				"temporary error",
				[]IdChange{
					NewIdChange(0, ChState(infrastructuremanagerv1.RuntimeStatePending)),
					NewIdChange(10, ChState(infrastructuremanagerv1.RuntimeStateFailed), ChMessage("some error")),
					NewIdChange(12, ChState(infrastructuremanagerv1.RuntimeStatePending)),
					NewIdChange(30, ChState(infrastructuremanagerv1.RuntimeStateReady), ChProvisioned(true)),
				},
				"",
				30,
			},
			{
				"persistent error",
				[]IdChange{
					NewIdChange(0, ChState(infrastructuremanagerv1.RuntimeStatePending)),
					NewIdChange(10, ChState(infrastructuremanagerv1.RuntimeStateFailed), ChMessage("some error")),
					// never called
					//newIdChange(30, chState(infrastructuremanagerv1.RuntimeStateReady), chProvisioned(true)),
				},
				`instance alias runtime-id has error "some error"`,
				15,
			},
			{
				"long error",
				[]IdChange{
					NewIdChange(0, ChState(infrastructuremanagerv1.RuntimeStatePending)),
					NewIdChange(10, ChState(infrastructuremanagerv1.RuntimeStateFailed), ChMessage("some error")),
					// never called
					NewIdChange(30, ChState(infrastructuremanagerv1.RuntimeStateReady), ChProvisioned(true)),
				},
				`instance alias runtime-id has error "some error"`,
				15,
			},

			// delete ============================
			{
				"delete...pending....gone",
				[]IdChange{
					NewIdChange(0, ChBeingDeleted(true), ChState(infrastructuremanagerv1.RuntimeStatePending)),
					NewIdChange(10, ChRemove(true)),
				},
				"",
				10,
			},
			{
				"delete...pending...tmp_error...pending....gone",
				[]IdChange{
					NewIdChange(0, ChBeingDeleted(true), ChState(infrastructuremanagerv1.RuntimeStatePending)),
					NewIdChange(10, ChState(infrastructuremanagerv1.RuntimeStateFailed), ChMessage("some error")),
					NewIdChange(13, ChState(infrastructuremanagerv1.RuntimeStatePending), ChMessage("")),
					NewIdChange(20, ChRemove(true)),
				},
				"",
				20,
			},
			{
				"delete...pending...long_error",
				[]IdChange{
					NewIdChange(0, ChBeingDeleted(true), ChState(infrastructuremanagerv1.RuntimeStatePending)),
					NewIdChange(10, ChState(infrastructuremanagerv1.RuntimeStateFailed), ChMessage("some error")),
					// never called
					NewIdChange(30, ChState(infrastructuremanagerv1.RuntimeStatePending), ChMessage("")),
				},
				`instance alias runtime-id has error "some error"`,
				15,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.title, func(t *testing.T) {
				t.Parallel()
				id, lister := setupInstanceListerMock()
				opts := []WaitOption{
					// default timeout is ok, since lister should check the ctx, and here we use mock that ignores it
					//WithTimeout(15 * time.Minute),
					// default interval is ok, since sleeper handles it, and here we use noop
					//WithInterval(10 * time.Millisecond),
					WithRuntime(id.RuntimeID),
					WithErrorCountThreshold(5), // important to calculate listCallCount
					WithSleeperFunc(func(_ context.Context, _ time.Duration) {}),
				}

				actualListCallCount := 0

				lister.BeforeListCalled(func(i int) error {
					actualListCallCount++
					if i > 100 {
						// IMPORTANT! This is the only thing that prevents test for running too long or in a dead-loop
						return fmt.Errorf("too many calls")
					}
					for _, ch := range tc.changes {
						if ch.callCount == i {
							id.Change(ch)
						}
					}
					return nil
				})

				err := WaitCompleted(context.Background(), lister, opts...)
				if tc.errMsg != "" {
					assert.ErrorContains(t, err, tc.errMsg)
				} else {
					assert.NoError(t, err)
				}

				if tc.listCallCount > 0 {
					assert.Equal(t, tc.listCallCount, actualListCallCount)
				}
			})
		}

	})
}
