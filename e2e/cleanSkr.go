package e2e

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/bootstrap"
	"k8s.io/apimachinery/pkg/api/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func cleanSkrNoWait(ctx context.Context, c client.Client) error {
	knownTypes := bootstrap.SkrScheme.KnownTypes(cloudresourcesv1beta1.GroupVersion)
	for kind, tp := range knownTypes {
		if !strings.HasSuffix(kind, "List") {
			continue
		}

		list := reflect.New(tp).Interface().(client.ObjectList)
		err := c.List(ctx, list)
		if meta.IsNoMatchError(err) {
			continue
		}
		if err != nil {
			return fmt.Errorf("error listing kind %s: %w", kind, err)
		}

		list.GetResourceVersion()
		arr, err := meta.ExtractList(list)
		if err != nil {
			return fmt.Errorf("error extracting %s list as dependants on %s: %w", kind, tp, err)
		}

		for _, rtObj := range arr {
			obj := rtObj.(client.Object)
			err = c.Delete(ctx, obj)
			if err != nil {
				return fmt.Errorf("error deleting %s %s/%s object: %w", kind, obj.GetNamespace(), obj.GetName(), err)
			}
		}
	}

	return nil
}
