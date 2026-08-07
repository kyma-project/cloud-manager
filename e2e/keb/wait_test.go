package keb

import (
	"context"
	"fmt"
	"testing"
	"time"

	gardenertypes "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/external/infrastructuremanagerv1"
	"github.com/stretchr/testify/assert"
)

func TestWait(t *testing.T) {

	t.Run("wait", func(t *testing.T) {

		// fakeNow returns a now function whose time advances by step on every call
		// after the given call number. Used to simulate time passing during error polling.
		fakeNowAfter := func(advanceAfterCall int, step time.Duration) (func() time.Time, *int) {
			callCount := 0
			base := time.Now()
			return func() time.Time {
				callCount++
				if callCount > advanceAfterCall {
					base = base.Add(step)
				}
				return base
			}, &callCount
		}

		// fixedNow returns a now function that never advances (for transient-but-recovering cases).
		fixedNow := func() func() time.Time {
			t := time.Now()
			return func() time.Time { return t }
		}

		errorDuration := 5 * time.Second

		testCases := []struct {
			title         string
			changes       []IdChange
			errMsg        string
			listCallCount int
			nowFunc       func() time.Time
		}{
			{
				title: "pending...completed",
				changes: []IdChange{
					NewIdChange(0, ChState(infrastructuremanagerv1.RuntimeStatePending)),
					NewIdChange(10, ChState(infrastructuremanagerv1.RuntimeStateReady), ChProvisioned(true)),
				},
				errMsg:        "",
				listCallCount: 10,
				nowFunc:       fixedNow(),
			},
			{
				title: "no_status...pending....completed",
				changes: []IdChange{
					NewIdChange(10, ChState(infrastructuremanagerv1.RuntimeStatePending)),
					NewIdChange(20, ChState(infrastructuremanagerv1.RuntimeStateReady), ChProvisioned(true)),
				},
				errMsg:        "",
				listCallCount: 20,
				nowFunc:       fixedNow(),
			},
			{
				// transient error: clock doesn't advance → never exceeds errorDuration → recovers
				title: "temporary error",
				changes: []IdChange{
					NewIdChange(0, ChState(infrastructuremanagerv1.RuntimeStatePending)),
					NewIdChange(10, ChState(infrastructuremanagerv1.RuntimeStateFailed), ChMessage("some error")),
					NewIdChange(12, ChState(infrastructuremanagerv1.RuntimeStatePending)),
					NewIdChange(30, ChState(infrastructuremanagerv1.RuntimeStateReady), ChProvisioned(true)),
				},
				errMsg:        "",
				listCallCount: 30,
				nowFunc:       fixedNow(),
			},
			{
				// persistent transient error: clock advances past errorDuration after first error poll
				title: "persistent error",
				changes: []IdChange{
					NewIdChange(0, ChState(infrastructuremanagerv1.RuntimeStatePending)),
					NewIdChange(10, ChState(infrastructuremanagerv1.RuntimeStateFailed), ChMessage("some error")),
				},
				errMsg:        `instance alias runtime-id has error "some error"`,
				listCallCount: 10,
				nowFunc: func() func() time.Time {
					now, _ := fakeNowAfter(0, errorDuration+time.Second)
					return now
				}(),
			},
			{
				// same as persistent error — clock advances so it times out even though recovery would come
				title: "long error",
				changes: []IdChange{
					NewIdChange(0, ChState(infrastructuremanagerv1.RuntimeStatePending)),
					NewIdChange(10, ChState(infrastructuremanagerv1.RuntimeStateFailed), ChMessage("some error")),
					NewIdChange(30, ChState(infrastructuremanagerv1.RuntimeStateReady), ChProvisioned(true)),
				},
				errMsg:        `instance alias runtime-id has error "some error"`,
				listCallCount: 10,
				nowFunc: func() func() time.Time {
					now, _ := fakeNowAfter(0, errorDuration+time.Second)
					return now
				}(),
			},
			{
				// user error code: HasTerminalShootError() → fast-fail immediately
				title: "terminal user error code",
				changes: []IdChange{
					NewIdChange(0, ChState(infrastructuremanagerv1.RuntimeStatePending)),
					NewIdChange(5, ChState(infrastructuremanagerv1.RuntimeStateFailed), ChMessage("quota exceeded"),
						ChShootLastErrors([]gardenertypes.LastError{
							{Description: "quota exceeded", Codes: []gardenertypes.ErrorCode{gardenertypes.ErrorInfraQuotaExceeded}},
						})),
				},
				errMsg:        `instance alias runtime-id has terminal shoot error: "quota exceeded"`,
				listCallCount: 5,
				nowFunc:       fixedNow(),
			},
			{
				// terminal LastOperation.State == Failed: fast-fail immediately
				title: "terminal last operation failed",
				changes: []IdChange{
					NewIdChange(0, ChState(infrastructuremanagerv1.RuntimeStatePending)),
					NewIdChange(5, ChState(infrastructuremanagerv1.RuntimeStateFailed), ChMessage("gardener gave up"),
						ChShootLastOperation(&gardenertypes.LastOperation{
							State: gardenertypes.LastOperationStateFailed,
						})),
				},
				errMsg:        `instance alias runtime-id has terminal shoot error: "gardener gave up"`,
				listCallCount: 5,
				nowFunc:       fixedNow(),
			},
			{
				// rate limit code only: transient, should tolerate
				title: "transient rate limit error recovers",
				changes: []IdChange{
					NewIdChange(0, ChState(infrastructuremanagerv1.RuntimeStatePending)),
					NewIdChange(5, ChState(infrastructuremanagerv1.RuntimeStateFailed), ChMessage("rate limited"),
						ChShootLastErrors([]gardenertypes.LastError{
							{Description: "rate limited", Codes: []gardenertypes.ErrorCode{gardenertypes.ErrorInfraRateLimitsExceeded}},
						})),
					NewIdChange(8, ChState(infrastructuremanagerv1.RuntimeStatePending), ChMessage(""), ChShootLastErrors(nil)),
					NewIdChange(15, ChState(infrastructuremanagerv1.RuntimeStateReady), ChProvisioned(true)),
				},
				errMsg:        "",
				listCallCount: 15,
				nowFunc:       fixedNow(),
			},

			// delete ============================
			{
				title: "delete...pending....gone",
				changes: []IdChange{
					NewIdChange(0, ChBeingDeleted(true), ChState(infrastructuremanagerv1.RuntimeStatePending)),
					NewIdChange(10, ChRemove(true)),
				},
				errMsg:        "",
				listCallCount: 10,
				nowFunc:       fixedNow(),
			},
			{
				title: "delete...pending...tmp_error...pending....gone",
				changes: []IdChange{
					NewIdChange(0, ChBeingDeleted(true), ChState(infrastructuremanagerv1.RuntimeStatePending)),
					NewIdChange(10, ChState(infrastructuremanagerv1.RuntimeStateFailed), ChMessage("some error")),
					NewIdChange(13, ChState(infrastructuremanagerv1.RuntimeStatePending), ChMessage("")),
					NewIdChange(20, ChRemove(true)),
				},
				errMsg:        "",
				listCallCount: 20,
				nowFunc:       fixedNow(),
			},
			{
				title: "delete...pending...long_error",
				changes: []IdChange{
					NewIdChange(0, ChBeingDeleted(true), ChState(infrastructuremanagerv1.RuntimeStatePending)),
					NewIdChange(10, ChState(infrastructuremanagerv1.RuntimeStateFailed), ChMessage("some error")),
					NewIdChange(30, ChState(infrastructuremanagerv1.RuntimeStatePending), ChMessage("")),
				},
				errMsg:        `instance alias runtime-id has error "some error"`,
				listCallCount: 10,
				nowFunc: func() func() time.Time {
					now, _ := fakeNowAfter(0, errorDuration+time.Second)
					return now
				}(),
			},
		}

		for _, tc := range testCases {
			t.Run(tc.title, func(t *testing.T) {
				t.Parallel()
				id, lister := setupInstanceListerMock()
				opts := []WaitOption{
					WithRuntime(id.RuntimeID),
					WithErrorDuration(errorDuration),
					WithNowFunc{fn: tc.nowFunc},
					WithSleeperFunc(func(_ context.Context, _ time.Duration) {}),
				}

				actualListCallCount := 0

				lister.BeforeListCalled(func(i int) error {
					actualListCallCount++
					if i > 100 {
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
