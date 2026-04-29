package sapnfsvolumesnapshot

import (
	"context"
	"testing"

	"github.com/gophercloud/gophercloud/v2/openstack/sharedfilesystems/v2/snapshots"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	commonscheme "github.com/kyma-project/cloud-manager/pkg/common/scheme"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	sapclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/client"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ sapclient.SnapshotClient = &snapshotClientStub{}

type snapshotClientStub struct {
	getResult  *snapshots.Snapshot
	getErr     error
	listResult []snapshots.Snapshot
	listErr    error
	listOpts   snapshots.ListOpts
}

func (s *snapshotClientStub) CreateSnapshot(_ context.Context, _ snapshots.CreateOpts) (*snapshots.Snapshot, error) {
	return nil, nil
}

func (s *snapshotClientStub) GetSnapshot(_ context.Context, _ string) (*snapshots.Snapshot, error) {
	return s.getResult, s.getErr
}

func (s *snapshotClientStub) DeleteSnapshot(_ context.Context, _ string) error {
	return nil
}

func (s *snapshotClientStub) ListSnapshots(_ context.Context, opts snapshots.ListOpts) ([]snapshots.Snapshot, error) {
	s.listOpts = opts
	return s.listResult, s.listErr
}

func (s *snapshotClientStub) RevertShareToSnapshot(_ context.Context, _ string, _ string) error {
	return nil
}

func TestSnapshotLoad(t *testing.T) {

	t.Run("snapshotLoad", func(t *testing.T) {

		var obj *cloudresourcesv1beta1.SapNfsVolumeSnapshot
		var state *State
		var client *snapshotClientStub

		setupTest := func() {
			obj = &cloudresourcesv1beta1.SapNfsVolumeSnapshot{
				ObjectMeta: metav1.ObjectMeta{Name: "snap-1", Namespace: "test"},
				Status: cloudresourcesv1beta1.SapNfsVolumeSnapshotStatus{
					Id:      "abc-123",
					ShareId: "share-789",
				},
			}

			client = &snapshotClientStub{}

			skrClient := fake.NewClientBuilder().
				WithScheme(commonscheme.SkrScheme).
				WithObjects(obj).
				WithStatusSubresource(obj).
				Build()
			skrCluster := composed.NewStateCluster(skrClient, skrClient, nil, commonscheme.SkrScheme)
			baseState := composed.NewStateFactory(skrCluster).NewState(
				types.NamespacedName{Name: obj.Name, Namespace: obj.Namespace}, obj,
			)
			state = &State{
				State:          baseState,
				SkrCluster:     skrCluster,
				snapshotClient: client,
			}
		}

		t.Run("Should: load snapshot by openstackId (fast path)", func(t *testing.T) {
			setupTest()
			ctx := t.Context()

			obj.Status.OpenstackId = "openstack-id-456"
			manilaSnapshot := &snapshots.Snapshot{ID: "openstack-id-456", Status: "available"}
			client.getResult = manilaSnapshot

			err, _ := snapshotLoad(ctx, state)

			assert.Nil(t, err, "should return nil err")
			assert.Equal(t, manilaSnapshot, state.snapshot)
		})

		t.Run("Should: set nil snapshot when openstackId not found (404)", func(t *testing.T) {
			setupTest()
			ctx := t.Context()

			obj.Status.OpenstackId = "openstack-id-456"
			client.getResult = nil

			err, _ := snapshotLoad(ctx, state)

			assert.Nil(t, err, "should return nil err")
			assert.Nil(t, state.snapshot, "snapshot should be nil")
		})

		t.Run("Should: fallback to list by name when openstackId is empty", func(t *testing.T) {
			setupTest()
			ctx := t.Context()

			foundSnapshot := snapshots.Snapshot{ID: "resolved-openstack-id", Status: "available"}
			client.listResult = []snapshots.Snapshot{foundSnapshot}

			err, _ := snapshotLoad(ctx, state)

			assert.Nil(t, err, "should return nil err")
			assert.NotNil(t, state.snapshot)
			assert.Equal(t, "resolved-openstack-id", state.snapshot.ID)
			assert.Equal(t, "cm-abc-123", client.listOpts.Name, "should use cm- prefix for OpenStack name")
			assert.Equal(t, "share-789", client.listOpts.ShareID)
		})

		t.Run("Should: persist openstackId after fallback lookup", func(t *testing.T) {
			setupTest()
			ctx := t.Context()

			foundSnapshot := snapshots.Snapshot{ID: "resolved-openstack-id", Status: "available"}
			client.listResult = []snapshots.Snapshot{foundSnapshot}

			err, _ := snapshotLoad(ctx, state)

			assert.Nil(t, err, "should return nil err")
			assert.Equal(t, "resolved-openstack-id", state.ObjAsSapNfsVolumeSnapshot().Status.OpenstackId)
		})

		t.Run("Should: set nil snapshot when list returns empty", func(t *testing.T) {
			setupTest()
			ctx := t.Context()

			client.listResult = []snapshots.Snapshot{}

			err, _ := snapshotLoad(ctx, state)

			assert.Nil(t, err, "should return nil err")
			assert.Nil(t, state.snapshot, "snapshot should be nil")
		})

		t.Run("Should: skip lookup when neither id nor shareId set", func(t *testing.T) {
			setupTest()
			ctx := t.Context()

			obj.Status.Id = ""
			obj.Status.ShareId = ""

			err, _ := snapshotLoad(ctx, state)

			assert.Nil(t, err, "should return nil err")
			assert.Nil(t, state.snapshot, "snapshot should be nil")
		})
	})
}
