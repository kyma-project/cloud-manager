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
		if strings.HasSuffix(kind, "List") {
			continue
		}
		typesToDelete = append(typesToDelete, tp)
	}

	deletedCount := 0
	skippedCount := 0

	for _, tp := range typesToDelete {
		objPtr := reflect.New(tp)
		obj, ok := objPtr.Interface().(client.Object)
		if !ok {
			logger.Debug("skipping non-Object type", "type", tp.Name())
			continue
		}
		kind := obj.GetObjectKind().GroupVersionKind().Kind

		if opts.ExcludeDefaultIpRange && kind == "IpRange" {
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
			logger.Debug("resource type not found in cluster", "kind", kind)
			continue
		}
		if err != nil {
			logger.Warn("error deleting all resources", "kind", kind, "error", err)
			continue
		}

		list := reflect.New(knownTypes[kind+"List"]).Interface().(client.ObjectList)
		countErr := c.List(ctx, list)
		if countErr == nil {
			arr, _ := meta.ExtractList(list)
			if len(arr) == 0 {
				deletedCount++
				logger.Info("deleted all resources", "kind", kind)
			} else {
				logger.Info("deletion initiated", "kind", kind, "remaining", len(arr))
			}
		} else {
			logger.Info("deletion initiated", "kind", kind)
		}
	}

	logger.Info("cleanup completed", "deleted", deletedCount, "skipped", skippedCount)
	return nil
}
