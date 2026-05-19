package infrastructuremanagerv1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"

	utilscheme "github.com/kyma-project/cloud-manager/pkg/util/scheme"
)

var (
	// GroupVersion is group version used to register these objects
	GroupVersion = schema.GroupVersion{Group: "infrastructuremanager.kyma-project.io", Version: "v1"} //nolint:gochecknoglobals

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &utilscheme.GroupVersionSchemeBuilder{GroupVersion: GroupVersion} //nolint:gochecknoglobals

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme //nolint:gochecknoglobals
)
