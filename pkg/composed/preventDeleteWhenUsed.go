package composed

import (
	"context"
	"fmt"
	"github.com/elliotchance/pie/v2"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

type UsedCallbackType func(
	ctx context.Context,
	state State,
	list client.ObjectList,
	usedByNames []string,
) (error, context.Context)

// PreventDeleteWhenUsed returns [Action] that checks if reconciled object in state is used
// by the kind specified in the list argument, by listing them on the condition indexFiled of
// dependant object equals to the specified name of the reconciled object.
// The argument [list] must be a list instance of the objects depending on/using/referencing the
// object being reconciled.
// Values of arguments name and indexField should be same as they were used when defining the index
// with [client.FieldIndexer] method [IndexField]. Usually they will be:
//   - just the name of the object, if it's global scope
//   - namespace and name of the object combined, ie namespace/name, when it's namespace scope
func PreventDeleteWhenUsed(list client.ObjectList, name, indexField string, usedCallback UsedCallbackType) Action {
	return func(ctx context.Context, st State) (error, context.Context) {
		reconciledKind := reflect.ValueOf(st.Obj()).Elem().Type().Name()
		dependantKind := strings.TrimSuffix(reflect.ValueOf(list).Elem().Type().Name(), "List")

		v := reflect.ValueOf(list)
		v.Elem().Type().Name()
		if !MarkedForDeletionPredicate(ctx, st) {
			// SKR IpRange is NOT marked for deletion, do not delete mirror in KCP
			return nil, nil
		}

		logger := LoggerFromCtx(ctx)

		listOps := &client.ListOptions{
			FieldSelector: fields.OneTermEqualSelector(indexField, name),
		}

		err := st.Cluster().K8sClient().List(ctx, list, listOps)
		if meta.IsNoMatchError(err) {
			// kind is unknown to the api - aka crd not installed
			return nil, nil
		}
		if err != nil {
			return LogErrorAndReturn(err, fmt.Sprintf("Error listing %s that are using %s", dependantKind, reconciledKind), StopWithRequeue, ctx)
		}

		arr, err := meta.ExtractList(list)
		if err != nil {
			logger.Error(err, fmt.Sprintf("Error extracting %s list as dependants on %s", dependantKind, reconciledKind))
			return StopAndForget, ctx
		}

		if len(arr) == 0 {
			return nil, nil
		}

		usedByNames := pie.Map(arr, func(x runtime.Object) string {
			xx := x.(client.Object)
			return fmt.Sprintf("%s/%s", xx.GetNamespace(), xx.GetName())
		})

		logger.
			WithValues(fmt.Sprintf("usedBy%s", dependantKind), fmt.Sprintf("%v", usedByNames)).
			Info(fmt.Sprintf("%s marked for deleting used by %s", reconciledKind, dependantKind))

		return usedCallback(ctx, st, list, usedByNames)
	}
}
