package dsl

import (
	"context"
	"errors"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Check(obj client.Object, asserts ...ObjAssertion) error {
	err := NewObjAssertions(asserts).
		AssertObj(obj)
	return err
}

func LoadAndCheck(ctx context.Context, clnt client.Client, obj client.Object, loadingOps ObjActions, asserts ...ObjAssertion) error {
	if obj == nil {
		return errors.New("the LoadAndCheck object can not be nil")
	}

	actions := NewObjActions(loadingOps...)

	switch obj.(type) {
	case *cloudcontrolv1beta1.IpRange,
		*cloudcontrolv1beta1.NfsInstance,
		*cloudcontrolv1beta1.Scope,
		*cloudcontrolv1beta1.VpcPeering:
		actions = actions.Append(WithNamespace(DefaultKcpNamespace))
	case *cloudresourcesv1beta1.IpRange,
		*cloudresourcesv1beta1.AwsNfsVolume,
		*cloudresourcesv1beta1.GcpNfsVolume:
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
