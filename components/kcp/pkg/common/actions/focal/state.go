package focal

import (
	"context"
	"fmt"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/components/kcp/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/components/lib/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type State interface {
	composed.State
	Scope() *cloudresourcesv1beta1.Scope
	SetScope(*cloudresourcesv1beta1.Scope)
	ObjAsCommonObj() CommonObject

	AddReadyCondition(ctx context.Context, message string) error
	AddErrorCondition(ctx context.Context, reason string, err error) error
}

type StateFactory interface {
	NewState(base composed.State) State
}

func NewStateFactory() StateFactory {
	return &stateFactory{}
}

type stateFactory struct{}

func (f *stateFactory) NewState(base composed.State) State {
	return &state{
		State: base,
	}
}

// ========================================================================

type state struct {
	composed.State

	scope *cloudresourcesv1beta1.Scope
}

func (s *state) Scope() *cloudresourcesv1beta1.Scope {
	return s.scope
}

func (s *state) SetScope(scope *cloudresourcesv1beta1.Scope) {
	s.scope = scope
}

func (s *state) ObjAsCommonObj() CommonObject {
	return s.Obj().(CommonObject)
}

func (s *state) AddReadyCondition(ctx context.Context, message string) error {
	meta.SetStatusCondition(s.ObjAsCommonObj().Conditions(), metav1.Condition{
		Type:    cloudresourcesv1beta1.ConditionTypeReady,
		Status:  "True",
		Reason:  cloudresourcesv1beta1.ReasonReady,
		Message: message,
	})

	return s.UpdateObjStatus(ctx)
}

func (s *state) AddErrorCondition(ctx context.Context, reason string, err error) error {
	meta.SetStatusCondition(s.ObjAsCommonObj().Conditions(), metav1.Condition{
		Type:    cloudresourcesv1beta1.ConditionTypeError,
		Status:  "True",
		Reason:  reason,
		Message: fmt.Sprint(err),
	})

	return s.UpdateObjStatus(ctx)
}
