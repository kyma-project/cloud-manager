package composed

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type applyType int

const (
	applyUpdate     applyType = 1
	applyServerSide applyType = 2
)

type ObjWithConditionsAndState interface {
	ObjWithConditions
	State() string
	SetState(v string)
}

type ObjWithConditions interface {
	client.Object
	Conditions() *[]metav1.Condition
	GetObjectMeta() *metav1.ObjectMeta
}

type ObjWithCloneForPatch interface {
	ObjWithConditions
	CloneForPatch() client.Object
}

type ObjWithCloneForPatchStatus interface {
	ObjWithConditions
	CloneForPatchStatus() client.Object
}

type ObjWithDeriveStateFromConditions interface {
	ObjWithConditions
	DeriveStateFromConditions() (changed bool)
}

func PatchStatus(obj ObjWithConditions) *UpdateStatusBuilder {
	return &UpdateStatusBuilder{
		applyType: applyServerSide,
		obj:       obj,
	}
}

func UpdateStatus(obj ObjWithConditions) *UpdateStatusBuilder {
	return &UpdateStatusBuilder{
		applyType: applyUpdate,
		obj:       obj,
	}
}

type UpdateStatusBuilder struct {
	applyType          applyType
	obj                ObjWithConditions
	conditionsToKeep   map[string]struct{}
	conditionsToRemove map[string]struct{}
	conditionsToSet    []metav1.Condition
	updateErrorLogMsg  string
	successLogMsg      string
	failedError        error
	successError       error
	successErrorNil    bool
	updateErrorWrapper func(err error) error
	onUpdateError      func(ctx context.Context, err error) (error, context.Context)
	onUpdateSuccess    func(ctx context.Context) (error, context.Context)
	conditionsToState  func(obj ObjWithConditions) (string, bool)
}

func (b *UpdateStatusBuilder) KeepConditions(conditionTypes ...string) *UpdateStatusBuilder {
	if b.conditionsToKeep == nil {
		b.conditionsToKeep = map[string]struct{}{}
	}
	b.conditionsToRemove = nil
	for _, c := range conditionTypes {
		b.conditionsToKeep[c] = struct{}{}
	}
	return b
}

func (b *UpdateStatusBuilder) RemoveConditions(conditionTypes ...string) *UpdateStatusBuilder {
	if b.conditionsToRemove == nil {
		b.conditionsToRemove = map[string]struct{}{}
	}
	b.conditionsToKeep = nil
	for _, c := range conditionTypes {
		b.conditionsToRemove[c] = struct{}{}
	}
	return b
}

func (b *UpdateStatusBuilder) RemoveConditionIfReasonMatched(conditionType, conditionReason string) *UpdateStatusBuilder {
	if b.conditionsToRemove == nil {
		b.conditionsToRemove = map[string]struct{}{}
	}
	b.conditionsToKeep = nil
	for i := range *b.obj.Conditions() {
		condition := (*b.obj.Conditions())[i]
		if condition.Type == conditionType && condition.Reason == conditionReason {
			b.conditionsToRemove[conditionType] = struct{}{}
		}
	}
	return b
}

func (b *UpdateStatusBuilder) SetCondition(cond metav1.Condition) *UpdateStatusBuilder {
	b.conditionsToSet = append(b.conditionsToSet, cond)
	return b
}

func (b *UpdateStatusBuilder) SetExclusiveConditions(conditions ...metav1.Condition) *UpdateStatusBuilder {
	// Remove all conditions and set the new ones passed as argument
	if b.conditionsToRemove == nil {
		b.conditionsToRemove = map[string]struct{}{}
	}
	for _, c := range *b.obj.Conditions() {
		b.conditionsToRemove[c.Type] = struct{}{}
	}
	b.conditionsToSet = conditions
	return b
}

func (b *UpdateStatusBuilder) ErrorLogMessage(msg string) *UpdateStatusBuilder {
	b.updateErrorLogMsg = msg
	return b
}

func (b *UpdateStatusBuilder) SuccessLogMsg(msg string) *UpdateStatusBuilder {
	b.successLogMsg = msg
	return b
}

func (b *UpdateStatusBuilder) UpdateErrorWrapper(f func(err error) error) *UpdateStatusBuilder {
	b.updateErrorWrapper = f
	return b
}

func (b *UpdateStatusBuilder) OnUpdateError(f func(ctx context.Context, err error) (error, context.Context)) *UpdateStatusBuilder {
	b.onUpdateError = f
	return b
}

func (b *UpdateStatusBuilder) OnUpdateSuccess(f func(ctx context.Context) (error, context.Context)) *UpdateStatusBuilder {
	b.onUpdateSuccess = f
	return b
}

func (b *UpdateStatusBuilder) FailedError(err error) *UpdateStatusBuilder {
	b.failedError = err
	return b
}

func (b *UpdateStatusBuilder) SuccessError(err error) *UpdateStatusBuilder {
	b.successError = err
	return b
}

func (b *UpdateStatusBuilder) SuccessErrorNil() *UpdateStatusBuilder {
	b.successErrorNil = true
	b.successError = nil
	return b
}
func (b *UpdateStatusBuilder) DeriveStateFromConditions(f func(obj ObjWithConditions) (string, bool)) *UpdateStatusBuilder {
	b.conditionsToState = f
	return b
}

func (b *UpdateStatusBuilder) Run(ctx context.Context, state State) (error, context.Context) {
	b.setDefaults()

	if b.conditionsToRemove == nil {
		if b.conditionsToKeep != nil {
			b.conditionsToRemove = map[string]struct{}{}
			for _, c := range *b.obj.Conditions() {
				_, keep := b.conditionsToKeep[c.Type]
				if !keep {
					b.conditionsToRemove[c.Type] = struct{}{}
				}
			}
		}
	}

	if b.conditionsToRemove != nil {
		for c := range b.conditionsToRemove {
			_ = meta.RemoveStatusCondition(b.obj.Conditions(), c)
		}
	}

	for _, c := range b.conditionsToSet {
		_ = meta.SetStatusCondition(b.obj.Conditions(), c)
	}

	//Set state based on conditions
	withState, ok := b.obj.(ObjWithConditionsAndState)
	if b.conditionsToState != nil && ok {
		newState, ok := b.conditionsToState(b.obj)
		if ok {
			withState.SetState(newState)
		}
	}
	var err error
	if b.applyType == applyUpdate {
		err = state.UpdateObjStatus(ctx)
	} else {
		err = state.PatchObjStatus(ctx)
	}
	if err != nil {
		err = b.updateErrorWrapper(err)
		return b.onUpdateError(ctx, err)
	}

	if len(b.successLogMsg) > 0 {
		logger := LoggerFromCtx(ctx)
		logger.Info(b.successLogMsg)
	}

	return b.onUpdateSuccess(ctx)
}

func (b *UpdateStatusBuilder) setDefaults() {
	b.updateErrorLogMsg = fmt.Sprintf("Error updating status for %T", b.obj)

	if b.updateErrorWrapper == nil {
		b.updateErrorWrapper = func(err error) error {
			return err
		}
	}

	if b.successError == nil && !b.successErrorNil {
		b.successError = StopAndForget
	}
	if b.failedError == nil {
		b.failedError = StopWithRequeue
	}

	if b.onUpdateError == nil {
		b.onUpdateError = func(ctx context.Context, err error) (error, context.Context) {
			return LogErrorAndReturn(err, b.updateErrorLogMsg, b.failedError, ctx)
		}
	}

	if b.onUpdateSuccess == nil {
		b.onUpdateSuccess = func(ctx context.Context) (error, context.Context) {
			return b.successError, ctx
		}
	}
}
