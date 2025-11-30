package sim

import (
	"fmt"

	gardenerapicore "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	gardenerhelper "github.com/gardener/gardener/pkg/apis/core/v1beta1/helper"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func IsShootReady(shoot *gardenerapicore.Shoot) bool {
	if len(shoot.Status.Conditions) == 0 {
		return false
	}
	for _, ct := range GardenerConditionTypes {
		cond := gardenerhelper.GetCondition(shoot.Status.Conditions, ct)
		if cond == nil {
			return false
		}
		if cond.Status != gardenerapicore.ConditionTrue {
			return false
		}
	}
	return true
}

// HavingShootReady implements testinfra.dsl.ObjAssertion
func HavingShootReady(obj client.Object) error {
	x, ok := obj.(*gardenerapicore.Shoot)
	if !ok {
		return fmt.Errorf("expected *Shoot type but got %T", obj)
	}
	if IsShootReady(x) {
		return nil
	}
	return fmt.Errorf("shoot is not ready")
}
