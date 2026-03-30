package keb

import (
	"context"
	"fmt"
	"testing"

	"github.com/elliotchance/pie/v2"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/stretchr/testify/assert"
)

// TestableInstanceDetails ========================================================

type TestableInstanceDetails struct {
	*InstanceDetails
	RemoveMe bool
}

func (tid *TestableInstanceDetails) Change(ch IdChange) {
	if ch.state != nil {
		tid.State = *ch.state
	}
	if ch.beingDeleted != nil {
		tid.BeingDeleted = *ch.beingDeleted
	}
	if ch.provisioningCompleted != nil {
		tid.ProvisioningCompleted = *ch.provisioningCompleted
	}
	if ch.message != nil {
		tid.Message = *ch.message
	}
	if ch.removeMe != nil {
		tid.RemoveMe = *ch.removeMe
	}
}

type IdChange struct {
	callCount int

	beingDeleted          *bool
	state                 *string
	message               *string
	provisioningCompleted *bool
	removeMe              *bool
}

func NewIdChange(callCount int, opts ...IdChangeOption) IdChange {
	ch := &IdChange{callCount: callCount}
	for _, o := range opts {
		o(ch)
	}
	return *ch
}

type IdChangeOption func(*IdChange)

func ChRemove(v bool) IdChangeOption {
	return func(ch *IdChange) {
		ch.removeMe = &v
	}
}

func ChBeingDeleted(v bool) IdChangeOption {
	return func(idChange *IdChange) {
		idChange.beingDeleted = &v
	}
}

func ChState(v string) IdChangeOption {
	return func(idChange *IdChange) {
		idChange.state = &v
	}
}

func ChMessage(v string) IdChangeOption {
	return func(idChange *IdChange) {
		idChange.message = &v
	}
}

func ChProvisioned(v bool) IdChangeOption {
	return func(idChange *IdChange) {
		idChange.provisioningCompleted = &v
	}
}

// InstanceListerMock ========================================================

type InstanceListerMock struct {
	items         []*TestableInstanceDetails
	listCallCount int
	cb            func(int) error
}

func (il *InstanceListerMock) Add(ids ...*TestableInstanceDetails) *InstanceListerMock {
	il.items = append(il.items, ids...)
	return il
}

func (il *InstanceListerMock) BeforeListCalled(cb func(int) error) {
	il.cb = cb
}

func (il *InstanceListerMock) List(_ context.Context, opts ...ListOption) ([]InstanceDetails, error) {
	// apply before list called callbacks to mutate items
	il.listCallCount++
	if il.cb != nil {
		if err := il.cb(il.listCallCount); err != nil {
			return nil, err
		}
	}

	// remove marked items
	il.items = pie.FilterNot(il.items, func(i *TestableInstanceDetails) bool {
		return i.RemoveMe
	})

	options := &listOptions{}
	for _, o := range opts {
		o.ApplyOnList(options)
	}

	var results []InstanceDetails
	for _, id := range il.items {
		ok := options.runtimeId == "" || id.RuntimeID == options.runtimeId
		ok = ok && options.alias == "" || id.Alias == options.alias
		ok = ok && options.provider == "" || id.Provider == options.provider
		ok = ok && options.globalAccount == "" || id.GlobalAccount == options.globalAccount
		ok = ok && options.subAccount == "" || id.SubAccount == options.subAccount
		if ok {
			x := *(id.InstanceDetails)
			results = append(results, x)
		}
	}
	return results, nil
}

// tests =============================================================

func newInstanceDetails() *TestableInstanceDetails {
	return &TestableInstanceDetails{
		InstanceDetails: &InstanceDetails{
			Alias:                 "alias",
			GlobalAccount:         "global-account",
			SubAccount:            "sub-account",
			Provider:              cloudcontrolv1beta1.ProviderGCP,
			Region:                "us-east1",
			ProvisioningCompleted: false,
			RuntimeID:             "runtime-id",
			ShootName:             "shoot-name",
			State:                 "",
			Message:               "",
			BeingDeleted:          false,
			Ignored:               false,
		},
	}
}

func setupInstanceListerMock() (*TestableInstanceDetails, *InstanceListerMock) {
	lister := &InstanceListerMock{}

	dummy1 := newInstanceDetails()
	dummy1.RuntimeID = "dummy-1-runtime-id"
	dummy1.Alias = "dummy-1-alias"
	lister.Add(dummy1)

	id := newInstanceDetails()
	lister.Add(id)

	dummy2 := newInstanceDetails()
	dummy2.RuntimeID = "dummy-2-runtime-id"
	dummy2.Alias = "dummy-2-alias"
	lister.Add(dummy2)

	return id, lister
}

func TestInstanceListerMock(t *testing.T) {
	testCases := []struct {
		title  string
		cb     func(int) error
		opts   []ListOption
		count  int
		errMsg string
	}{
		{
			"finds one by runtime-id",
			nil,
			[]ListOption{WithRuntime("runtime-id")},
			1,
			"",
		},
		{
			"finds one by alias",
			nil,
			[]ListOption{WithAlias("alias")},
			1,
			"",
		},
		{
			"returns all",
			nil,
			nil,
			3,
			"",
		},
		{
			"returns error when cb returns error",
			func(_ int) error {
				return fmt.Errorf("some error")
			},
			nil,
			0,
			"some error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			id, lister := setupInstanceListerMock()
			ctx := context.Background()

			lister.BeforeListCalled(tc.cb)

			arr, err := lister.List(ctx, tc.opts...)
			if tc.errMsg != "" {
				assert.EqualError(t, err, tc.errMsg)
			} else {
				assert.Equal(t, tc.count, len(arr))
				if tc.count == 1 {
					assert.Equal(t, id.RuntimeID, arr[0].RuntimeID)
					assert.Equal(t, id.Alias, arr[0].Alias)
				}
			}
		})
	}
}
