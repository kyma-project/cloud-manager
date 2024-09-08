package dsl

import (
	"context"
	"errors"
	"github.com/kyma-project/cloud-manager/pkg/testinfra/infraScheme"
	"github.com/kyma-project/cloud-manager/pkg/testinfra/infraTypes"

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

	return errors.New("object is not deleted")
}
