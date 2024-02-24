package nfsinstance

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	nfsTypes "github.com/kyma-project/cloud-manager/pkg/kcp/nfsinstance/types"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	client2 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsinstance/client"
	"google.golang.org/api/file/v1"
	"google.golang.org/api/option"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2"
	"net/http"
	"net/http/httptest"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"time"
)

var kymaRef = klog.ObjectRef{
	Name:      "skr",
	Namespace: "test",
}

func getDeletedGcpNfsInstance() *cloudcontrolv1beta1.NfsInstance {
	return &cloudcontrolv1beta1.NfsInstance{
		ObjectMeta: v1.ObjectMeta{
			Name:      "deleted-gcp-nfs-instance",
			Namespace: "kcp-system",
			DeletionTimestamp: &v1.Time{
				Time: time.Now(),
			},
			Finalizers: []string{"test-finalizer"},
		},
		Spec: cloudcontrolv1beta1.NfsInstanceSpec{},
	}
}

func getGcpNfsInstance() *cloudcontrolv1beta1.NfsInstance {
	return &cloudcontrolv1beta1.NfsInstance{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test-gcp-nfs-instance",
			Namespace: kymaRef.Namespace,
			Labels: map[string]string{
				cloudcontrolv1beta1.LabelKymaName:        kymaRef.Name,
				cloudcontrolv1beta1.LabelRemoteName:      "test-gcp-nfs-volume",
				cloudcontrolv1beta1.LabelRemoteNamespace: "test",
			},
		},
		Spec: cloudcontrolv1beta1.NfsInstanceSpec{
			RemoteRef: cloudcontrolv1beta1.RemoteRef{
				Namespace: "test",
				Name:      "test-gcp-nfs-volume",
			},
			IpRange: cloudcontrolv1beta1.IpRangeRef{
				Name: "test-gcp-ip-range",
			},
			Scope: cloudcontrolv1beta1.ScopeRef{
				Name: kymaRef.Name,
			},
			Instance: cloudcontrolv1beta1.NfsInstanceInfo{
				Gcp: &cloudcontrolv1beta1.NfsInstanceGcp{
					Location:      "us-west1",
					Tier:          "BASIC_HDD",
					FileShareName: "vol1",
					CapacityGb:    1024,
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
			CapacityGb: 1024,
		},
	}
}

func getInstanceWithoutStatus() *cloudcontrolv1beta1.NfsInstance {
	var gcpNfsInstance2 = getGcpNfsInstance().DeepCopy()
	gcpNfsInstance2.Name = "test-gcp-nfs-instance-2"
	gcpNfsInstance2.Labels[cloudcontrolv1beta1.LabelRemoteName] = "test-gcp-nfs-volume-2"
	gcpNfsInstance2.Spec.RemoteRef.Name = "test-gcp-nfs-volume-2"
	gcpNfsInstance2.Status = cloudcontrolv1beta1.NfsInstanceStatus{}
	return gcpNfsInstance2
}

type testStateFactory struct {
	factory                StateFactory
	kcpCluster             composed.StateCluster
	nfsInstance            *cloudcontrolv1beta1.NfsInstance
	nfsInstanceNoCondition *cloudcontrolv1beta1.NfsInstance
	deletedNfsInstance     *cloudcontrolv1beta1.NfsInstance
	fakeHttpServer         *httptest.Server
}

func newTestStateFactory() (*testStateFactory, error) {

	kcpScheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(kcpScheme))
	utilruntime.Must(cloudcontrolv1beta1.AddToScheme(kcpScheme))

	kcpClient := fake.NewClientBuilder().
		WithScheme(kcpScheme).
		WithObjects(getGcpNfsInstance()).
		WithObjects(getInstanceWithoutStatus()).
		WithObjects(getDeletedGcpNfsInstance()).
		Build()
	kcpCluster := composed.NewStateCluster(kcpClient, nil, kcpScheme)
	fakeFileStoreClientProvider, fakeHttpServer := NewFakeFilestoreClientProvider()
	factory := NewStateFactory(fakeFileStoreClientProvider, abstractions.NewMockedEnvironment(map[string]string{"GCP_SA_JSON_KEY_PATH": "test"}))

	return &testStateFactory{
		factory:                factory,
		kcpCluster:             kcpCluster,
		nfsInstance:            getGcpNfsInstance(),
		nfsInstanceNoCondition: getInstanceWithoutStatus(),
		deletedNfsInstance:     getDeletedGcpNfsInstance(),
		fakeHttpServer:         fakeHttpServer,
	}, nil

}

