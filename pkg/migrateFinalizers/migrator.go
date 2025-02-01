package migrateFinalizers

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type migratorOptions struct {
	*finalizerInfo
	Client    client.Client
	List      client.ObjectList
	Logger    logr.Logger
	Namespace string
}

// migrator runs a migration of adding new and removing old finalizers on a single kind
// for the given ObjectList
type migrator interface {
	Migrate(ctx context.Context, options *migratorOptions) error
}

func newMigrator() migrator {
	return &defaultMigrator{}
}

var _ migrator = &defaultMigrator{}

type defaultMigrator struct{}

func (m *defaultMigrator) Migrate(ctx context.Context, options *migratorOptions) error {
	var listOptions []client.ListOption
	if options.Namespace != "" {
		listOptions = append(listOptions, client.InNamespace(options.Namespace))
	}
	err := options.Client.List(ctx, options.List, listOptions...)
	if meta.IsNoMatchError(err) {
		// kind not installed in this cluster
		return nil
	}
	if err != nil {
		return fmt.Errorf("error listing objects: %w", err)
	}

	arr, err := meta.ExtractList(options.List)
	if err != nil {
		return fmt.Errorf("error extracting objects from list: %w", err)
	}
	for _, o := range arr {
		obj, ok := o.(client.Object)
		if !ok {
			options.Logger.Info(fmt.Sprintf("Skipping a non client.Client object: %T", o))
			continue
		}
		logger := options.Logger.WithValues(
			"namespace", obj.GetNamespace(),
			"name", obj.GetName(),
		)

		// only replace finalizers if object already has some of the old finalizers
		// otherwise just ignore it and do not change it
		hasSomeOldFinalizer := false
		for _, finalizer := range options.RemoveFinalizers {
			if controllerutil.ContainsFinalizer(obj, finalizer) {
				hasSomeOldFinalizer = true
			}
		}
		if !hasSomeOldFinalizer {
			continue
		}

		for _, finalizer := range options.AddFinalizers {
			added, err := composed.PatchObjAddFinalizer(ctx, finalizer, obj, options.Client)
			if err != nil {
				return fmt.Errorf("error adding finalizer %q to %s/%s: %w", finalizer, obj.GetNamespace(), obj.GetName(), err)
			}
			if added {
				logger.WithValues("finalizer", finalizer).Info("Added finalizer")
			}
		}
		for _, finalizer := range options.RemoveFinalizers {
			removed, err := composed.PatchObjRemoveFinalizer(ctx, finalizer, obj, options.Client)
			if err != nil {
				return fmt.Errorf("error adding finalizer %q to %s/%s: %w", finalizer, obj.GetNamespace(), obj.GetName(), err)
			}
			if removed {
				logger.WithValues("finalizer", finalizer).Info("Removed finalizer")
			}
		}
	}

	return nil
}
