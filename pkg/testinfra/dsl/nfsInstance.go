package dsl

import (
	"context"
	"errors"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	DefaultNfsInstanceHost = "nfs.instance.local"
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
