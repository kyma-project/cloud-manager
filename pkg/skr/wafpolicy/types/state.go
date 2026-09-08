package types

import (
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	commonscope "github.com/kyma-project/cloud-manager/pkg/skr/common/scope"
)

// State is the shared interface for WafPolicy state that all provider-specific states extend.
// It provides common functionality needed by all cloud providers.
type State interface {
	commonscope.State
	ObjAsWafPolicy() *cloudresourcesv1beta1.WafPolicy
}
