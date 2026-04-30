package sapnfsvolumesnapshot

import (
	"testing"
	"time"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	commonscheme "github.com/kyma-project/cloud-manager/pkg/common/scheme"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	clocktesting "k8s.io/utils/clock/testing"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestTTLExpiry(t *testing.T) {

	t.Run("ttlExpiry", func(t *testing.T) {

		var obj *cloudresourcesv1beta1.SapNfsVolumeSnapshot
		var state *State
		var fakeClock *clocktesting.FakeClock

		setupTest := func() {
			fakeClock = clocktesting.NewFakeClock(time.Now())
			obj = &cloudresourcesv1beta1.SapNfsVolumeSnapshot{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "snap-1",
					Namespace:         "test",
					CreationTimestamp: metav1.Time{Time: time.Now().Add(-1 * time.Hour)},
				},
				Spec: cloudresourcesv1beta1.SapNfsVolumeSnapshotSpec{
					SourceVolume:    cloudresourcesv1beta1.SapNfsVolumeRef{Name: "vol-1"},
					DeleteAfterDays: 7,
				},
				Status: cloudresourcesv1beta1.SapNfsVolumeSnapshotStatus{
					State: cloudresourcesv1beta1.StateReady,
				},
			}

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
				State:      baseState,
				SkrCluster: skrCluster,
				clock:      fakeClock,
			}
		}

		t.Run("Should: continue when TTL not expired", func(t *testing.T) {
			setupTest()
			ctx := t.Context()

			// Created 1 hour ago, TTL is 7 days — not expired
			err, _ := ttlExpiry(ctx, state)

			assert.Nil(t, err, "should return nil err")
		})

		t.Run("Should: trigger deletion when TTL expired", func(t *testing.T) {
			setupTest()
			ctx := t.Context()

			obj.CreationTimestamp = metav1.Time{Time: time.Now().Add(-8 * 24 * time.Hour)}

			err, _ := ttlExpiry(ctx, state)

			assert.Equal(t, composed.StopWithRequeue, err, "should stop with requeue")
		})

		t.Run("Should: skip when deleteAfterDays is zero", func(t *testing.T) {
			setupTest()
			ctx := t.Context()

			obj.Spec.DeleteAfterDays = 0
			obj.CreationTimestamp = metav1.Time{Time: time.Now().Add(-100 * 24 * time.Hour)}

			err, _ := ttlExpiry(ctx, state)

			assert.Nil(t, err, "should return nil err")
		})

		t.Run("Should: skip when snapshot not in Ready state", func(t *testing.T) {
			setupTest()
			ctx := t.Context()

			obj.Status.State = cloudresourcesv1beta1.StateCreating
			obj.CreationTimestamp = metav1.Time{Time: time.Now().Add(-8 * 24 * time.Hour)}

			err, _ := ttlExpiry(ctx, state)

			assert.Nil(t, err, "should return nil err")
		})

		t.Run("Should: trigger deletion at exact boundary", func(t *testing.T) {
			setupTest()
			ctx := t.Context()

			// Created exactly 7 days and 1 second ago
			obj.CreationTimestamp = metav1.Time{Time: time.Now().Add(-7*24*time.Hour - time.Second)}

			err, _ := ttlExpiry(ctx, state)

			assert.Equal(t, composed.StopWithRequeue, err, "should stop with requeue")
		})
	})
}
