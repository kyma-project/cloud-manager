package dsl

import (
	"context"
	"errors"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func LoadAndCheck(ctx context.Context, clnt client.Client, obj client.Object, loadingOps ObjActions, asserts ...ObjAssertion) error {
	if obj == nil {
		return errors.New("the LoadAndCheck object can not be nil")
	}

	actions := NewObjActions(loadingOps...)

	switch obj.(type) {
	case *cloudcontrolv1beta1.IpRange,
		*cloudcontrolv1beta1.NfsInstance,
		*cloudcontrolv1beta1.Scope,
		*cloudcontrolv1beta1.VpcPeering,
		*cloudcontrolv1beta1.RedisInstance,
		*cloudcontrolv1beta1.Network:
		actions = actions.Append(WithNamespace(DefaultKcpNamespace))
	case *cloudresourcesv1beta1.IpRange,
		*cloudresourcesv1beta1.AwsNfsVolume,
		*cloudresourcesv1beta1.GcpNfsVolume,
		*cloudresourcesv1beta1.GcpNfsVolumeBackup,
		*cloudresourcesv1beta1.GcpNfsVolumeRestore,
		*cloudresourcesv1beta1.GcpRedisInstance,
		*cloudresourcesv1beta1.AwsRedisInstance,
		*cloudresourcesv1beta1.AzureRedisInstance:
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

	switch obj.(type) {
	case *cloudcontrolv1beta1.IpRange,
		*cloudcontrolv1beta1.NfsInstance,
		*cloudcontrolv1beta1.Scope,
		*cloudcontrolv1beta1.VpcPeering,
		*cloudcontrolv1beta1.RedisInstance:
		actions = actions.Append(WithNamespace(DefaultKcpNamespace))
	case *cloudresourcesv1beta1.IpRange,
		*cloudresourcesv1beta1.AwsNfsVolume,
		*cloudresourcesv1beta1.GcpNfsVolume,
		*cloudresourcesv1beta1.GcpNfsVolumeBackup,
		*cloudresourcesv1beta1.GcpNfsVolumeRestore,
		*cloudresourcesv1beta1.GcpRedisInstance,
		*cloudresourcesv1beta1.AwsRedisInstance,
		*cloudresourcesv1beta1.AzureRedisInstance:
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