type TestState struct {
	*State
	FakeHttpServer *httptest.Server
}

func (f *testStateFactory) newState(ctx context.Context) (*TestState, error) {
	return f.newStateWith(ctx, getGcpNfsInstance())
}

func (f *testStateFactory) newStateWith(ctx context.Context, nfsInstance *cloudcontrolv1beta1.NfsInstance) (*TestState, error) {
	focalState := focal.NewStateFactory().NewState(composed.NewStateFactory(f.kcpCluster).NewState(
		types.NamespacedName{
			Name:      nfsInstance.Name,
			Namespace: nfsInstance.Namespace,
		}, nfsInstance))
	typesState := newTypesState(focalState)

	state, err := f.factory.NewState(ctx, typesState)
	if err != nil {
		return nil, err
	} else {
		return &TestState{
			State:          state,
			FakeHttpServer: f.fakeHttpServer,
		}, nil
	}
}

func NewFakeFilestoreClientProvider() (client.ClientProvider[client2.FilestoreClient], *httptest.Server) {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var opResp *file.Operation
		switch r.Method {
		case http.MethodGet:
			if r.URL.Path == "/projects/test-project/locations/us-west1/instances/test-gcp-nfs-volume" {

			}
			if r.URL.Path == "/projects/test-project/locations/us-west1/instances/test-gcp-nfs-volume-2" {

			}
			//Check if url path is for operation
			if r.URL.Path == "/projects/test-project/locations/us-west1/operations/create-operation" {
				// check if operation name is create-operation
			}

		case http.MethodPost:
			opResp = &file.Operation{
				Name: "create-operation",
				Done: false,
			}
		case http.MethodPatch:
			opResp = &file.Operation{
				Name: "patch-operation",
				Done: false,
			}
		case http.MethodDelete:
			opResp = &file.Operation{
				Name: "delete-operation",
				Done: false,
			}
		}
		b, err := json.Marshal(opResp)
		if err != nil {
			http.Error(w, "unable to marshal request: "+err.Error(), http.StatusBadRequest)
			return
		}
		_, err = w.Write(b)
		if err != nil {
			http.Error(w, "unable to write to provided ResponseWriter: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}))
	return client.NewCachedClientProvider(
		func(ctx context.Context, saJsonKeyPath string) (client2.FilestoreClient, error) {
			fsClient, err := file.NewService(ctx, option.WithoutAuthentication(), option.WithEndpoint(fakeHttpServer.URL))
			if err != nil {
				return nil, err
			}
			return client2.NewFilestoreClient(fsClient), nil
		},
	), fakeHttpServer
}

type typesState struct {
	focal.State

	ipRange *cloudcontrolv1beta1.IpRange
}

func (s *typesState) ObjAsNfsInstance() *cloudcontrolv1beta1.NfsInstance {
	return s.Obj().(*cloudcontrolv1beta1.NfsInstance)
}

func (s *typesState) IpRange() *cloudcontrolv1beta1.IpRange {
	return s.ipRange
}

func (s *typesState) SetIpRange(r *cloudcontrolv1beta1.IpRange) {
	s.ipRange = r
}

func newTypesState(focalState focal.State) nfsTypes.State {
	return &typesState{State: focalState}
}
