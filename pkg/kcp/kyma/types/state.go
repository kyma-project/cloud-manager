package types

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type State interface {
	composed.State

	ObjAsKyma() *unstructured.Unstructured
	Scope() *cloudcontrolv1beta1.Scope
	DesiredEnabled() bool
}
