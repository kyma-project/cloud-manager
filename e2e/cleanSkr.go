package e2e

import (
	"context"
	"reflect"
	"strings"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	commonscheme "github.com/kyma-project/cloud-manager/pkg/common/scheme"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CleanSkrNoWait(ctx context.Context, c client.Client) error {
	logger := composed.LoggerFromCtx(ctx)

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
		logger.Info(tp.Name())
		objPtr := reflect.New(tp)
		obj, ok := objPtr.Interface().(client.Object)
		if !ok {
			continue
		}

		err := c.DeleteAllOf(ctx, obj)
		if meta.IsNoMatchError(err) || apierrors.IsNotFound(err) {
			logger.Info("not found")
			continue
		}
		if err != nil {
			logger.Error(err, "Unexpected error")
			continue
		}
	}

	logger.Info("cleanup completed")
	return nil
}
