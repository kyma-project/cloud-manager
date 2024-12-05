package types

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ResourceKindState struct {
	Kind    string
	List    client.ObjectList
	Objects []focal.CommonObject
}

type ProviderResourceKindState struct {
	Kind     string
	Provider cloudcontrolv1beta1.ProviderType
	Objects  []ProviderResourceObject
}

type ProviderResourceObject interface {
	GetId() string
	GetObject() interface{}
}

type State interface {
	focal.State
	ObjAsNuke() *cloudcontrolv1beta1.Nuke
}
