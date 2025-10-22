package e2e

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/external/infrastructuremanagerv1"
)

type SkrCluster interface {
	Cluster

	SubscriptionName() string
	Provider() cloudcontrolv1beta1.ProviderType
	RuntimeID() string
	ShootName() string

	ModuleAdd(ctx context.Context, moduleName string)
}

type defaultSkrCluster struct {
	Cluster
	subscriptionName string
	provider         cloudcontrolv1beta1.ProviderType
	runtimeID        string
	shootName        string
	kubeConfigBytes  []byte
}

func NewSkrCluster(clstr Cluster, rt *infrastructuremanagerv1.Runtime) SkrCluster {
	return &defaultSkrCluster{
		Cluster:          clstr,
		subscriptionName: rt.Spec.Shoot.SecretBindingName,
		provider:         cloudcontrolv1beta1.ProviderType(rt.Spec.Shoot.Provider.Type),
		runtimeID:        rt.Name,
		shootName:        rt.Spec.Shoot.Name,
	}
}

func (s *defaultSkrCluster) SubscriptionName() string {
	return s.subscriptionName
}

func (s *defaultSkrCluster) Provider() cloudcontrolv1beta1.ProviderType {
	return s.provider
}

func (s *defaultSkrCluster) RuntimeID() string {
	return s.runtimeID
}

func (s *defaultSkrCluster) ShootName() string {
	return s.shootName
}

func (s *defaultSkrCluster) ModuleAdd(ctx context.Context, moduleName string) {

}
