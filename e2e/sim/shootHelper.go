package sim

import (
	"fmt"

	gardenerhelper "github.com/gardener/gardener/pkg/api/core/v1beta1/helper"
	gardenerapicore "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func IsShootReady(shoot *gardenerapicore.Shoot) bool {
	if shoot.Status.IsHibernated {
		return false
	}
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
