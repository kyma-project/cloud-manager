package dsl

import (
	"context"
	"errors"
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	DefaultNfsInstanceHost             = "nfs.instance.local"
	DefaultGcpNfsInstanceFileShareName = "vol1"
	DefaultGcpNfsInstanceCapacityGb    = 1024
	DefaultGcpNfsInstanceConnectMode   = "PRIVATE_SERVICE_ACCESS"
	DefaultGcpNfsInstanceTier          = "BASIC_HDD"
)

func WithNfsInstanceStatusHost(host string) ObjStatusAction {
	return &objStatusAction{
		f: func(obj client.Object) {
			if host == "" {
				host = DefaultNfsInstanceHost
			}
			if x, ok := obj.(*cloudcontrolv1beta1.NfsInstance); ok {
				if len(x.Status.Hosts) == 0 {
					x.Status.Hosts = []string{host}
				}
			}
		},
	}
}

func WithNfsInstanceAws() ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudcontrolv1beta1.NfsInstance); ok {
				x.Spec.Instance.Aws = &cloudcontrolv1beta1.NfsInstanceAws{}
			}
		},
	}
}

func WithNfsInstanceGcp(location string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudcontrolv1beta1.NfsInstance); ok {
				x.Spec.Instance.Gcp = &cloudcontrolv1beta1.NfsInstanceGcp{}
				x.Spec.Instance.Gcp.ConnectMode = DefaultGcpNfsInstanceConnectMode
				x.Spec.Instance.Gcp.CapacityGb = DefaultGcpNfsInstanceCapacityGb
				x.Spec.Instance.Gcp.FileShareName = DefaultGcpNfsInstanceFileShareName
				x.Spec.Instance.Gcp.Location = location
				x.Spec.Instance.Gcp.Tier = DefaultGcpNfsInstanceTier
			}
		},
	}
}

func WithNfsInstanceIpRange(ipRangeName string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudcontrolv1beta1.NfsInstance); ok {
				x.Spec.IpRange.Name = ipRangeName
				return
			}
			panic("unhandled type in WithNfsInstanceIpRange")
		},
	}
}

func WithNfsInstanceScope(scopeName string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudcontrolv1beta1.NfsInstance); ok {
				x.Spec.Scope.Name = scopeName
				return
			}
			panic("unhandled type in WithNfsInstanceScope")
		},
	}
}

func WithRemoteRef(name string) ObjAction {
	remoteRef := &cloudcontrolv1beta1.RemoteRef{
		Name:      name,
		Namespace: DefaultSkrNamespace,
	}
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudcontrolv1beta1.NfsInstance); ok {
				x.Spec.RemoteRef = *remoteRef
				return
			}
			panic("unhandled type in WithNfsInstanceScope")
		},
	}
}

func CreateNfsInstance(ctx context.Context, clnt client.Client, obj *cloudcontrolv1beta1.NfsInstance, opts ...ObjAction) error {
	if obj == nil {
		obj = &cloudcontrolv1beta1.NfsInstance{}
	}
	NewObjActions(opts...).
		Append(
			WithNamespace(DefaultKcpNamespace),
		).
		ApplyOnObject(obj)

	if obj.Name == "" {
		return errors.New("the KCP NfsInstance must have name set")
	}

	err := clnt.Create(ctx, obj)
	return err
}

func UpdateNfsInstance(ctx context.Context, clnt client.Client, obj *cloudcontrolv1beta1.NfsInstance, opts ...ObjAction) error {
	if obj == nil {
		return errors.New("for updating the KCP NfsInstance, the object must be provided")
	}
	obj.Spec.Instance.Gcp.CapacityGb = 2 * DefaultGcpNfsInstanceCapacityGb
	NewObjActions(opts...).
		Append(
			WithNamespace(DefaultKcpNamespace),
		).
		ApplyOnObject(obj)

	if obj.Name == "" {
		return errors.New("the KCP NfsInstance must have name set")
	}

	err := clnt.Update(ctx, obj)
	return err
}

func DeleteNfsInstance(ctx context.Context, clnt client.Client, obj *cloudcontrolv1beta1.NfsInstance, opts ...ObjAction) error {
	if obj == nil {
		return errors.New("for deleting the KCP NfsInstance, the object must be provided")
	}
	NewObjActions(opts...).
		Append(
			WithNamespace(DefaultKcpNamespace),
		).
		ApplyOnObject(obj)

	if obj.Name == "" {
		return errors.New("the KCP NfsInstance must have name set")
	}

	err := clnt.Delete(ctx, obj)
	return err
}

func HavingNfsInstanceStatusId() ObjAssertion {
	return func(obj client.Object) error {
		x, ok := obj.(*cloudcontrolv1beta1.NfsInstance)
		if !ok {
			return fmt.Errorf("the object %T is not SKR AwsNfsVolume", obj)
		}
		if x.Status.Id == "" {
			return errors.New("the SKR AwsNfsVolume ID not set")
		}
		return nil
	}
}
