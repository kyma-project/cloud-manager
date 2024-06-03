package defaultiprange

import (
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

type ObjWithIpRangeRef interface {
	composed.ObjWithConditionsAndState
	GetIpRangeRef() cloudresourcesv1beta1.IpRangeRef
}

type State interface {
	composed.State
	GetSkrIpRange() *cloudresourcesv1beta1.IpRange
	SetSkrIpRange(skrIpRange *cloudresourcesv1beta1.IpRange)
	ObjAsObjWithIpRangeRef() ObjWithIpRangeRef
}
