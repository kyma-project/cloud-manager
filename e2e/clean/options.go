package clean

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/elliotchance/pie/v2"
	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type match struct {
	gvkObj  schema.GroupVersionKind
	gvkList schema.GroupVersionKind
	obj     client.Object
	list    client.ObjectList

	objectVersions map[string]string
}

func (m *match) mapConditionToString(obj composed.ObjWithConditions) string {
	arr := pie.Map(ptr.Deref(obj.Conditions(), nil), func(cond metav1.Condition) string {
		return fmt.Sprintf("{%s %s %s %s}", cond.Type, cond.Status, cond.Reason, cond.Message)
	})
	return strings.Join(arr, ", ")
}

func (m *match) objectCount() int {
	return len(m.objectVersions)
}

func (m *match) observeObjects(arr []client.Object, logger logr.Logger) {
	logger = logger.WithValues("gvk", m.gvkObj.String())
	if m.objectVersions == nil {
		m.objectVersions = make(map[string]string, len(arr))
	}
	for _, obj := range arr {
		key := client.ObjectKeyFromObject(obj)
		currentVersion := obj.GetResourceVersion()
		oldVersion, ok := m.objectVersions[key.String()]
		if !ok {
			oldVersion = currentVersion
		}
		m.objectVersions[key.String()] = obj.GetResourceVersion()
		if oldVersion == currentVersion {
			continue
		}
		objWithConditions, ok := obj.(composed.ObjWithConditions)
		if !ok {
			continue
		}
		logger.
			WithValues("name", key).
			WithValues("conditions", m.mapConditionToString(objWithConditions)).
			Info("observed")
	}
}

type options struct {
	client   client.Client
	scheme   *runtime.Scheme
	matchers []Matcher

	matches []*match

	wait bool

	pollInterval time.Duration

	timeout time.Duration

	sleeper util.Sleeper

	forceDeleteOnTimeout bool

	logger logr.Logger

	dryRun bool
}

func (o *options) validate() error {
	var result []error
	if o.client == nil {
		result = append(result, errors.New("client is required"))
	}
	if o.scheme == nil {
		result = append(result, errors.New("scheme is required"))
	}
	if o.matchers == nil {
		result = append(result, errors.New("matcher is required"))
	}

	if o.pollInterval == 0 {
		o.pollInterval = 10 * time.Second
	}
	if o.timeout == 0 {
		o.timeout = 30 * time.Minute
	}
	if o.sleeper == nil {
		o.sleeper = util.SleeperFunc(util.RealSleeperFunc)
	}

	if len(result) > 0 {
		return errors.Join(result...)
	}
	return nil
}

func (o *options) observe(gvkList schema.GroupVersionKind, tp reflect.Type) error {
	if !strings.HasSuffix(gvkList.Kind, "List") {
		return nil
	}
	ok := false
	for _, m := range o.matchers {
		if m(gvkList, tp) {
			ok = true
			break
		}
	}
	if !ok {
		return nil
	}

	gvkObj := gvkList
	gvkObj.Kind = strings.TrimSuffix(gvkObj.Kind, "List")

	obj, err := o.scheme.New(gvkObj)
	if err != nil {
		return fmt.Errorf("could not obtain object from scheme: %w", err)
	}
	list, err := o.scheme.New(gvkList)
	if err != nil {
		return fmt.Errorf("could not obtain list from scheme: %w", err)
	}

	cObj, ok := obj.(client.Object)
	if !ok {
		return nil
	}
	cList, ok := list.(client.ObjectList)
	if !ok {
		return nil
	}

	o.matches = append(o.matches, &match{
		gvkObj:  gvkObj,
		gvkList: gvkList,
		obj:     cObj,
		list:    cList,
	})

	return nil
}

