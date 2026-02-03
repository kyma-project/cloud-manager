package composed

import (
	"context"
	"errors"
	"testing"

	"github.com/elliotchance/pie/v2"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	commonscheme "github.com/kyma-project/cloud-manager/pkg/common/scheme"
	"github.com/stretchr/testify/assert"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

type statusPatchErrorHandlerMock struct {
	h         StatusPatchErrorHandler
	callCount int
}

func (m *statusPatchErrorHandlerMock) handle(ctx context.Context, err error) (bool, error) {
	m.callCount++
	if m.h != nil {
		return m.h(ctx, err)
	}
	return false, nil
}

type statusPatchErrorHandlerMocks []*statusPatchErrorHandlerMock

func (h statusPatchErrorHandlerMocks) toHandlers() []StatusPatchErrorHandler {
	return pie.Map(h, func(x *statusPatchErrorHandlerMock) StatusPatchErrorHandler {
		return x.handle
	})
}

func TestStatusPatcherComposed(t *testing.T) {

	objName := "test-subscription"

	create := func(err error) (client.Client, *cloudcontrolv1beta1.Subscription) {
		obj := cloudcontrolv1beta1.NewSubscriptionBuilder().
			WithName(objName).
			WithNamespace("default").
			WithAws("test-account-id").
			Build()
		cb := fake.NewClientBuilder().
			WithScheme(commonscheme.KcpScheme).
			WithRuntimeObjects(obj).
			WithStatusSubresource(obj)
		var c client.Client
		if err != nil {
			cb.WithInterceptorFuncs(interceptor.Funcs{
				SubResourcePatch: func(ctx context.Context, client client.Client, subResourceName string, obj client.Object, patch client.Patch, opts ...client.SubResourcePatchOption) error {
					if subResourceName == "status" {
						return err
					}
					return c.Status().Patch(ctx, obj, patch, opts...)
				},
			})
		}
		c = cb.Build()
		return c, obj
	}

	mutateStatusReady := func(x *cloudcontrolv1beta1.Subscription) {
		x.SetStatusReady()
	}

	hm := func(h StatusPatchErrorHandler) *statusPatchErrorHandlerMock {
		return &statusPatchErrorHandlerMock{h: h}
	}

	verify := func(
		t *testing.T,
		patchErr error,
		mutate bool,
		expectedError error,
		onSuccess statusPatchErrorHandlerMocks,
		expectedOnSuccessCalls []bool,
		onStatusChanged statusPatchErrorHandlerMocks,
		expectedOnStatusChangedCalls []bool,
		onFailure statusPatchErrorHandlerMocks,
		expectedOnFailureCalls []bool,
	) {
		assert.Equal(t, len(onSuccess), len(expectedOnSuccessCalls), "invalid test case: len(onSuccess) != len(expectedOnSuccessCalls)")
		assert.Equal(t, len(onStatusChanged), len(expectedOnStatusChangedCalls), "invalid test case: len(onStatusChanged) != len(expectedOnStatusChangedCalls)")
		assert.Equal(t, len(onFailure), len(expectedOnFailureCalls), "invalid test case: len(onFailure) != len(expectedOnFailureCalls)")
		c, patchedObj := create(patchErr)
		p := NewStatusPatcherComposed(patchedObj)
		if mutate {
			p = p.MutateStatus(mutateStatusReady)
		}
		if onSuccess != nil {
			p = p.OnSuccess(onSuccess.toHandlers()...)
		}
		if onStatusChanged != nil {
			p = p.OnStatusChanged(onStatusChanged.toHandlers()...)
		}
		err, ctx := p.Run(context.Background(), c)
		if expectedError != nil {
			assert.EqualError(t, err, expectedError.Error())
		} else {
			assert.NoError(t, err)
		}
		for i, hm := range onSuccess {
			if expectedOnSuccessCalls[i] {
				assert.True(t, hm.callCount == 1, "onSuccess handler %d expected to be called once, but was called %d times", i, hm.callCount)
			} else {
				assert.True(t, hm.callCount == 0, "onSuccess handler %d not expected to be called, but was called %d times", i, hm.callCount)
			}
		}
		for i, hm := range onStatusChanged {
			if expectedOnStatusChangedCalls[i] {
				assert.True(t, hm.callCount == 1, "onStatusChanged handler %d expected to be called once, but was called %d times", i, hm.callCount)
			} else {
				assert.True(t, hm.callCount == 0, "onStatusChanged handler %d not expected to be called, but was called %d times", i, hm.callCount)
			}
		}

		if patchErr != nil {
			// no change was saved
			return
		}

		// Primary goal of the test is to verify which handlers will be called,
		// this is not really necessary, but we also verify that StatusPatcher actually really calls the client
		// to patch the status, though the change will happen only if mutate is true
		// Not really crucial since client.Patch is obvious in the StatusPatcher, where
		// handler calling is not so obvious and needs detailed testing
		loadedObj := &cloudcontrolv1beta1.Subscription{}
		err = c.Get(ctx, client.ObjectKeyFromObject(patchedObj), loadedObj)
		assert.NoError(t, err)
		assert.Equal(t, patchedObj.Status, loadedObj.Status)
	}

	t.Run("success_mutate", func(t *testing.T) {
		testCases := []struct {
			title                        string
			expectedError                error
			onStatusChanged              statusPatchErrorHandlerMocks
			expectedOnStatusChangedCalls []bool
			onSuccess                    statusPatchErrorHandlerMocks
			expectedOnSuccessCalls       []bool
			onFailure                    statusPatchErrorHandlerMocks
			expectedOnFailureCalls       []bool
		}{
			{
				"by default it continues", nil,
				nil, nil,
				nil, nil,
				nil, nil,
			},
			{
				"doesnt call any failure handler", nil,
				nil, nil,
				nil, nil,
				statusPatchErrorHandlerMocks{hm(nil)}, []bool{false},
			},
			{
				"calls all given onStatusChanged handlers", nil,
				statusPatchErrorHandlerMocks{hm(nil), hm(Forget), hm(nil)}, []bool{true, true, true},
				nil, nil,
				nil, nil,
			},
			{
				"calls given success handlers until flow control provided", StopWithRequeue,
				nil, nil,
				statusPatchErrorHandlerMocks{hm(nil), hm(Requeue), hm(nil)}, []bool{true, true, false},
				nil, nil,
			},
			{
				"status changed handlers called and success handlers called until flow control provided", StopAndForget,
				statusPatchErrorHandlerMocks{hm(nil), hm(Forget), hm(nil)}, []bool{true, true, true},
				statusPatchErrorHandlerMocks{hm(nil), hm(Forget), hm(nil), hm(Requeue)}, []bool{true, true, false, false},
				nil, nil,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.title, func(t *testing.T) {
				verify(t, nil, true, tc.expectedError, tc.onSuccess, tc.expectedOnSuccessCalls, tc.onStatusChanged, tc.expectedOnStatusChangedCalls, tc.onFailure, tc.expectedOnFailureCalls)
			})
		}
	})

	t.Run("success_no_change", func(t *testing.T) {
		testCases := []struct {
			title                        string
			expectedError                error
			onStatusChanged              statusPatchErrorHandlerMocks
			expectedOnStatusChangedCalls []bool
			onSuccess                    statusPatchErrorHandlerMocks
			expectedOnSuccessCalls       []bool
			onFailure                    statusPatchErrorHandlerMocks
			expectedOnFailureCalls       []bool
		}{
			{
				"by default it continues", nil,
				nil, nil,
				nil, nil,
				nil, nil,
			},
			{
				"doesnt call any failure handler", nil,
				nil, nil,
				nil, nil,
				statusPatchErrorHandlerMocks{hm(nil)}, []bool{false},
			},
			{
				"calls no onStatusChanged handlers", nil,
				statusPatchErrorHandlerMocks{hm(nil), hm(Forget), hm(nil)}, []bool{false, false, false},
				nil, nil,
				nil, nil,
			},
			{
				"calls given success handlers until flow control provided", StopWithRequeue,
				nil, nil,
				statusPatchErrorHandlerMocks{hm(nil), hm(Requeue), hm(nil)}, []bool{true, true, false},
				nil, nil,
			},
			{
				"status changed handlers are not called and success handlers called until flow control provided", StopAndForget,
				statusPatchErrorHandlerMocks{hm(nil), hm(Forget), hm(nil)}, []bool{false, false, false},
				statusPatchErrorHandlerMocks{hm(nil), hm(Forget), hm(nil), hm(Requeue)}, []bool{true, true, false, false},
				nil, nil,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.title, func(t *testing.T) {
				verify(t, nil, false, tc.expectedError, tc.onSuccess, tc.expectedOnSuccessCalls, tc.onStatusChanged, tc.expectedOnStatusChangedCalls, tc.onFailure, tc.expectedOnFailureCalls)
			})
		}
	})

	t.Run("failure_mutate", func(t *testing.T) {
		testCases := []struct {
			title                        string
			expectedError                error
			onStatusChanged              statusPatchErrorHandlerMocks
			expectedOnStatusChangedCalls []bool
			onSuccess                    statusPatchErrorHandlerMocks
			expectedOnSuccessCalls       []bool
			onFailure                    statusPatchErrorHandlerMocks
			expectedOnFailureCalls       []bool
		}{
			{
				"by default it requeues", StopWithRequeue,
				nil, nil,
				nil, nil,
				nil, nil,
			},
			{
				"calls all failure handlers", StopWithRequeue,
				nil, nil,
				nil, nil,
				statusPatchErrorHandlerMocks{hm(nil), hm(nil)}, []bool{false, false},
			},
			{
				"doesnt call any success handler", StopWithRequeue,
				nil, nil,
				statusPatchErrorHandlerMocks{hm(nil), hm(Forget), hm(nil)}, []bool{false, false, false},
				nil, nil,
			},
			{
				"doesnt call any status changed handler", StopWithRequeue,
				statusPatchErrorHandlerMocks{hm(nil), hm(Forget), hm(nil)}, []bool{false, false, false},
				nil, nil,
				nil, nil,
			},
			{
				"calls failure handlers until some returns flow control", StopWithRequeue,
				nil, nil,
				nil, nil,
				statusPatchErrorHandlerMocks{hm(nil), hm(Forget), hm(nil), hm(Requeue)}, []bool{true, true, false, false},
			},
		}

		patchErr := apierrors.NewConflict(
			schema.GroupResource{Group: cloudcontrolv1beta1.GroupVersion.Group, Resource: "subscriptions"},
			objName,
			errors.New("status conflict"),
		)

		for _, tc := range testCases {
			t.Run(tc.title, func(t *testing.T) {
				verify(t, patchErr, true, tc.expectedError, tc.onSuccess, tc.expectedOnSuccessCalls, tc.onStatusChanged, tc.expectedOnStatusChangedCalls, tc.onFailure, tc.expectedOnFailureCalls)
			})
		}

	})
}
