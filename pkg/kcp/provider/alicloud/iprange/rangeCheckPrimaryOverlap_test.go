package iprange

import (
	"testing"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	commonscheme "github.com/kyma-project/cloud-manager/pkg/common/scheme"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// iprangeTestState is a minimal stub satisfying iprangetypes.State for unit tests.
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

func buildCheckOverlapState(t *testing.T, statusCidr string, vpcCidr string, secondaryCidrs []string) *State {
	t.Helper()

	scope := &cloudcontrolv1beta1.Scope{
		Spec: cloudcontrolv1beta1.ScopeSpec{
			Scope: cloudcontrolv1beta1.ScopeInfo{
				Alicloud: &cloudcontrolv1beta1.AlicloudScope{
					Network: cloudcontrolv1beta1.AlicloudNetwork{
						VPC: cloudcontrolv1beta1.AlicloudVPC{
							CIDR: vpcCidr,
						},
					},
				},
			},
		},
	}

	ipRange := &cloudcontrolv1beta1.IpRange{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-iprange",
			Namespace: "kcp-system",
		},
		Status: cloudcontrolv1beta1.IpRangeStatus{
			Cidr: statusCidr,
		},
	}

	k8sClient := fake.NewClientBuilder().
		WithScheme(commonscheme.KcpScheme).
		WithStatusSubresource(ipRange).
		WithObjects(ipRange).
		Build()

	cluster := composed.NewStateCluster(k8sClient, k8sClient, nil, k8sClient.Scheme())
	focalState := focal.NewStateFactory().NewState(
		composed.NewStateFactory(cluster).NewState(types.NamespacedName{Name: "test-iprange", Namespace: "kcp-system"}, ipRange),
	)
	focalState.SetScope(scope)

	return &State{
		State:               &iprangeTestState{State: focalState},
		secondaryCidrBlocks: secondaryCidrs,
	}
}

func TestRangeCheckPrimaryOverlap(t *testing.T) {
	const vpcCidr = "10.180.0.0/16"

	t.Run("passes when CIDR is outside VPC range", func(t *testing.T) {
		st := buildCheckOverlapState(t, "10.181.0.0/22", vpcCidr, nil)
		err, _ := rangeCheckPrimaryOverlap(t.Context(), st)
		assert.Nil(t, err)
		cond := findCondition(st.ObjAsIpRange().Status.Conditions, cloudcontrolv1beta1.ConditionTypeError)
		assert.Nil(t, cond, "no error condition expected")
	})

	t.Run("errors when CIDR overlaps VPC primary CIDR", func(t *testing.T) {
		st := buildCheckOverlapState(t, "10.180.64.0/22", vpcCidr, nil)
		err, _ := rangeCheckPrimaryOverlap(t.Context(), st)
		require.NotNil(t, err)
		cond := findCondition(st.ObjAsIpRange().Status.Conditions, cloudcontrolv1beta1.ConditionTypeError)
		require.NotNil(t, cond)
		assert.Equal(t, cloudcontrolv1beta1.ReasonCidrOverlap, cond.Reason)
	})

	t.Run("errors when CIDR overlaps an existing secondary CIDR block", func(t *testing.T) {
		st := buildCheckOverlapState(t, "10.181.0.0/22", vpcCidr, []string{"10.181.0.0/20"})
		err, _ := rangeCheckPrimaryOverlap(t.Context(), st)
		require.NotNil(t, err)
		cond := findCondition(st.ObjAsIpRange().Status.Conditions, cloudcontrolv1beta1.ConditionTypeError)
		require.NotNil(t, cond)
		assert.Equal(t, cloudcontrolv1beta1.ReasonCidrOverlap, cond.Reason)
	})

	t.Run("passes when CIDR is identical to an existing secondary CIDR block (idempotent)", func(t *testing.T) {
		st := buildCheckOverlapState(t, "10.181.0.0/22", vpcCidr, []string{"10.181.0.0/22"})
		err, _ := rangeCheckPrimaryOverlap(t.Context(), st)
		assert.Nil(t, err)
		cond := findCondition(st.ObjAsIpRange().Status.Conditions, cloudcontrolv1beta1.ConditionTypeError)
		assert.Nil(t, cond, "identical secondary block is idempotent, not an overlap error")
	})

	t.Run("errors when status CIDR is unparseable", func(t *testing.T) {
		st := buildCheckOverlapState(t, "not-a-cidr", vpcCidr, nil)
		err, _ := rangeCheckPrimaryOverlap(t.Context(), st)
		require.NotNil(t, err)
		cond := findCondition(st.ObjAsIpRange().Status.Conditions, cloudcontrolv1beta1.ConditionTypeError)
		require.NotNil(t, cond)
		assert.Equal(t, cloudcontrolv1beta1.ReasonInvalidCidr, cond.Reason)
	})
}

func findCondition(conditions []metav1.Condition, condType string) *metav1.Condition {
	for i := range conditions {
		if conditions[i].Type == condType {
			return &conditions[i]
		}
	}
	return nil
}