func (o *options) deleteAll(ctx context.Context) error {
	for _, m := range o.matches {
		list := m.list.DeepCopyObject().(client.ObjectList)
		err := o.client.List(ctx, list)
		if client.IgnoreNotFound(err) != nil && util.IgnoreNoMatch(err) != nil {
			return fmt.Errorf("could not list objects on %T: %w", list, err)
		}
		arr, err := util.ExtractList(list)
		if err != nil {
			return fmt.Errorf("could not list objects on %T: %w", list, err)
		}
		m.observeObjects(arr, o.logger)
		if m.objectCount() > 0 {
			objectNames := pie.Map(arr, func(obj client.Object) string {
				return client.ObjectKeyFromObject(obj).String()
			})
			msg := "deleting..."
			if o.dryRun {
				msg += " (dry-run)"
			}
			o.logger.
				WithValues("objectCount", len(objectNames)).
				WithValues("objectNames", pie.Join(objectNames, ", ")).
				WithValues("gvk", m.gvkObj.String()).
				Info(msg)

			if o.dryRun {
				continue
			}

			err := o.client.DeleteAllOf(ctx, m.obj)
			if client.IgnoreNotFound(err) != nil && util.IgnoreNoMatch(err) != nil {
				return fmt.Errorf("error deleting all of %T: %w", m.obj, err)
			}
		}
	}
	return nil
}

func (o *options) allGone(ctx context.Context) (bool, error) {
	for _, m := range o.matches {
		if m.objectCount() == 0 {
			continue
		}
		list := m.list.DeepCopyObject().(client.ObjectList)
		o.logger.WithValues("gvk", m.gvkObj.String()).Info("listing...")
		err := o.client.List(ctx, list)
		if client.IgnoreNotFound(err) != nil && util.IgnoreNoMatch(err) != nil {
			return false, fmt.Errorf("error listing %T: %w", m.list, err)
		}
		arr, err := util.ExtractList(list)
		if err != nil {
			return false, fmt.Errorf("could not extract list from %T: %w", list, err)
		}
		m.observeObjects(arr, o.logger)
		if len(arr) > 0 {
			return false, nil
		}
	}

	return true, nil
}

func (o *options) waitAllGone(ctx context.Context) (bool, error) {
	if !o.wait || o.dryRun {
		return false, nil
	}

	o.logger.Info("waiting for all objects to be deleted...")

	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, o.timeout)
	defer cancel()
	for {
		if util.IsContextDone(ctx) {
			return false, ctx.Err()
		}
		ok, err := o.allGone(ctx)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}

		select {
		case <-ctx.Done():
			return false, nil
		default:
			o.sleeper.Sleep(ctx, o.pollInterval)
		}
	}
}

func (o *options) forceDelete(ctx context.Context) error {
	if !o.forceDeleteOnTimeout || o.dryRun {
		return nil
	}

	for _, m := range o.matches {
		list := m.list.DeepCopyObject().(client.ObjectList)
		err := o.client.List(ctx, list)
		if client.IgnoreNotFound(err) != nil && util.IgnoreNoMatch(err) != nil {
			return fmt.Errorf("error listing %T: %w", m.list, err)
		}
		if err != nil {
			continue
		}
		arr, err := util.ExtractList(list)
		if err != nil {
			return fmt.Errorf("could not extract list from %T: %w", list, err)
		}

		for _, obj := range arr {
			if len(obj.GetFinalizers()) == 0 {
				continue
			}
			o.logger.
				WithValues("gvk", m.gvkObj.String()).
				WithValues("name", client.ObjectKeyFromObject(obj)).
				Info("force deleting / removing finalizers")

			p := []byte(`{"metadata":{"finalizers":[]}}`)
			err := o.client.Patch(ctx, obj, client.RawPatch(types.MergePatchType, p))
			if client.IgnoreNotFound(err) != nil && util.IgnoreNoMatch(err) != nil {
				reason := metav1.StatusReasonUnknown
				var code int32
				if status, ok := err.(apierrors.APIStatus); ok || errors.As(err, &status) {
					reason = status.Status().Reason
					code = status.Status().Code
				}
				return fmt.Errorf("could not remove finalizers %T (reason: %s, code: %d): %w", err, reason, code, err)
			}
			err = o.client.Get(ctx, client.ObjectKeyFromObject(obj), obj)
			if client.IgnoreNotFound(err) != nil && util.IgnoreNoMatch(err) != nil {
				return fmt.Errorf("failed to load obj after finalizer remove: %w", err)
			}
			if len(obj.GetFinalizers()) > 0 {
				return fmt.Errorf("expected no finalizers after finalizer remove")
			}
		}
	}
	return nil
}
