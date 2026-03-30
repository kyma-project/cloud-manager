package subnet

import (
	"context"
	"testing"

	"cloud.google.com/go/compute/apiv1/computepb"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	commonscheme "github.com/kyma-project/cloud-manager/pkg/common/scheme"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func newTestState(gcpSubnet *cloudcontrolv1beta1.GcpSubnet) *State {
	k8sClient := fake.NewClientBuilder().
		WithScheme(commonscheme.KcpScheme).
		WithObjects(gcpSubnet).
		WithStatusSubresource(gcpSubnet).
		Build()

	cluster := composed.NewStateCluster(k8sClient, k8sClient, nil, commonscheme.KcpScheme)
	composedState := composed.NewStateFactory(cluster).NewState(
		types.NamespacedName{
			Name:      gcpSubnet.Name,
			Namespace: gcpSubnet.Namespace,
		},
		gcpSubnet,
	)
	focalState := focal.NewStateFactory().NewState(composedState)

	return &State{
		State: focalState,
	}
}

func newGcpSubnet() *cloudcontrolv1beta1.GcpSubnet {
	return &cloudcontrolv1beta1.GcpSubnet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-subnet",
			Namespace: "kcp-system",
		},
	}
}

func TestSetErrorIfOperationErrors(t *testing.T) {
	t.Run("Should return nil when subnetCreationOperation is nil", func(t *testing.T) {
		state := newTestState(newGcpSubnet())
		state.subnetCreationOperation = nil

		err, _ := setErrorIfOperationErrors(context.Background(), state)

		assert.Nil(t, err)
	})

	t.Run("Should return nil when operation has no error", func(t *testing.T) {
		state := newTestState(newGcpSubnet())
		state.subnetCreationOperation = &computepb.Operation{
			Status: computepb.Operation_DONE.Enum(),
			Error:  nil,
		}

		err, _ := setErrorIfOperationErrors(context.Background(), state)

		assert.Nil(t, err)
	})

	t.Run("Should set error condition when operation has errors", func(t *testing.T) {
		gcpSubnet := newGcpSubnet()
		state := newTestState(gcpSubnet)
		state.subnetCreationOperation = &computepb.Operation{
			Status: computepb.Operation_DONE.Enum(),
			Error: &computepb.Error{
				Errors: []*computepb.Errors{
					{
						Code:    proto.String("QUOTA_EXCEEDED"),
						Message: proto.String("Subnet quota exceeded"),
					},
				},
			},
		}

		err, _ := setErrorIfOperationErrors(context.Background(), state)

		require.Equal(t, composed.StopAndForget, err)
		cond := meta.FindStatusCondition(gcpSubnet.Status.Conditions, cloudcontrolv1beta1.ConditionTypeError)
		require.NotNil(t, cond)
		assert.Equal(t, "True", string(cond.Status))
		assert.Equal(t, cloudcontrolv1beta1.ReasonCloudProviderError, cond.Reason)
		assert.Equal(t, cloudcontrolv1beta1.StateError, gcpSubnet.Status.State)
	})

	t.Run("Should return StopAndForget when error condition already exists", func(t *testing.T) {
		gcpSubnet := newGcpSubnet()
		meta.SetStatusCondition(&gcpSubnet.Status.Conditions, metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  cloudcontrolv1beta1.ReasonCloudProviderError,
			Message: "KCP GcpSubnet creation operation failed",
		})
		state := newTestState(gcpSubnet)
		state.subnetCreationOperation = &computepb.Operation{
			Status: computepb.Operation_DONE.Enum(),
			Error: &computepb.Error{
				Errors: []*computepb.Errors{
					{
						Code:    proto.String("QUOTA_EXCEEDED"),
						Message: proto.String("Subnet quota exceeded"),
					},
				},
			},
		}

		err, _ := setErrorIfOperationErrors(context.Background(), state)

		assert.Equal(t, composed.StopAndForget, err)
	})
}
