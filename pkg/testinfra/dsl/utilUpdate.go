package dsl

import (
	"context"
	"errors"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func UpdateObj(ctx context.Context, clnt client.Client, obj client.Object, opts ...ObjAction) error {
	if obj == nil {
		return errors.New("the object for UpdateObj() can not be nil")
	}

	NewObjActions(opts...).
		ApplyOnObject(obj)

	err := clnt.Update(ctx, obj)
	if err != nil {
		return err
	}

	return nil
}

func UpdateStatus(ctx context.Context, clnt client.Client, obj client.Object, opts ...ObjAction) error {
	if obj == nil {
		return errors.New("the object for UpdateStatus() can not be nil")
	}

	NewObjActions(opts...).
		ApplyOnStatus(obj)

	err := clnt.Status().Update(ctx, obj)
	return err
}

func UpdatePvPhase(ctx context.Context, clnt client.Client, obj *corev1.PersistentVolume, phase corev1.PersistentVolumePhase) error {
	if obj == nil {
		return errors.New("the PersistentVolume for UpdatePvPhase() can not be nil")
	}

	if err := clnt.Get(ctx, client.ObjectKeyFromObject(obj), obj); err != nil {
		return err
	}

	obj.Status.Phase = phase

	err := clnt.Status().Update(ctx, obj)
	return err
}
