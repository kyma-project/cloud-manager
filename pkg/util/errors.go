package util

import (
	"context"
	"errors"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
)

func IgnoreForbidden(err error) error {
	if apierrors.IsForbidden(err) {
		return nil
	}
	return err
}

func IgnoreNoMatch(err error) error {
	if meta.IsNoMatchError(err) {
		return nil
	}
	return err
}

func IgnoreContextCanceled(err error) error {
	if errors.Is(err, context.Canceled) {
		return nil
	}
	return err
}

func IgnoreContextDeadlineExceeded(err error) error {
	if errors.Is(err, context.DeadlineExceeded) {
		return nil
	}
	return err
}

func IgnoreContextCanceledAndDeadlineExceeded(err error) error {
	return IgnoreContextCanceled(IgnoreContextDeadlineExceeded(err))
}
