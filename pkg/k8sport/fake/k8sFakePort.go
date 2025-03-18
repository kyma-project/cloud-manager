package fake

import (
	"context"
	"errors"
	"fmt"
	"github.com/elliotchance/pie/v2"
	composedv2 "github.com/kyma-project/cloud-manager/pkg/composed/v2"
	"github.com/kyma-project/cloud-manager/pkg/k8sport"
	"github.com/kyma-project/cloud-manager/pkg/util"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"strings"
	"sync"
)

type K8sFakePort interface {
	k8sport.K8sPort
	Events() chan string
	Set(objects ...client.Object) error
}

var _ K8sFakePort = &k8sFakePort{}

func NewFakeK8sPortOnDefaultCluster(scheme *runtime.Scheme) K8sFakePort {
	return NewFakeK8sPort(composedv2.DefaultClusterID, scheme)
}

func NewFakeK8sPort(clusterId string, scheme *runtime.Scheme) K8sFakePort {
	return &k8sFakePort{
		clusterId:         clusterId,
		scheme:            scheme,
		fakeEventRecorder: record.NewFakeRecorder(100),
		objects:           map[string][]client.Object{},
	}
}

type k8sFakePort struct {
	m sync.Mutex

	clusterId         string
	fakeEventRecorder *record.FakeRecorder
	scheme            *runtime.Scheme
	objects           map[string][]client.Object
}

func (f *k8sFakePort) Events() chan string {
	return f.fakeEventRecorder.Events
}

func (f *k8sFakePort) Set(objects ...client.Object) error {
	f.m.Lock()
	defer f.m.Unlock()

	return f.setNoLock(objects...)
}

func (f *k8sFakePort) setNoLock(objects ...client.Object) error {
	for _, obj := range objects {
		gvk, err := apiutil.GVKForObject(obj, f.scheme)
		if err != nil {
			return err
		}
		key := fmt.Sprintf("%s/%s", gvk.Group, gvk.Kind)
		f.objects[key] = append(f.objects[key], obj)
	}
	return nil
}

func (f *k8sFakePort) findOneNoLockNoCopy(name types.NamespacedName, obj client.Object) (client.Object, schema.GroupVersionKind, error) {
	gvk, err := apiutil.GVKForObject(obj, f.scheme)
	if err != nil {
		return nil, schema.GroupVersionKind{}, err
	}
	key := fmt.Sprintf("%s/%s", gvk.Group, gvk.Kind)
	arr := f.objects[key]
	for _, o := range arr {
		if o.GetName() == name.Name && o.GetNamespace() == name.Namespace {
			return o, gvk, nil
		}
	}
	return nil, gvk, apierrors.NewNotFound(schema.GroupResource{
		Group:    gvk.Group,
		Resource: gvk.Kind,
	}, name.Name)
}

func (f *k8sFakePort) listNoLockNoCopy(list client.ObjectList, opts ...client.ListOption) error {
	gvk, err := apiutil.GVKForObject(list, f.scheme)
	if err != nil {
		return err
	}
	gvk.Kind = strings.TrimSuffix(gvk.Kind, "List")

	options := client.ListOptions{}
	for _, o := range opts {
		o.ApplyToList(&options)
	}

	if options.FieldSelector != nil && !options.FieldSelector.Empty() {
		return errors.New("field selector not supported")
	}

	key := fmt.Sprintf("%s/%s", gvk.Group, gvk.Kind)
	arr := f.objects[key]
	var items []runtime.Object
	for _, o := range arr {
		if options.Namespace != "" && o.GetNamespace() != options.Namespace {
			continue
		}
		if options.LabelSelector != nil && !options.LabelSelector.Matches(labels.Set(o.GetLabels())) {
			continue
		}
		items = append(items, o)
		if options.Limit != 0 && options.Limit <= int64(len(items)) {
			break
		}
	}
	err = meta.SetList(list, items)
	return err
}

func (f *k8sFakePort) ClusterId() string {
	return f.clusterId
}

func (f *k8sFakePort) LoadStateObj(ctx context.Context) error {
	state := composedv2.StateFromCtx[composedv2.State](ctx)
	return f.LoadObj(ctx, state.Name(), state.Obj())
}

