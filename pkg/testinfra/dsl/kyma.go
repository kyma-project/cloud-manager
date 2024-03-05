package dsl

import (
	"context"
	"errors"
	"fmt"
	"github.com/kyma-project/cloud-manager/pkg/testinfra"
	"github.com/kyma-project/cloud-manager/pkg/util"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	DefaultModuleName = "cloud-manager"
)

func WithKymaSpecChannel(channel string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if channel == "" {
				channel = "dev"
			}
			x, ok := obj.(*unstructured.Unstructured)
			if ok {
				val, exists, err := unstructured.NestedString(x.Object, "spec", "channel")
				if err != nil {
					panic(err)
				}
				if exists && val == channel {
					return
				}
				err = unstructured.SetNestedField(x.Object, channel, "spec", "channel")
				if err != nil {
					panic(err)
				}
			}
		},
	}
}

func WithKymaModuleListedInSpec() ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			x, ok := obj.(*unstructured.Unstructured)
			if ok {
				err := util.SetKymaModuleInSpec(x, DefaultModuleName)
				if err != nil {
					panic(err)
				}
			}
		},
	}
}

func WithKymaModuleRemovedInSpec() ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			x, ok := obj.(*unstructured.Unstructured)
			if ok {
				err := util.RemoveKymaModuleFromSpec(x, DefaultModuleName)
				if err != nil {
					panic(err)
				}
			}
		},
	}
}

func WithKymaStatusModuleState(state util.KymaModuleState) ObjStatusAction {
	return &objStatusAction{
		f: func(obj client.Object) {
			x, ok := obj.(*unstructured.Unstructured)
			if ok {
				err := util.SetKymaModuleStateFromStatus(x, DefaultModuleName, state)
				if err != nil {
					panic(err)
				}
			}
		},
	}
}

func CreateKymaCR(ctx context.Context, infra testinfra.Infra, kymaCR *unstructured.Unstructured, opts ...ObjAction) error {
	if kymaCR == nil {
		kymaCR = util.NewKymaUnstructured()
	}

	NewObjActions(opts...).
		Append(
			WithNamespace(DefaultKcpNamespace),
		).
		ApplyOnObject(kymaCR)

	if kymaCR.GetName() == "" {
		return errors.New("the Kyma CR must have name set")
	}

	kymaCR.SetLabels(map[string]string{
		"kyma-project.io/shoot-name": kymaCR.GetName(),
	})

	err := infra.KCP().Client().Get(ctx, client.ObjectKeyFromObject(kymaCR), kymaCR)
	if err == nil {
		// already exist
		return nil
	}
	if client.IgnoreNotFound(err) != nil {
		// some error
		return err
	}
	err = infra.KCP().Client().Create(ctx, kymaCR)
	if err != nil {
		return err
	}

	// Kubeconfig secret
	{
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: kymaCR.GetNamespace(),
				Name:      fmt.Sprintf("kubeconfig-%s", kymaCR.GetName()),
			},
		}
		err := infra.KCP().Client().Get(ctx, client.ObjectKeyFromObject(secret), secret)
		if client.IgnoreNotFound(err) != nil {
			return err
		}
		if apierrors.IsNotFound(err) {
			b, err := kubeconfigToBytes(restConfigToKubeconfig(infra.SKR().Cfg()))
			if err != nil {
				return fmt.Errorf("error getting SKR kubeconfig bytes: %w", err)
			}
			secret.Data = map[string][]byte{
				"config": b,
			}

			err = infra.KCP().Client().Create(ctx, secret)
			if client.IgnoreAlreadyExists(err) != nil {
				return fmt.Errorf("error creating SKR secret: %w", err)
			}
		}
	}

	return nil
}

func KymaCRModuleStateUpdate(ctx context.Context, kcpClient client.Client, kymaCR *unstructured.Unstructured, opts ...ObjAction) error {
	if kymaCR == nil {
		kymaCR = util.NewKymaUnstructured()
	}

	NewObjActions(opts...).
		Append(
			WithNamespace(DefaultKcpNamespace),
		).
		ApplyOnObject(kymaCR).
		ApplyOnStatus(kymaCR)

	err := kcpClient.Status().Update(ctx, kymaCR)
	if err != nil {
		return err
	}

	return nil
}
