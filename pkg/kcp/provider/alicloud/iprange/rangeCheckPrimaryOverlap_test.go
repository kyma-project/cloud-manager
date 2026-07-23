package iprange

import (
	"testing"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	commonscheme "github.com/kyma-project/cloud-manager/pkg/common/scheme"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

type iprangeTestState struct {
	focal.State
}

func (s *iprangeTestState) ObjAsIpRange() *cloudcontrolv1beta1.IpRange {
	return s.Obj().(*cloudcontrolv1beta1.IpRange)
}
func (s *iprangeTestState) ExistingCidrRanges() []string                     { return nil }
func (s *iprangeTestState) SetExistingCidrRanges(_ []string)                 {}
func (s *iprangeTestState) Network() *cloudcontrolv1beta1.Network            { return nil }
func (s *iprangeTestState) SetNetwork(_ *cloudcontrolv1beta1.Network)        {}
func (s *iprangeTestState) NetworkKey() client.ObjectKey                     { return client.ObjectKey{} }
func (s *iprangeTestState) SetNetworkKey(_ client.ObjectKey)                 {}
func (s *iprangeTestState) IsCloudManagerNetwork() bool                      { return false }
func (s *iprangeTestState) SetIsCloudManagerNetwork(_ bool)                  {}
func (s *iprangeTestState) IsKymaNetwork() bool                              { return false }
func (s *iprangeTestState) SetIsKymaNetwork(_ bool)                          {}
func (s *iprangeTestState) KymaNetwork() *cloudcontrolv1beta1.Network        { return nil }
func (s *iprangeTestState) SetKymaNetwork(_ *cloudcontrolv1beta1.Network)    {}
func (s *iprangeTestState) KymaPeering() *cloudcontrolv1beta1.VpcPeering     { return nil }
func (s *iprangeTestState) SetKymaPeering(_ *cloudcontrolv1beta1.VpcPeering) {}

func newOverlapCheckState(t *testing.T, statusCidr, vpcCidr string, secondaryCidrs []string) *State {
	t.Helper()

	scope := &cloudcontrolv1beta1.Scope{
		Spec: cloudcontrolv1beta1.ScopeSpec{
			Scope: cloudcontrolv1beta1.ScopeInfo{
				Alicloud: &cloudcontrolv1beta1.AlicloudScope{
					Network: cloudcontrolv1beta1.AlicloudNetwork{
						VPC: cloudcontrolv1beta1.AlicloudVPC{CIDR: vpcCidr},
					},
				},
			},
		},
	}

	ipRange := &cloudcontrolv1beta1.IpRange{
		ObjectMeta: metav1.ObjectMeta{Name: "test-iprange", Namespace: "kcp-system"},
		Status:     cloudcontrolv1beta1.IpRangeStatus{Cidr: statusCidr},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(commonscheme.KcpScheme).
		WithObjects(ipRange).
		WithStatusSubresource(ipRange).
		Build()

	cluster := composed.NewStateCluster(fakeClient, fakeClient, nil, fakeClient.Scheme())
	focalState := focal.NewStateFactory().NewState(
		composed.NewStateFactory(cluster).NewState(k8stypes.NamespacedName{Name: "test-iprange", Namespace: "kcp-system"}, ipRange),
	)
	focalState.SetScope(scope)

	return &State{
		State:               &iprangeTestState{State: focalState},
		secondaryCidrBlocks: secondaryCidrs,
	}
}

func TestRangeCheckPrimaryOverlap(t *testing.T) {
	const vpcCidr = "10.180.0.0/16"

	t.Run("continues pipeline when CIDR does not overlap", func(t *testing.T) {
		st := newOverlapCheckState(t, "10.181.0.0/22", vpcCidr, nil)
		err, _ := rangeCheckPrimaryOverlap(t.Context(), st)
		assert.Nil(t, err, "action must return nil to continue pipeline")
		assert.Empty(t, st.ObjAsIpRange().Status.Conditions, "no conditions expected on clean path")
	})

	t.Run("stops pipeline with CidrOverlap condition when CIDR overlaps VPC primary", func(t *testing.T) {
		st := newOverlapCheckState(t, "10.180.64.0/22", vpcCidr, nil)
		err, _ := rangeCheckPrimaryOverlap(t.Context(), st)
		require.NotNil(t, err, "action must return non-nil to stop pipeline")
		cond := meta.FindStatusCondition(st.ObjAsIpRange().Status.Conditions, cloudcontrolv1beta1.ConditionTypeError)
		require.NotNil(t, cond)
		assert.Equal(t, cloudcontrolv1beta1.ReasonCidrOverlap, cond.Reason)
		assert.Equal(t, cloudcontrolv1beta1.StateError, st.ObjAsIpRange().Status.State)
	})

	t.Run("stops pipeline with InvalidCidr condition when status CIDR is unparseable", func(t *testing.T) {
		st := newOverlapCheckState(t, "not-a-cidr", vpcCidr, nil)
		err, _ := rangeCheckPrimaryOverlap(t.Context(), st)
		require.NotNil(t, err, "action must return non-nil to stop pipeline")
		cond := meta.FindStatusCondition(st.ObjAsIpRange().Status.Conditions, cloudcontrolv1beta1.ConditionTypeError)
		require.NotNil(t, cond)
		assert.Equal(t, cloudcontrolv1beta1.ReasonInvalidCidr, cond.Reason)
		assert.Equal(t, cloudcontrolv1beta1.StateError, st.ObjAsIpRange().Status.State)
	})
}
