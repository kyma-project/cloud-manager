package types

import (
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	commonscope "github.com/kyma-project/cloud-manager/pkg/skr/common/scope"
)

type State interface {
	commonscope.State

	ObjAsCloudResources() *cloudresourcesv1beta1.CloudResources
}
