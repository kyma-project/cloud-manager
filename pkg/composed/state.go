package composed

import (
	"context"
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func NewStateClusterFromCluster(clstr cluster.Cluster) StateCluster {
	return NewStateCluster(clstr.GetClient(), clstr.GetAPIReader(), clstr.GetEventRecorderFor("cloud-manager"), clstr.GetScheme())
}

func NewStateCluster(
	clnt client.Client,
	reader client.Reader,
	eventRecorder record.EventRecorder,
	scheme *runtime.Scheme,
) StateCluster {
	return &stateCluster{
		client:        clnt,
		reader:        reader,
		eventRecorder: eventRecorder,
		scheme:        scheme,
	}
}

type StateCluster interface {
	K8sClient() client.Client
	ApiReader() client.Reader
	EventRecorder() record.EventRecorder
	Scheme() *runtime.Scheme
}

type stateCluster struct {
	client        client.Client
	reader        client.Reader
	eventRecorder record.EventRecorder
	scheme        *runtime.Scheme
}

func (c *stateCluster) K8sClient() client.Client {
	return c.client
}

func (c *stateCluster) ApiReader() client.Reader {
	return c.reader
}

func (c *stateCluster) EventRecorder() record.EventRecorder {
	return c.eventRecorder
}

func (c *stateCluster) Scheme() *runtime.Scheme {
	return c.scheme
}

type State interface {
	// Cluster returns the used cluster
	Cluster() StateCluster

	// K8sClient returns the client to the cluster
	//
	// Deprecated: Use Cluster().K8sClient() instead
	K8sClient() client.Client

	// EventRecorder returns the event recorder of the connected cluster
	//
	// Deprecated: Use Cluster().EventRecorder() instead
	EventRecorder() record.EventRecorder

	// Scheme returns the Scheme
	//
	// Deprecated: Use Cluster().Scheme() instead
	Scheme() *runtime.Scheme

	Name() types.NamespacedName
	Obj() client.Object
	SetObj(client.Object)

	LoadObj(ctx context.Context, opts ...client.GetOption) error
	UpdateObj(ctx context.Context, opts ...client.UpdateOption) error
	UpdateObjStatus(ctx context.Context, opts ...client.SubResourceUpdateOption) error
	PatchObjStatus(ctx context.Context) error

	PatchObjAddFinalizer(ctx context.Context, f string) (bool, error)
	PatchObjRemoveFinalizer(ctx context.Context, f string) (bool, error)
}

type StateFactory interface {
	NewState(name types.NamespacedName, obj client.Object) State
}

func NewStateFactory(clstr StateCluster) StateFactory {
	return &stateFactory{cluster: clstr}
}

type stateFactory struct {
	cluster StateCluster
}

func (f *stateFactory) NewState(name types.NamespacedName, obj client.Object) State {
	return &baseState{
		cluster: f.cluster,
		name:    name,
		obj:     obj,
	}
}

// ========================================================================

type baseState struct {
	cluster StateCluster
	//client        client.Client
	//eventRecorder record.EventRecorder
	//scheme        *runtime.Scheme

	name types.NamespacedName
	obj  client.Object
}

func (s *baseState) Cluster() StateCluster {
	return s.cluster
}

func (s *baseState) K8sClient() client.Client {
	return s.cluster.K8sClient()
}

func (s *baseState) EventRecorder() record.EventRecorder {
	return s.cluster.EventRecorder()
}

func (s *baseState) Scheme() *runtime.Scheme {
	return s.cluster.Scheme()
}

func (s *baseState) Name() types.NamespacedName {
	return s.name
}

func (s *baseState) Obj() client.Object {
	return s.obj
}

func (s *baseState) SetObj(obj client.Object) {
	s.obj = obj
}

func (s *baseState) LoadObj(ctx context.Context, opts ...client.GetOption) error {
	return s.Cluster().K8sClient().Get(ctx, s.name, s.obj, opts...)
}

func (s *baseState) UpdateObj(ctx context.Context, opts ...client.UpdateOption) error {
	return s.Cluster().K8sClient().Update(ctx, s.Obj(), opts...)
}

func (s *baseState) UpdateObjStatus(ctx context.Context, opts ...client.SubResourceUpdateOption) error {
	return s.Cluster().K8sClient().Status().Update(ctx, s.Obj(), opts...)
}

func (s *baseState) PatchObjStatus(ctx context.Context) error {
	return PatchObjStatus(ctx, s.Obj(), s.Cluster().K8sClient())
}

// PatchObjAddFinalizer uses controllerutil.AddFinalizer() to add finalizer, if it returns false
// meaning that object already had that finalizer and that object is not modified it returns nil,
// if the finalizer didn't exist and object is modified, then it
// patches obj with MergePatchType. Finalizer name f must consist of alphanumeric
// characters, '-', '_' or '.', and must start and end with an alphanumeric character (e.g. 'MyName',
// or 'my.name',  or '123-abc', regex used for validation is '([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9]')"}
func (s *baseState) PatchObjAddFinalizer(ctx context.Context, f string) (bool, error) {
	return PatchObjAddFinalizer(ctx, f, s.Obj(), s.Cluster().K8sClient())
}

// PatchObjRemoveFinalizer patches obj with JSONPatchType removing the specified finalizer name
func (s *baseState) PatchObjRemoveFinalizer(ctx context.Context, f string) (bool, error) {
	return PatchObjRemoveFinalizer(ctx, f, s.Obj(), s.Cluster().K8sClient())
}

func MergePatchObj(ctx context.Context, obj client.Object, patch map[string]interface{}, clnt client.Writer) error {
	p, err := json.Marshal(patch)
	if err != nil {
		return fmt.Errorf("error json patching object when marshaling given patch: %w", err)
	}
	return clnt.Patch(ctx, obj, client.RawPatch(types.MergePatchType, p))
}

func PatchObjStatus(ctx context.Context, obj client.Object, clnt client.StatusClient) error {

	objBytes, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	var o map[string]interface{}
	if err = json.Unmarshal(objBytes, &o); err != nil {
		return err
	}

	status, ok := o["status"]

	if !ok {
		return fmt.Errorf("status not found in object %T", obj)
	}

	patch := []map[string]interface{}{
		{
			"op":    "replace",
			"path":  "/status",
			"value": status,
		},
	}

	patchBytes, err := json.Marshal(patch)
	if err != nil {
		return err
	}

	return clnt.Status().Patch(ctx, obj, client.RawPatch(types.JSONPatchType, patchBytes))
}

func PatchObjAddFinalizer(ctx context.Context, f string, obj client.Object, clnt client.Writer) (bool, error) {
	added := controllerutil.AddFinalizer(obj, f)
	if !added {
		return false, nil
	}
	p := []byte(fmt.Sprintf(`{"metadata": {"finalizers":["%s"]}}`, f))
	return true, clnt.Patch(ctx, obj, client.RawPatch(types.MergePatchType, p))
}

func PatchObjMergeAnnotation(ctx context.Context, k, v string, obj client.Object, clnt client.Writer) (bool, error) {
	if obj.GetAnnotations() != nil && obj.GetAnnotations()[k] == v {
		return false, nil
	}
	if obj.GetAnnotations() == nil {
		obj.SetAnnotations(map[string]string{})
	}
	obj.GetAnnotations()[k] = v
	p := []byte(fmt.Sprintf(`{"metadata": {"annotations":{"%s": "%s"}}}`, k, v))
	return true, clnt.Patch(ctx, obj, client.RawPatch(types.MergePatchType, p))
}

func PatchObjRemoveFinalizer(ctx context.Context, f string, obj client.Object, clnt client.Writer) (bool, error) {
	idx := -1
	for i, s := range obj.GetFinalizers() {
		if s == f {
			idx = i
			break
		}
	}
	if idx == -1 {
		return false, nil
	}
	controllerutil.RemoveFinalizer(obj, f)
	p := []byte(fmt.Sprintf(`[{"op": "remove", "path": "/metadata/finalizers/%d"}]`, idx))
	return true, clnt.Patch(ctx, obj, client.RawPatch(types.JSONPatchType, p))
}
