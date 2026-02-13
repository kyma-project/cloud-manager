package e2e

import (
	"context"
	"reflect"
	"strings"

	"github.com/go-logr/logr"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	commonscheme "github.com/kyma-project/cloud-manager/pkg/common/scheme"
	"k8s.io/apimachinery/pkg/api/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CleanSkrNoWait(ctx context.Context, c client.Client) error {
	logger := logr.FromContextOrDiscard(ctx)

	knownTypes := commonscheme.SkrScheme.KnownTypes(cloudresourcesv1beta1.GroupVersion)

	var typesToDelete []reflect.Type
	for kind, tp := range knownTypes {
		// Skip List types and nil types
		if strings.HasSuffix(kind, "List") || tp == nil {
			continue
		}
		typesToDelete = append(typesToDelete, tp)
	}

	for _, tp := range typesToDelete {
		objPtr := reflect.New(tp)
		obj, ok := objPtr.Interface().(client.Object)
		if !ok {
			logger.V(1).Info("skipping non-Object type", "type", tp.Name())
			continue
		}

		err := c.DeleteAllOf(ctx, obj)
		if meta.IsNoMatchError(err) {
			logger.V(1).Info("resource type not found in cluster", "kind", tp.Name())
			continue
		}
		if err != nil {
			logger.Info("error deleting all resources", "kind", tp.Name(), "error", err)
			continue
		}

		logger.Info("deletion initiated", "kind", tp.Name())
	}

	logger.Info("cleanup completed")
	return nil
}
