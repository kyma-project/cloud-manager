package e2e

import (
	"context"
	"log/slog"
	"reflect"
	"strings"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	commonscheme "github.com/kyma-project/cloud-manager/pkg/common/scheme"
	"k8s.io/apimachinery/pkg/api/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CleanSkrOptions struct {
	Logger                *slog.Logger
	ExcludeDefaultIpRange bool
}

func CleanSkrNoWait(ctx context.Context, c client.Client, opts *CleanSkrOptions) error {
	if opts == nil {
		opts = &CleanSkrOptions{}
	}

	logger := opts.Logger
	if logger == nil {
		logger = slog.Default()
	}

	knownTypes := commonscheme.SkrScheme.KnownTypes(cloudresourcesv1beta1.GroupVersion)

	var typesToDelete []reflect.Type
	for kind, tp := range knownTypes {
		// Skip List types and nil types
		if strings.HasSuffix(kind, "List") || tp == nil {
			continue
		}
		typesToDelete = append(typesToDelete, tp)
	}

	deletedCount := 0
	skippedCount := 0

	for _, tp := range typesToDelete {
		if tp == nil {
			logger.Debug("skipping nil type")
			continue
		}

		objPtr := reflect.New(tp)
		obj, ok := objPtr.Interface().(client.Object)
		if !ok {
			logger.Debug("skipping non-Object type", "type", tp.Name())
			continue
		}

		// Check type name directly since GetObjectKind().GroupVersionKind().Kind is empty for new instances
		if opts.ExcludeDefaultIpRange && tp.Name() == "IpRange" {
			listObj := &cloudresourcesv1beta1.IpRangeList{}
			err := c.List(ctx, listObj)
			if err == nil && len(listObj.Items) > 0 {
				for _, item := range listObj.Items {
					if item.Name != "default" {
						err := c.Delete(ctx, &item)
						if err != nil {
							logger.Warn("error deleting IpRange", "name", item.Name, "error", err)
						} else {
							deletedCount++
							logger.Info("deleted IpRange", "name", item.Name)
						}
					} else {
						skippedCount++
						logger.Info("skipped default IpRange", "name", item.Name)
					}
				}
				continue
			}
		}

		err := c.DeleteAllOf(ctx, obj)
		if meta.IsNoMatchError(err) {
			logger.Debug("resource type not found in cluster", "kind", tp.Name())
			continue
		}
		if err != nil {
			logger.Warn("error deleting all resources", "kind", tp.Name(), "error", err)
			continue
		}

		// Try to count remaining resources (optional, for logging)
		listType := knownTypes[tp.Name()+"List"]
		if listType != nil {
			list := reflect.New(listType).Interface().(client.ObjectList)
			countErr := c.List(ctx, list)
			if countErr == nil {
				arr, _ := meta.ExtractList(list)
				if len(arr) == 0 {
					deletedCount++
					logger.Info("deleted all resources", "kind", tp.Name())
				} else {
					logger.Info("deletion initiated", "kind", tp.Name(), "remaining", len(arr))
				}
			} else {
				logger.Info("deletion initiated", "kind", tp.Name())
			}
		} else {
			deletedCount++
			logger.Info("deletion initiated", "kind", tp.Name())
		}
	}

	logger.Info("cleanup completed", "deleted", deletedCount, "skipped", skippedCount)
	return nil
}
