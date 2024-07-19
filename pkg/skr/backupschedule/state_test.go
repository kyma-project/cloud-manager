package backupschedule

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	ObjectMeta: v1.ObjectMeta{
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

var awsScope = cloudcontrolv1beta1.Scope{
	ObjectMeta: v1.ObjectMeta{
		Name:      "skr",
		Namespace: "test",
	},
	Spec: cloudcontrolv1beta1.ScopeSpec{
		Provider: "aws",
		Scope: cloudcontrolv1beta1.ScopeInfo{
			Aws: &cloudcontrolv1beta1.AwsScope{
				VpcNetwork: "aws-nw",
				Network:    cloudcontrolv1beta1.AwsNetwork{},
				AccountId:  "aws-account-id",
			},
		},
	},
}

var gcpNfsVolume = cloudresourcesv1beta1.GcpNfsVolume{
	ObjectMeta: v1.ObjectMeta{
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
		Conditions: []v1.Condition{
			{
				Type:               "Ready",
				Status:             "True",
				LastTransitionTime: v1.Time{Time: time.Now()},
				Reason:             "Ready",
				Message:            "NFS instance is ready",
			},
		},
	},
}

var awsNfsVolume = cloudresourcesv1beta1.AwsNfsVolume{
	ObjectMeta: v1.ObjectMeta{
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
		Conditions: []v1.Condition{
			{
				Type:               "Ready",
				Status:             "True",
				LastTransitionTime: v1.Time{Time: time.Now()},
				Reason:             "Ready",
				Message:            "NFS instance is ready",
			},
		},
	},
}

var gcpNfsBackupSchedule = cloudresourcesv1beta1.GcpNfsBackupSchedule{
	ObjectMeta: v1.ObjectMeta{
		Name:      "test-nfs-backup-backupschedule",
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

var deletingGcpBackupSchedule = cloudresourcesv1beta1.GcpNfsBackupSchedule{
	ObjectMeta: v1.ObjectMeta{
		Name:              "test-nfs-backup-backupschedule-01",
		Namespace:         "test",
		DeletionTimestamp: &v1.Time{Time: time.Now()},
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
