package dsl

import (
	"context"
	"errors"
	"fmt"

	"github.com/elliotchance/pie/v2"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/testinfra/infraScheme"
	"github.com/kyma-project/cloud-manager/pkg/testinfra/infraTypes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func LoadAndCheck(ctx context.Context, clnt client.Client, obj client.Object, loadingOps ObjActions, asserts ...ObjAssertion) error {
	if obj == nil {
		return errors.New("the LoadAndCheck object can not be nil")
	}

	actions := NewObjActions(loadingOps...)

	switch infraScheme.ObjToClusterType(obj) {
	case infraTypes.ClusterTypeKcp:
		actions = actions.Append(WithNamespace(DefaultKcpNamespace))
	case infraTypes.ClusterTypeSkr:
		actions = actions.Append(WithNamespace(DefaultSkrNamespace))
	}

	actions.ApplyOnObject(obj)

	if obj.GetName() == "" {
		panic("the LoadAndCheck object must have a name")
	}

	if err := clnt.Get(ctx, client.ObjectKeyFromObject(obj), obj); err != nil {
		return err
	}

	err := NewObjAssertions(asserts).
		AssertObj(obj)
	return err
}

func IsDeleted(ctx context.Context, clnt client.Client, obj client.Object, opts ...ObjAction) error {
	if obj == nil {
		return errors.New("the IsDeleted object can not be nil")
	}

	actions := NewObjActions(opts...)

	switch infraScheme.ObjToClusterType(obj) {
	case infraTypes.ClusterTypeKcp:
		actions = actions.Append(WithNamespace(DefaultKcpNamespace))
	case infraTypes.ClusterTypeSkr:
		actions = actions.Append(WithNamespace(DefaultSkrNamespace))
	}

	actions.ApplyOnObject(obj)

	err := clnt.Get(ctx, client.ObjectKeyFromObject(obj), obj)

	if apierrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}

	state := ""
	if x, ok := obj.(composed.ObjWithConditionsAndState); ok {
		state = x.State()
	}
	conditions := []string{}
	if x, ok := obj.(composed.ObjWithConditions); ok {
		conditions = pie.Map(*x.Conditions(), func(c metav1.Condition) string {
			return fmt.Sprintf("%s/%s/{%s}", c.Type, c.Reason, c.Message)
		})
	}
	txtDT := "w/out deletion timestamp"
	if obj.GetDeletionTimestamp() != nil {
		txtDT = "with deletion timestamp"
	}
	return fmt.Errorf("object is not deleted, found %s in state %s with conditions %v", txtDT, state, conditions)
}
