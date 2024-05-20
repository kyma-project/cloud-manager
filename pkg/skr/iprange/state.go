package iprange

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
)

type State struct {
	composed.State
	KymaRef    klog.ObjectRef
	KcpCluster composed.StateCluster
	Provider   *cloudcontrolv1beta1.ProviderType

	KcpIpRange *cloudcontrolv1beta1.IpRange
}

func newStateFactory(baseStateFactory composed.StateFactory, kymaRef klog.ObjectRef, kcpCluster composed.StateCluster, provider *cloudcontrolv1beta1.ProviderType) *stateFactory {
	return &stateFactory{
		baseStateFactory: baseStateFactory,
		kymaRef:          kymaRef,
		kcpCluster:       kcpCluster,
		provider:         provider,
	}
}

type stateFactory struct {
	baseStateFactory composed.StateFactory
	kymaRef          klog.ObjectRef
	kcpCluster       composed.StateCluster
	provider         *cloudcontrolv1beta1.ProviderType
}

func (f *stateFactory) NewState(req ctrl.Request) *State {
	return &State{
		State:      f.baseStateFactory.NewState(req.NamespacedName, &cloudresourcesv1beta1.IpRange{}),
		KymaRef:    f.kymaRef,
		KcpCluster: f.kcpCluster,
		Provider:   f.provider,
	}
}

func (s *State) ObjAsIpRange() *cloudresourcesv1beta1.IpRange {
	return s.Obj().(*cloudresourcesv1beta1.IpRange)
}

func (s *State) MapConditionToState() (f func(obj composed.ObjWithConditions) (string, bool)) {
	return func(obj composed.ObjWithConditions) (string, bool) {
		if obj == nil || len(*obj.Conditions()) == 0 {
			return "", false
		}
		// if Ready condition exists, return Ready state
		readyCondition := meta.FindStatusCondition(*obj.Conditions(), cloudresourcesv1beta1.ConditionTypeReady)
		if readyCondition != nil && readyCondition.Status == metav1.ConditionTrue {
			return cloudresourcesv1beta1.StateReady, true
		}
		// if Submitted condition exists, return Creating state
		submittedCondition := meta.FindStatusCondition(*obj.Conditions(), cloudresourcesv1beta1.ConditionTypeSubmitted)
		if submittedCondition != nil && submittedCondition.Status == metav1.ConditionTrue && len(*obj.Conditions()) == 1 {
			return cloudresourcesv1beta1.StateCreating, true
		}
		// if quota exceeded condition exists, return Error state
		quotaExceededCondition := meta.FindStatusCondition(*obj.Conditions(), cloudresourcesv1beta1.ConditionTypeQuotaExceeded)
		if quotaExceededCondition != nil && quotaExceededCondition.Status == metav1.ConditionTrue && len(*obj.Conditions()) == 1 {
			return cloudresourcesv1beta1.StateError, true
		}
		// if Error condition exists, return Error state
		errorCondition := meta.FindStatusCondition(*obj.Conditions(), cloudresourcesv1beta1.ConditionTypeError)
		if errorCondition != nil && errorCondition.Status == metav1.ConditionTrue {
			return cloudresourcesv1beta1.StateError, true
		}
		// if warning condition exists, return Error state
		warningCondition := meta.FindStatusCondition(*obj.Conditions(), cloudresourcesv1beta1.ConditionTypeWarning)
		if warningCondition != nil && warningCondition.Status == metav1.ConditionTrue {
			return cloudresourcesv1beta1.StateError, true
		}
		return "", false
	}
}
