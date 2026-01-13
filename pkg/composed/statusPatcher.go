package composed

import (
	"context"
	"fmt"
	"slices"
	"time"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ObjWithStatus interface {
	client.Object
	Conditions() *[]metav1.Condition
	ObservedGeneration() int64
	SetObservedGeneration(int64)
}

type StatusPatcher[T ObjWithStatus] struct {
	objWithoutChanges T
	obj               T
}

// NewStatusPatcher handles proper status changes. Do not modify obj spec from the parcher creation to the patch call.
// Ideally should be short-lived used only for one status change. For (if) any other, a new patcher instance should
// be created.
//
// IMPORTANT! All changes made to the object status MUST be done after patcher is created, and before call to Run(). Status
// changes made before patcher is created will be omitted.
//
//	obj := loadTheObj()
//
//	// Phase 1 - initial mark of transient status as stale - once per generation, run at the start of the loop
//	if sp := NewStausPatcher(obj); sp.IsStale() {
//		// preferably kind has defined precise status conditions and mutating funcs
//		obj.SetStatusProcessing()
//		// or, sadly you're wandering and have not systemized status conditions
//		sp.SetConditions(metav1.Condition{
//			Type: "Ready",
//			Status: metav1.ConditionUnknown,
//			Reason: "Processing",
//			Message: "Processing",
//			// no need to set condition ObservedGeneration to status ObservedGeneration, patcher does it for you
//		})
//		if err := sp.Patch(ctx, client); err != nil {
//			return ctrl.Result{}, err
//		}
//		// Do NOT return — continue to work
//	}
//
//	// Phase 2 - reconcile the object and make actial state closed to desired
//	doTheActualWork(ctx, obj)
//
//	// Phase 3 - terminal status - after work is done at the end of the loop
//	// must create new StatusPatch so it observes spec changes done in Phase 2 as the status comparing base
//	sp := NewStausPatcher(obj)
//	// preferably kind has defined precise status conditions and mutating funcs
//	obj.SetStatusReady()
//	// or, sadly you're wandering and have not systemized status conditions
//	sp.SetConditions(metav1.Condition{
//		Type: "Ready",
//		Status: metav1.ConditionTrue,
//		Reason: "Ready",
//		Message: "Ready",
//		// no need to set condition ObservedGeneration to status ObservedGeneration, patcher does it for you
//	})
//	if err := sp.Patch(ctx, client); err != nil {
//		return ctrl.Result{}, err
//	}
//	return ctrl.Result{}, nil
func NewStatusPatcher[T ObjWithStatus](obj T) *StatusPatcher[T] {
	return &StatusPatcher[T]{
		objWithoutChanges: obj.DeepCopyObject().(T),
		obj:               obj,
	}
}

// IsStale returns true if object generation is different then observed generation
func (u *StatusPatcher[T]) IsStale() bool {
	if u.obj.GetGeneration() != u.obj.ObservedGeneration() {
		return true
	}
	if u.obj.Conditions() == nil || len(*u.obj.Conditions()) == 0 {
		return true
	}
	return false
}

// MutateStatus provides a convenient API to mutate object status in the fluid form
func (u *StatusPatcher[T]) MutateStatus(cb func(T)) *StatusPatcher[T] {
	cb(u.obj)
	return u
}

// SetConditions providers a helper/backup API to set object conditions, but ideally it should have
// own status mutating functions that are called from MutateStatus()
func (u *StatusPatcher[T]) SetConditions(conditions ...metav1.Condition) *StatusPatcher[T] {
	for _, c := range conditions {
		c.ObservedGeneration = u.obj.GetGeneration()
		meta.SetStatusCondition(u.obj.Conditions(), c)
	}
	return u
}

// RemoveConditions providers a helper/backup API to set object conditions, but ideally it should have
// own status mutating functions that are called from MutateStatus()
func (u *StatusPatcher[T]) RemoveConditions(conditionTypes ...string) *StatusPatcher[T] {
	for _, ct := range conditionTypes {
		meta.RemoveStatusCondition(u.obj.Conditions(), ct)
	}
	return u
}

// Patch call the k8s api patch with changes made on the object since the patcher was created. It also
// sets the observed generation to the object generation.
func (u *StatusPatcher[T]) Patch(ctx context.Context, c client.Client) error {
	if u.obj.GetGeneration() != u.obj.ObservedGeneration() {
		u.obj.SetObservedGeneration(u.obj.GetGeneration())
	}
	return c.Status().Patch(ctx, u.obj, client.MergeFrom(u.objWithoutChanges))
}

// ===============

type StatusPatchErrorHandler func(context.Context, error) (bool, error)

func Continue(_ context.Context, err error) (bool, error) {
	return true, nil
}

func Requeue(_ context.Context, err error) (bool, error) {
	return true, StopWithRequeue
}

func RequeueAfter(requeueAfter time.Duration) StatusPatchErrorHandler {
	return func(_ context.Context, err error) (bool, error) {
		return true, StopWithRequeueDelay(requeueAfter)
	}
}

func Forget(_ context.Context, err error) (bool, error) {
	return true, StopAndForget
}

func LogError(err error, msg string) StatusPatchErrorHandler {
	return func(ctx context.Context, _ error) (bool, error) {
		logger := LoggerFromCtx(ctx)
		logger.Error(err, msg)
		return false, nil
	}
}

func Log(msg string) StatusPatchErrorHandler {
	return func(ctx context.Context, err error) (bool, error) {
		logger := LoggerFromCtx(ctx)
		if err == nil {
			logger.Info(msg)
		} else {
			logger.Error(err, msg)
		}
		return false, err
	}
}

// NewStatusPatcherComposed handles proper status changes and composed Action results. This patcher embeds the StatusPatcher
// and introduces new functions:
//   - Run() that returns Action results
//   - OnSuccess(), OnFailure() and OnStatusChanged() that defines behavior and Action results returned after patch has been done
//
// IMPORTANT! All changes made to the object status MUST be done after patcher is created, and before call to Run(). Status
// changes made before patcher is created will be omitted.
//
// Do not modify obj spec from the parcher creation to the patch call.
// Patcher should be short-lived and used only for one status change. For (if) any other, a new patcher instance should
// be created.
//
// Use .OnSuccess() and .OnFailure() to add custom outcome handlers of the type StatusPatchErrorHandler. Handlers are
// appended to the list and executed in the given and added order. The execution of handlers stops when first of them
// returns action result. By default, the patcher adds OnSuccess(Continue()), OnFailure(Requeue(), Log(genericMsg)).
//
// Use .OnStatusChanged() to add handlers that are called on success but only if status has changed. If an object is
// already in the desired status, then its status will not change, no next reconciliation will be triggered, and no
// status changge handlers will be called. All added status changed handlers are called, and their action result if any
// given is ignored. Only success handlers can affect the action result.
//
//	// obj is loaded and reconciled upon
//	// this is the terminal status change after all work has been done in an Action
//	return NewStatusPatcherComposed(obj).
//		// preferably kind has defined precise status conditions and mutating funcs
//		MutateStatus(func (obj MyObjType) {
//			obj.SetStatusReady()
//		}).
//		// or, sadly you're wandering and have not systemized status conditions
//		SetConditions(metav1.Condition{
//			Type: "Ready",
//			Status: metav1.ConditionTrue,
//			Reason: "Ready",
//			Message: "Provisioned",
//			// no need to set ObservedGeneration, since patcher does it for you
//		}).
//		OnFailure(Log("Error patching status of the SomeKind to ready status")).
//		OnSuccess(Requeue()).
//		Run(ctx, client)
func NewStatusPatcherComposed[T ObjWithStatus](obj T) *StatusPatcherComposed[T] {
	return &StatusPatcherComposed[T]{
		StatusPatcher: NewStatusPatcher(obj),
		failureHandlers: []StatusPatchErrorHandler{
			Requeue,
			Log(fmt.Sprintf("failed to patch status for object %T %s/%s", obj, obj.GetNamespace(), obj.GetName())),
		},
		successHandlers: []StatusPatchErrorHandler{
			Continue,
		},
		// none is added by default, and all are executed
		statusChangedHandlers: []StatusPatchErrorHandler{},
	}
}

type StatusPatcherComposed[T ObjWithStatus] struct {
	*StatusPatcher[T]
	successHandlers       []StatusPatchErrorHandler
	statusChangedHandlers []StatusPatchErrorHandler
	failureHandlers       []StatusPatchErrorHandler
}

// OnFailure adds handlers that will be called on failed patch - ie when non nil error is returned.
// They are executed in given order all up until
// some provides composed action result. This means on Log() execution will continue to the next, but on Continue(),
// RequeueAfter() and similar the handler execution will stop and no other successive handlers will be called.
func (u *StatusPatcherComposed[T]) OnFailure(handlers ...StatusPatchErrorHandler) *StatusPatcherComposed[T] {
	for _, handler := range slices.Backward(handlers) {
		u.failureHandlers = append(u.failureHandlers, handler)
	}
	return u
}

// OnSuccess adds handlers that will be called on successful patch. They are executed in given order all up until
// some provides composed action result. This means on Log() execution will continue to the next, but on Continue(),
// RequeueAfter() and similar the handler execution will stop and no other successive handlers will be called.
func (u *StatusPatcherComposed[T]) OnSuccess(handlers ...StatusPatchErrorHandler) *StatusPatcherComposed[T] {
	for _, handler := range slices.Backward(handlers) {
		u.successHandlers = append(u.successHandlers, handler)
	}
	return u
}

// OnStatusChanged adds handlers that will be called on successful patch but only if the status is changed. They
// are all executed and do not affect the composed action result, so only valid handler is Log().
func (u *StatusPatcherComposed[T]) OnStatusChanged(handlers ...StatusPatchErrorHandler) *StatusPatcherComposed[T] {
	for _, handler := range slices.Backward(handlers) {
		u.statusChangedHandlers = append(u.statusChangedHandlers, handler)
	}
	return u
}

// Run the patch and call appropriate handlers depending on if status has changed (object has new revision) and
// the returned error.
func (u *StatusPatcherComposed[T]) Run(ctx context.Context, c client.Client) (error, context.Context) {
	resourceVersionBeforePatch := u.obj.GetResourceVersion()
	err := u.Patch(ctx, c)
	var result error
	if err != nil {
		for _, h := range slices.Backward(u.failureHandlers) {
			handled, res := h(ctx, err)
			if handled {
				result = res
				break
			}
		}
	} else {
		if resourceVersionBeforePatch != u.obj.GetResourceVersion() {
			// statusChangedHandlers can not affect the composed result and all are executed
			for _, h := range slices.Backward(u.statusChangedHandlers) {
				_, _ = h(ctx, err)
			}
		}
		for _, h := range slices.Backward(u.successHandlers) {
			handled, res := h(ctx, err)
			if handled {
				result = res
				break
			}
		}
	}
	return result, ctx
}

// reimplement inner StatusPatcher methods due to return type

func (u *StatusPatcherComposed[T]) SetConditions(conditions ...metav1.Condition) *StatusPatcherComposed[T] {
	u.StatusPatcher.SetConditions(conditions...)
	return u
}

func (u *StatusPatcherComposed[T]) RemoveConditions(conditionTypes ...string) *StatusPatcherComposed[T] {
	u.StatusPatcher.RemoveConditions(conditionTypes...)
	return u
}

func (u *StatusPatcherComposed[T]) MutateStatus(cb func(T)) *StatusPatcherComposed[T] {
	u.StatusPatcher.MutateStatus(cb)
	return u
}
