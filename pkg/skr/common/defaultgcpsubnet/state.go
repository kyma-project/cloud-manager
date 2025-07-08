package defaultgcpsubnet

import (
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

type ObjWithGcpSubnetRef interface {
	composed.ObjWithConditionsAndState
	GetGcpSubnetRef() cloudresourcesv1beta1.GcpSubnetRef
}

type State interface {
	composed.State
	GetSkrGcpSubnet() *cloudresourcesv1beta1.GcpSubnet
	SetSkrGcpSubnet(skrGcpSubnet *cloudresourcesv1beta1.GcpSubnet)
	ObjAsObjWithGcpSubnetRef() ObjWithGcpSubnetRef
}
