package dsl

import (
	"context"
	"errors"
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	scope2 "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

func GivenAwsScopeExists(ctx context.Context, clnt client.Client, name string) error {
	shootNamespace := scope2.ScopeConfig.GardenerNamespace
	project := strings.TrimPrefix(shootNamespace, "garden-")
	scope := &cloudcontrolv1beta1.Scope{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: DefaultKcpNamespace,
			Name:      name,
		},
		Spec: cloudcontrolv1beta1.ScopeSpec{
			KymaName:  name,
			ShootName: name,
			Region:    "us-west-1",
			Provider:  cloudcontrolv1beta1.ProviderAws,
			Scope: cloudcontrolv1beta1.ScopeInfo{
				Aws: &cloudcontrolv1beta1.AwsScope{
					VpcNetwork: fmt.Sprintf("shoot--%s--%s", project, name),
					Network: cloudcontrolv1beta1.AwsNetwork{
						Nodes:    "10.250.0.0/22",
						Pods:     "10.96.0.0/13",
						Services: "10.104.0.0/13",
					},
				},
			},
		},
	}
	err := clnt.Get(ctx, client.ObjectKeyFromObject(scope), scope)
	if client.IgnoreNotFound(err) != nil {
		return err
	}
	if apierrors.IsNotFound(err) {
		err := clnt.Create(ctx, scope)
		if err != nil {
			return err
		}
	}
	return nil
}

func CreateAwsNfsVolumeBackup(ctx context.Context, clnt client.Client, obj *cloudresourcesv1beta1.AwsNfsVolumeBackup, opts ...ObjAction) error {
	if obj == nil {
		obj = &cloudresourcesv1beta1.AwsNfsVolumeBackup{}
	}
	NewObjActions(opts...).
		Append(
			WithNamespace(DefaultSkrNamespace),
		).
		ApplyOnObject(obj)

	if obj.Name == "" {
		return errors.New("the SKR AwsNfsVolumeBackup must have name set")
	}
	if obj.Spec.Source.Volume.Name == "" {
		return errors.New("the SKR AwsNfsVolumeBackup must have spec.source.volume.name set")
	}
	if obj.Spec.Source.Volume.Namespace == "" {
		obj.Spec.Source.Volume.Namespace = DefaultSkrNamespace
	}

	err := clnt.Get(ctx, client.ObjectKeyFromObject(obj), obj)
	if err == nil {
		// already exists
		return nil
	}
	if client.IgnoreNotFound(err) != nil {
		// some error
		return err
	}
	err = clnt.Create(ctx, obj)
	return err
}

func WithAwsNfsVolume(name string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.AwsNfsVolumeBackup); ok {
				x.Spec.Source.Volume.Name = name
				if x.Spec.Source.Volume.Namespace == "" {
					x.Spec.Source.Volume.Namespace = DefaultSkrNamespace
				}
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAwsNfsVolume", obj))
		},
	}
}

func AssertAwsNfsVolumeBackupHasState(state string) ObjAssertion {
	return func(obj client.Object) error {
		x, ok := obj.(*cloudresourcesv1beta1.AwsNfsVolumeBackup)
		if !ok {
			return fmt.Errorf("the object %T is not AwsNfsVolumeBackup", obj)
		}
		if x.Status.State == "" {
			return errors.New("the AwsNfsVolumeBackup state not set")
		}
		if x.Status.State != state {
			return fmt.Errorf("the AwsNfsVolumeBackup state is %s, expected %s", x.Status.State, state)
		}
		return nil
	}
}
