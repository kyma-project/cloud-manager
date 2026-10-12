package types

import (
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	commonscope "github.com/kyma-project/cloud-manager/pkg/skr/common/scope"
)

// State is the shared interface for Certificate state that all provider-specific states extend.
type State interface {
	commonscope.State
	ObjAsCertificate() *cloudresourcesv1beta1.Certificate
}
