package fake

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/stretchr/testify/assert"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
)

var scheme = runtime.NewScheme()

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(cloudcontrolv1beta1.AddToScheme(scheme))
}

func mustNewFakeWithData(t *testing.T) K8sFakePort {
	f, err := newFakeWithData()
	assert.NoError(t, err)
	return f
}

func newFakeWithData() (K8sFakePort, error) {
	f := NewFakeK8sPortOnDefaultCluster(scheme)
	err := f.Set(
		&cloudcontrolv1beta1.Scope{ObjectMeta: metav1.ObjectMeta{
			Namespace: "nsA",
			Name:      "a",
			Labels: map[string]string{
				"group": "x",
			},
		}},
		&cloudcontrolv1beta1.Scope{ObjectMeta: metav1.ObjectMeta{
			Namespace: "nsB",
			Name:      "b",
			Labels: map[string]string{
				"group": "y",
			},
		}},
		&cloudcontrolv1beta1.Scope{ObjectMeta: metav1.ObjectMeta{
			Namespace: "nsA",
			Name:      "c",
			Labels: map[string]string{
				"group": "x",
			},
		}},
	)
	return f, err
}

func TestFakeK8sPort(t *testing.T) {

	t.Run("list", func(t *testing.T) {
		t.Run("list_all", func(t *testing.T) {
			f := mustNewFakeWithData(t)
			list := &cloudcontrolv1beta1.ScopeList{}
			assert.NoError(t, f.List(context.Background(), list))
			assert.Len(t, list.Items, 3)
		})

		t.Run("list_in_namespace", func(t *testing.T) {
			f := mustNewFakeWithData(t)
			list := &cloudcontrolv1beta1.ScopeList{}
			assert.NoError(t, f.List(context.Background(), list, client.InNamespace("nsA")))
			assert.Len(t, list.Items, 2)
			assert.NoError(t, f.List(context.Background(), list, client.InNamespace("nsB")))
			assert.Len(t, list.Items, 1)
		})

		t.Run("list_matching_labels", func(t *testing.T) {
			f := mustNewFakeWithData(t)
			list := &cloudcontrolv1beta1.ScopeList{}
			assert.NoError(t, f.List(context.Background(), list, client.MatchingLabels{"group": "x"}))
			assert.Len(t, list.Items, 2)
			assert.NoError(t, f.List(context.Background(), list, client.MatchingLabels{"group": "y"}))
			assert.Len(t, list.Items, 1)
		})

		t.Run("list_limit", func(t *testing.T) {
			f := mustNewFakeWithData(t)
			list := &cloudcontrolv1beta1.ScopeList{}
			assert.NoError(t, f.List(context.Background(), list, client.Limit(1)))
			assert.Len(t, list.Items, 1)
			assert.NoError(t, f.List(context.Background(), list, client.Limit(2)))
			assert.Len(t, list.Items, 2)
			assert.NoError(t, f.List(context.Background(), list, client.Limit(3)))
			assert.Len(t, list.Items, 3)
		})
	})

	t.Run("create", func(t *testing.T) {
		t.Run("create_new", func(t *testing.T) {
			f := mustNewFakeWithData(t)
			assert.NoError(t, f.Create(context.Background(), &cloudcontrolv1beta1.Scope{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "nsC",
					Name:      "just-created",
					Labels: map[string]string{
						"just-created": "true",
					},
				},
			}))
			obj := &cloudcontrolv1beta1.Scope{}
			assert.NoError(t, f.LoadObj(context.Background(), types.NamespacedName{Namespace: "nsC", Name: "just-created"}, obj))
			assert.Equal(t, "true", obj.Labels["just-created"])
		})

		t.Run("create_existing", func(t *testing.T) {
			f := mustNewFakeWithData(t)
			err := f.Create(context.Background(), &cloudcontrolv1beta1.Scope{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "nsA",
					Name:      "a",
				},
			})
			assert.True(t, apierrors.IsAlreadyExists(err))
		})
	})

	t.Run("delete", func(t *testing.T) {
		t.Run("delete_existing", func(t *testing.T) {
			f := mustNewFakeWithData(t)
			assert.NoError(t, f.Delete(context.Background(), &cloudcontrolv1beta1.Scope{
				ObjectMeta: metav1.ObjectMeta{Namespace: "nsA", Name: "a"},
			}))
			x := &cloudcontrolv1beta1.Scope{}
			err := f.LoadObj(context.Background(), types.NamespacedName{Namespace: "nsA", Name: "a"}, x)
			assert.True(t, apierrors.IsNotFound(err))
		})

		t.Run("delete_non_existing", func(t *testing.T) {
			f := mustNewFakeWithData(t)
			err := f.Delete(context.Background(), &cloudcontrolv1beta1.Scope{
				ObjectMeta: metav1.ObjectMeta{Namespace: "non-existing", Name: "non-existing"},
			})
			assert.True(t, apierrors.IsNotFound(err))
		})
	})

	t.Run("patch_merge_labels", func(t *testing.T) {
		t.Run("add_label", func(t *testing.T) {
			f := mustNewFakeWithData(t)
			obj := &cloudcontrolv1beta1.Scope{ObjectMeta: metav1.ObjectMeta{Namespace: "nsA", Name: "a"}}
			changed, err := f.PatchMergeLabels(
				context.Background(), obj,
				map[string]string{"new-label": "new-value"},
			)
			assert.NoError(t, err)
			assert.True(t, changed)

			obj = &cloudcontrolv1beta1.Scope{ObjectMeta: metav1.ObjectMeta{Namespace: "nsA", Name: "a"}}
			assert.NoError(t, f.LoadObj(context.Background(), client.ObjectKeyFromObject(obj), obj))
			assert.Equal(t, "new-value", obj.Labels["new-label"])
		})

		t.Run("existing_label", func(t *testing.T) {
			f := mustNewFakeWithData(t)
			obj := &cloudcontrolv1beta1.Scope{ObjectMeta: metav1.ObjectMeta{Namespace: "nsA", Name: "a"}}
			changed, err := f.PatchMergeLabels(
				context.Background(), obj,
				map[string]string{"group": "x"},
			)
			assert.NoError(t, err)
			assert.False(t, changed)

			obj = &cloudcontrolv1beta1.Scope{ObjectMeta: metav1.ObjectMeta{Namespace: "nsA", Name: "a"}}
			assert.NoError(t, f.LoadObj(context.Background(), client.ObjectKeyFromObject(obj), obj))
			assert.Len(t, obj.Labels, 1)
		})
	})

	t.Run("patch_merge_annotations", func(t *testing.T) {
		t.Run("add_annotation", func(t *testing.T) {
			f := mustNewFakeWithData(t)
			obj := &cloudcontrolv1beta1.Scope{ObjectMeta: metav1.ObjectMeta{Namespace: "nsA", Name: "a"}}
			changed, err := f.PatchMergeAnnotations(
				context.Background(), obj,
				map[string]string{"new-annot": "new-value"},
			)
			assert.NoError(t, err)
			assert.True(t, changed)

			obj = &cloudcontrolv1beta1.Scope{ObjectMeta: metav1.ObjectMeta{Namespace: "nsA", Name: "a"}}
			assert.NoError(t, f.LoadObj(context.Background(), client.ObjectKeyFromObject(obj), obj))
			assert.Equal(t, "new-value", obj.Annotations["new-annot"])
		})
	})
}
