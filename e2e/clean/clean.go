package clean

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
)

type Option func(*options)

func Clean(ctx context.Context, opts ...Option) error {
	o := &options{
		logger: logr.Discard(),
	}
	for _, opt := range opts {
		opt(o)
	}
	if err := o.validate(); err != nil {
		return err
	}

	if o.dryRun {
		fmt.Printf("cleaning dry-run\n")
	}

	for gvk, tp := range o.scheme.AllKnownTypes() {
		if err := o.observe(gvk, tp); err != nil {
			return err
		}
	}

	if err := o.deleteAll(ctx); err != nil {
		return err
	}

	if !o.wait {
		return nil
	}

	allGone, err := o.waitAllGone(ctx)
	if err != nil && !errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("wait all gone failed: %w", err)
	}

	if !allGone {
		if err := o.forceDelete(ctx); err != nil {
			return fmt.Errorf("force delete failed: %w", err)
		}
	}

	return nil
}
