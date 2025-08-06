package e2e

import (
	gardenertypes "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/external/keb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ShootBuilder struct {
	Obj gardenertypes.Shoot
}

func NewShootBuilder() *ShootBuilder {
	return &ShootBuilder{
		Obj: gardenertypes.Shoot{
			TypeMeta: metav1.TypeMeta{
				APIVersion: gardenertypes.SchemeGroupVersion.String(),
				Kind:       "Shoot",
			},
			ObjectMeta: metav1.ObjectMeta{
				Namespace: Config.GardenNamespace,
			},
			Spec: gardenertypes.ShootSpec{},
		},
	}
}

func (b *ShootBuilder) WithRuntime(runtime *keb.Runtime) *ShootBuilder {
	b.Obj.Name = runtime.Spec.Shoot.Name
	// TODO ...
	return b
}

func (b *ShootBuilder) Validate() error {
	return nil
}

func (b *ShootBuilder) Build() *gardenertypes.Shoot {
	return &b.Obj
}