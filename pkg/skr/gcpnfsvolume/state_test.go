package gcpnfsvolume

import (
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"time"
)

var gcpNfsVolume = cloudresourcesv1beta1.GcpNfsVolume{
	ObjectMeta: v1.ObjectMeta{
		Name:      "test-gcp-nfs-volume",
		Namespace: "test",
	},
	Spec: cloudresourcesv1beta1.GcpNfsVolumeSpec{},
}

var deletedGcpNfsVolume = cloudresourcesv1beta1.GcpNfsVolume{
	ObjectMeta: v1.ObjectMeta{
		Name:      "deleted-gcp-nfs-volume",
		Namespace: "test",
		DeletionTimestamp: &v1.Time{
			Time: time.Now(),
		},
		Finalizers: []string{"test-finalizer"},
	},
	Spec: cloudresourcesv1beta1.GcpNfsVolumeSpec{},
}

type testStateFactory struct {
	factory             StateFactory
	skrCluster          composed.StateCluster
	gcpNfsVolume        *cloudresourcesv1beta1.GcpNfsVolume
	deletedGcpNfsVolume *cloudresourcesv1beta1.GcpNfsVolume
}

func newTestStateFactory() (*testStateFactory, error) {

	skrScheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(skrScheme))
	utilruntime.Must(cloudresourcesv1beta1.AddToScheme(skrScheme))

	skrClient := fake.NewClientBuilder().
		WithScheme(skrScheme).
		WithObjects(&gcpNfsVolume).
		Build()
	skrCluster := composed.NewStateCluster(skrClient, nil, nil)

	factory := NewStateFactory(klog.ObjectRef{}, nil, skrCluster)

	return &testStateFactory{
		factory:             factory,
		skrCluster:          skrCluster,
		gcpNfsVolume:        &gcpNfsVolume,
		deletedGcpNfsVolume: &deletedGcpNfsVolume,
	}, nil

}

func (f *testStateFactory) newState() *State {
	return f.factory.NewState(composed.NewStateFactory(f.skrCluster).NewState(
		types.NamespacedName{
			Name:      f.gcpNfsVolume.Name,
			Namespace: f.gcpNfsVolume.Namespace,
		}, f.gcpNfsVolume))
}

func (f *testStateFactory) newStateWithDeleted() *State {
	return f.factory.NewState(composed.NewStateFactory(f.skrCluster).NewState(
		types.NamespacedName{
			Name:      f.deletedGcpNfsVolume.Name,
			Namespace: f.deletedGcpNfsVolume.Namespace,
		}, f.deletedGcpNfsVolume))
}
