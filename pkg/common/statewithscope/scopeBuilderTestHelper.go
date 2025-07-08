package statewithscope

import (
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/types"
	"math/rand"
)

var _ StateWithObjAsScope = &testStateWithObjAsScope{}

type testStateWithObjAsScope struct {
	composed.State
	scope *cloudcontrolv1beta1.Scope
}

func (s *testStateWithObjAsScope) ObjAsScope() *cloudcontrolv1beta1.Scope {
	return s.scope
}

// ===========================================

type testStateBuilder struct {
	scope *cloudcontrolv1beta1.Scope
}

func sb(scope *cloudcontrolv1beta1.Scope) *testStateBuilder {
	return &testStateBuilder{scope: scope}
}

func (b *testStateBuilder) BuildFocal() focal.State {
	baseFactory := composed.NewStateFactory(composed.NewStateCluster(nil, nil, nil, nil))
	baseState := baseFactory.NewState(types.NamespacedName{
		Namespace: "default",
		Name:      fmt.Sprintf("rnd-%d", rand.Intn(1000)),
	}, &cloudcontrolv1beta1.NfsInstance{})
	focalFactory := focal.NewStateFactory()
	focalState := focalFactory.NewState(baseState)
	focalState.SetScope(b.scope)
	return focalState
}

func (b *testStateBuilder) BuildObjAsScope() composed.State {
	return &testStateWithObjAsScope{scope: b.scope}
}

func (b *testStateBuilder) BuildScope() composed.State {
	baseFactory := composed.NewStateFactory(composed.NewStateCluster(nil, nil, nil, nil))
	baseState := baseFactory.NewState(types.NamespacedName{
		Namespace: "default",
		Name:      fmt.Sprintf("rnd-%d", rand.Intn(1000)),
	}, b.scope)
	return baseState
}

func s() *cloudcontrolv1beta1.ScopeBuilder {
	return cloudcontrolv1beta1.NewScopeBuilder()
}
