package operatorv1beta2

import (
	"github.com/kyma-project/cloud-manager/pkg/external/operatorshared"
	utilscheme "github.com/kyma-project/cloud-manager/pkg/util/scheme"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	// GroupVersion is group version used to register these objects.
	GroupVersion = schema.GroupVersion{
		Group:   operatorshared.OperatorGroup,
		Version: "v1beta2",
	}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme.
	SchemeBuilder = &utilscheme.GroupVersionSchemeBuilder{GroupVersion: GroupVersion}

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)
