package backupschedule

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"time"
)

var kymaRef = klog.ObjectRef{
	Name:      "skr",
	Namespace: "test",
}

var gcpScope = cloudcontrolv1beta1.Scope{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "skr",
		Namespace: "test",
	},
	Spec: cloudcontrolv1beta1.ScopeSpec{
		Provider: "gcp",
		Scope: cloudcontrolv1beta1.ScopeInfo{
			Gcp: &cloudcontrolv1beta1.GcpScope{
				Project:    "test-project",
				VpcNetwork: "test-network",
			},
		},
	},
}

var gcpNfsVolume = cloudresourcesv1beta1.GcpNfsVolume{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "test-gcp-nfs-volume",
		Namespace: "test",
	},
	Spec: cloudresourcesv1beta1.GcpNfsVolumeSpec{
		IpRange: cloudresourcesv1beta1.IpRangeRef{
			Name: "test-gcp-ip-range",
		},
		Location:      "us-west1",
		Tier:          "BASIC_HDD",
		FileShareName: "vol1",
		CapacityGb:    1024,
	},
	Status: cloudresourcesv1beta1.GcpNfsVolumeStatus{
		Id:         "test-gcp-nfs-instance",
		Hosts:      []string{"10.20.30.2"},
		CapacityGb: 1024,
		Conditions: []metav1.Condition{
			{
				Type:               "Ready",
				Status:             "True",
				LastTransitionTime: metav1.Time{Time: time.Now()},
				Reason:             "Ready",
				Message:            "NFS instance is ready",
			},
		},
	},
}

var awsNfsVolume = cloudresourcesv1beta1.AwsNfsVolume{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "test-aws-nfs-volume",
		Namespace: "test",
	},
	Spec: cloudresourcesv1beta1.AwsNfsVolumeSpec{
		IpRange: cloudresourcesv1beta1.IpRangeRef{
			Name: "test-ip-range",
		},
	},
	Status: cloudresourcesv1beta1.AwsNfsVolumeStatus{
		Id: "test-aws-nfs-instance",
		Conditions: []metav1.Condition{
			{
				Type:               "Ready",
				Status:             "True",
				LastTransitionTime: metav1.Time{Time: time.Now()},
				Reason:             "Ready",
				Message:            "NFS instance is ready",
			},
		},
	},
}

var gcpNfsBackupSchedule = cloudresourcesv1beta1.GcpNfsBackupSchedule{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "test-nfs-backup-schedule",
		Namespace: "test",
	},
	Spec: cloudresourcesv1beta1.GcpNfsBackupScheduleSpec{
		NfsVolumeRef: corev1.ObjectReference{
			Name:      "test-gcp-nfs-volume",
			Namespace: "test",
		},
		Location: "us-west1",
	},
}

var backup1Meta = metav1.ObjectMeta{
	Name:              "test-backup-1",
	Namespace:         gcpNfsBackupSchedule.Namespace,
	CreationTimestamp: metav1.Time{Time: time.Now()},
	Labels: map[string]string{
		cloudresourcesv1beta1.LabelScheduleName:      gcpNfsBackupSchedule.Name,
		cloudresourcesv1beta1.LabelScheduleNamespace: gcpNfsBackupSchedule.Namespace,
	},
}
var backup2Meta = metav1.ObjectMeta{
	Name:              "test-backup-2",
	Namespace:         gcpNfsBackupSchedule.Namespace,
	CreationTimestamp: metav1.Time{Time: time.Now().AddDate(0, 0, -2)},
	Labels: map[string]string{
		cloudresourcesv1beta1.LabelScheduleName:      gcpNfsBackupSchedule.Name,
		cloudresourcesv1beta1.LabelScheduleNamespace: gcpNfsBackupSchedule.Namespace,
	},
}
var gcpSpec = cloudresourcesv1beta1.GcpNfsVolumeBackupSpec{
	Source: cloudresourcesv1beta1.GcpNfsVolumeBackupSource{
		Volume: cloudresourcesv1beta1.GcpNfsVolumeRef{
			Name:      gcpNfsVolume.Name,
			Namespace: gcpNfsVolume.Namespace,
		},
	},
	Location: "us-west1-a",
}

var gcpBackup1 = &cloudresourcesv1beta1.GcpNfsVolumeBackup{
	ObjectMeta: backup1Meta,
	Spec:       gcpSpec,
}
var gcpBackup2 = &cloudresourcesv1beta1.GcpNfsVolumeBackup{
	ObjectMeta: backup2Meta,
	Spec:       gcpSpec,
}

var deletingGcpBackupSchedule = cloudresourcesv1beta1.GcpNfsBackupSchedule{
	ObjectMeta: metav1.ObjectMeta{
		Name:              "test-nfs-backup-schedule",
		Namespace:         "test",
		DeletionTimestamp: &metav1.Time{Time: time.Now()},
		Finalizers:        []string{cloudresourcesv1beta1.Finalizer},
	},
	Spec: cloudresourcesv1beta1.GcpNfsBackupScheduleSpec{
		NfsVolumeRef: corev1.ObjectReference{
			Name:      "test-gcp-nfs-volume",
			Namespace: "test",
		},
		Location: "us-west1",
	},
}

type testStateFactory struct {
	factory    StateFactory
	skrCluster composed.StateCluster
	kcpCluster composed.StateCluster
	env        abstractions.Environment
	gcpConfig  *gcpclient.GcpConfig
}

func newTestStateFactoryWithObj(object client.Object) (*testStateFactory, error) {

	kcpScheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(kcpScheme))
	utilruntime.Must(cloudcontrolv1beta1.AddToScheme(kcpScheme))

	kcpClient := fake.NewClientBuilder().
		WithScheme(kcpScheme).
		WithObjects(&gcpScope).
		Build()
	kcpCluster := composed.NewStateCluster(kcpClient, kcpClient, nil, kcpScheme)

	skrScheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(skrScheme))
	utilruntime.Must(cloudresourcesv1beta1.AddToScheme(skrScheme))

	skrClient := fake.NewClientBuilder().
		WithScheme(skrScheme).
		WithObjects(&gcpNfsVolume).
		WithStatusSubresource(&gcpNfsVolume).
		WithObjects(&awsNfsVolume).
		WithStatusSubresource(&awsNfsVolume).
		WithObjects(object).
		WithStatusSubresource(object).
		Build()
	skrCluster := composed.NewStateCluster(skrClient, skrClient, nil, skrScheme)
	env := abstractions.NewMockedEnvironment(map[string]string{"GCP_SA_JSON_KEY_PATH": "test"})
	factory := NewStateFactory(kymaRef, kcpCluster, skrCluster, env)

	return &testStateFactory{
		factory:    factory,
		skrCluster: skrCluster,
		kcpCluster: kcpCluster,
		env:        env,
		gcpConfig:  gcpclient.GetGcpConfig(env),
	}, nil

}

func (f *testStateFactory) newStateWith(schedule *cloudresourcesv1beta1.GcpNfsBackupSchedule) (*State, error) {
	return f.factory.NewState(context.Background(), composed.NewStateFactory(f.skrCluster).NewState(
		types.NamespacedName{
			Name:      schedule.Name,
			Namespace: schedule.Namespace,
		}, schedule))
}

// Fake client doesn't support type "apply" for patching so falling back on update for unit tests.
func (s *State) PatchObjStatus(ctx context.Context) error {
	return s.Cluster().K8sClient().Status().Update(ctx, s.Obj())
}