func (f *k8sFakePort) LoadObj(ctx context.Context, name types.NamespacedName, obj client.Object) error {
	f.m.Lock()
	defer f.m.Unlock()
	if util.IsContextDone(ctx) {
		return context.Canceled
	}

	o, _, err := f.findOneNoLockNoCopy(name, obj)
	if err != nil {
		return err
	}

	return util.JsonCloneInto(o, obj)
}

func (f *k8sFakePort) PatchMergeLabels(ctx context.Context, obj client.Object, labels map[string]string) (bool, error) {
	f.m.Lock()
	defer f.m.Unlock()
	if util.IsContextDone(ctx) {
		return false, context.Canceled
	}

	o, _, err := f.findOneNoLockNoCopy(client.ObjectKeyFromObject(obj), obj)
	if err != nil {
		return false, err
	}

	changed := false
	if o.GetLabels() == nil {
		o.SetLabels(map[string]string{})
	}
	for k, v := range labels {
		if o.GetLabels()[k] != v {
			o.GetLabels()[k] = v
			changed = true
		}
	}

	return changed, nil
}

func (f *k8sFakePort) PatchMergeAnnotations(ctx context.Context, obj client.Object, annotations map[string]string) (bool, error) {
	f.m.Lock()
	defer f.m.Unlock()
	if util.IsContextDone(ctx) {
		return false, context.Canceled
	}

	o, _, err := f.findOneNoLockNoCopy(client.ObjectKeyFromObject(obj), obj)
	if err != nil {
		return false, err
	}

	changed := false
	if o.GetAnnotations() == nil {
		o.SetAnnotations(map[string]string{})
	}
	for k, v := range annotations {
		if o.GetAnnotations()[k] != v {
			o.GetAnnotations()[k] = v
			changed = true
		}
	}

	return changed, nil
}

func (f *k8sFakePort) Event(_ context.Context, object client.Object, eventtype, reason, message string) {
	f.fakeEventRecorder.Eventf(object, eventtype, reason, message)
}

func (f *k8sFakePort) Eventf(_ context.Context, object client.Object, eventtype, reason, messageFmt string, args ...interface{}) {
	f.fakeEventRecorder.Eventf(object, eventtype, reason, messageFmt, args...)
}

func (f *k8sFakePort) AnnotatedEventf(_ context.Context, object client.Object, annotations map[string]string, eventtype, reason, messageFmt string, args ...interface{}) {
	f.fakeEventRecorder.AnnotatedEventf(object, annotations, eventtype, reason, messageFmt, args...)
}

func (f *k8sFakePort) List(ctx context.Context, obj client.ObjectList, opts ...client.ListOption) error {
	f.m.Lock()
	defer f.m.Unlock()
	if util.IsContextDone(ctx) {
		return context.Canceled
	}

	err := f.listNoLockNoCopy(obj, opts...)
	if err != nil {
		return err
	}

	arr, err := meta.ExtractList(obj)
	if err != nil {
		return err
	}
	for i, o := range arr {
		arr[i] = o.DeepCopyObject().(client.Object)
	}

	err = meta.SetList(obj, arr)
	return err
}

func (f *k8sFakePort) Create(ctx context.Context, obj client.Object) error {
	f.m.Lock()
	defer f.m.Unlock()
	if util.IsContextDone(ctx) {
		return context.Canceled
	}

	_, gvk, err := f.findOneNoLockNoCopy(client.ObjectKeyFromObject(obj), obj)
	if apierrors.IsNotFound(err) {
		return f.setNoLock(obj)
	}

	return apierrors.NewAlreadyExists(schema.GroupResource{
		Group:    gvk.Group,
		Resource: gvk.Kind,
	}, obj.GetName())
}

func (f *k8sFakePort) Delete(ctx context.Context, obj client.Object) error {
	f.m.Lock()
	defer f.m.Unlock()
	if util.IsContextDone(ctx) {
		return context.Canceled
	}

	_, gvk, err := f.findOneNoLockNoCopy(client.ObjectKeyFromObject(obj), obj)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("%s/%s", gvk.Group, gvk.Kind)
	f.objects[key] = pie.FilterNot(f.objects[key], func(o client.Object) bool {
		return o.GetNamespace() == obj.GetNamespace() && o.GetName() == obj.GetName()
	})

	return nil
}
