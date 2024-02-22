package gcpnfsvolume

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
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

var kymaRef = klog.ObjectRef{
	Name:      "skr",
	Namespace: "test",
}

var gcpNfsVolume = cloudresourcesv1beta1.GcpNfsVolume{
	ObjectMeta: v1.ObjectMeta{
		Name:      "test-gcp-nfs-volume",
		Namespace: "test",
	},
	Spec: cloudresourcesv1beta1.GcpNfsVolumeSpec{
		IpRange: cloudresourcesv1beta1.IpRangeRef{
			Name:      "test-gcp-ip-range",
			Namespace: "test",
		},
		Location:      "us-west1",
		Tier:          "BASIC_HDD",
		FileShareName: "vol1",
		CapacityGb:    1024,
	},
	Status: cloudresourcesv1beta1.GcpNfsVolumeStatus{
		Id:         "test-gcp-nfs-instance",
		Hosts:      []string{"10.0.0.2"},
		CapacityGb: 1024,
		Conditions: []v1.Condition{
			{
				Type:               "Ready",
				Status:             "True",
				LastTransitionTime: v1.Time{Time: time.Now()},
				Reason:             "Ready",
				Message:            "NFS is instance is ready",
			},
		},
	},
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

var gcpNfsInstance = cloudcontrolv1beta1.NfsInstance{
	ObjectMeta: v1.ObjectMeta{
		Name:      gcpNfsVolume.Status.Id,
		Namespace: kymaRef.Namespace,
		Labels: map[string]string{
			cloudcontrolv1beta1.LabelKymaName:        kymaRef.Name,
			cloudcontrolv1beta1.LabelRemoteName:      gcpNfsVolume.Name,
			cloudcontrolv1beta1.LabelRemoteNamespace: gcpNfsVolume.Namespace,
		},
	},
	Spec: cloudcontrolv1beta1.NfsInstanceSpec{
		RemoteRef: cloudcontrolv1beta1.RemoteRef{
			Namespace: gcpNfsVolume.Namespace,
			Name:      gcpNfsVolume.Name,
		},
		IpRange: cloudcontrolv1beta1.IpRangeRef{
			Name: gcpNfsVolume.Spec.IpRange.Name,
		},
		Scope: cloudcontrolv1beta1.ScopeRef{
			Name: kymaRef.Name,
		},
		Instance: cloudcontrolv1beta1.NfsInstanceInfo{
			Gcp: &cloudcontrolv1beta1.NfsInstanceGcp{
				Location:      gcpNfsVolume.Spec.Location,
				Tier:          cloudcontrolv1beta1.GcpFileTier(gcpNfsVolume.Spec.Tier),
				FileShareName: gcpNfsVolume.Spec.FileShareName,
				CapacityGb:    gcpNfsVolume.Spec.CapacityGb,
				ConnectMode:   cloudcontrolv1beta1.PRIVATE_SERVICE_ACCESS,
			},
		},
	},
	Status: cloudcontrolv1beta1.NfsInstanceStatus{
		State: "Ready",
		Id:    "gcp-filestore-1",
		Conditions: []v1.Condition{
			{
				Type:               "Ready",
				Status:             "True",
				LastTransitionTime: v1.Time{Time: time.Now()},
				Reason:             "Ready",
				Message:            "NFS is instance is ready",
			},
		},
		Hosts:      []string{"10.0.0.2"},
		CapacityGb: gcpNfsVolume.Spec.CapacityGb,
	},
}

var gcpNfsInstance2 = cloudcontrolv1beta1.NfsInstance{
	ObjectMeta: v1.ObjectMeta{
		Name:      "test-gcp-nfs-instance-2",
		Namespace: kymaRef.Namespace,
		Labels: map[string]string{
			cloudcontrolv1beta1.LabelKymaName:        kymaRef.Name,
			cloudcontrolv1beta1.LabelRemoteName:      "test-gcp-nfs-volume-2",
			cloudcontrolv1beta1.LabelRemoteNamespace: "test",
		},
	},
	Spec: cloudcontrolv1beta1.NfsInstanceSpec{
		RemoteRef: cloudcontrolv1beta1.RemoteRef{
			Namespace: "test",
			Name:      "test-gcp-nfs-volume-2",
		},
		IpRange: cloudcontrolv1beta1.IpRangeRef{
			Name: "test-gcp-ip-range",
		},
		Scope: cloudcontrolv1beta1.ScopeRef{
			Name: kymaRef.Name,
		},
		Instance: cloudcontrolv1beta1.NfsInstanceInfo{
			Gcp: &cloudcontrolv1beta1.NfsInstanceGcp{
				Location:      gcpNfsVolume.Spec.Location,
				Tier:          cloudcontrolv1beta1.GcpFileTier(gcpNfsVolume.Spec.Tier),
				FileShareName: gcpNfsVolume.Spec.FileShareName,
				CapacityGb:    gcpNfsVolume.Spec.CapacityGb,
				ConnectMode:   cloudcontrolv1beta1.PRIVATE_SERVICE_ACCESS,
			},
		},
	},
}

type testStateFactory struct {
	factory             StateFactory
	skrCluster          composed.StateCluster
	kcpCluster          composed.StateCluster
	gcpNfsVolume        *cloudresourcesv1beta1.GcpNfsVolume
	deletedGcpNfsVolume *cloudresourcesv1beta1.GcpNfsVolume
}

func newTestStateFactory() (*testStateFactory, error) {

	kcpScheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(kcpScheme))
	utilruntime.Must(cloudcontrolv1beta1.AddToScheme(kcpScheme))

	kcpClient := fake.NewClientBuilder().
		WithScheme(kcpScheme).
		WithObjects(&gcpNfsInstance).
		WithObjects(&gcpNfsInstance2).
		Build()
	kcpCluster := composed.NewStateCluster(kcpClient, nil, kcpScheme)

	skrScheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(skrScheme))
	utilruntime.Must(cloudresourcesv1beta1.AddToScheme(skrScheme))

	skrClient := fake.NewClientBuilder().
		WithScheme(skrScheme).
		WithObjects(&gcpNfsVolume).
		WithObjects(&deletedGcpNfsVolume).
		Build()
	skrCluster := composed.NewStateCluster(skrClient, nil, skrScheme)

	factory := NewStateFactory(kymaRef, kcpCluster, skrCluster)

	return &testStateFactory{
		factory:             factory,
		skrCluster:          skrCluster,
		kcpCluster:          kcpCluster,
		gcpNfsVolume:        &gcpNfsVolume,
		deletedGcpNfsVolume: &deletedGcpNfsVolume,
	}, nil

}

func (f *testStateFactory) newState() *State {
	return f.newStateWith(&gcpNfsVolume)
}

func (f *testStateFactory) newStateWith(nfsVolume *cloudresourcesv1beta1.GcpNfsVolume) *State {
	return f.factory.NewState(composed.NewStateFactory(f.skrCluster).NewState(
		types.NamespacedName{
			Name:      nfsVolume.Name,
			Namespace: nfsVolume.Namespace,
		}, nfsVolume))
}
