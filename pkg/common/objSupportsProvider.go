package common

import (
	"errors"
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/objkind"
	featuretypes "github.com/kyma-project/cloud-manager/pkg/feature/types"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	providerSpecificTypes = map[string]cloudcontrolv1beta1.ProviderType{
		// AWS
		fmt.Sprintf("%T", &cloudresourcesv1beta1.AwsNfsVolume{}): cloudcontrolv1beta1.ProviderAws,

		// GCP
		fmt.Sprintf("%T", &cloudresourcesv1beta1.GcpNfsVolume{}):        cloudcontrolv1beta1.ProviderGCP,
		fmt.Sprintf("%T", &cloudresourcesv1beta1.GcpNfsVolumeBackup{}):  cloudcontrolv1beta1.ProviderGCP,
		fmt.Sprintf("%T", &cloudresourcesv1beta1.GcpNfsVolumeRestore{}): cloudcontrolv1beta1.ProviderGCP,
	}
)

// ObjSupportsProvider returns true if given client.Object supports the given provider. The
// o client.Object parameter can be a typed instance or unstructured. First it will try to find
// its GVK in scheme and create new instance. If that instance implements ProviderAwareObject
// it will obtain slice of supported providers from object and return true if that slice is empty
// or given provider is listed in that slice. Otherwise, if object is not ProviderAwareObject or
// not able to be created new from scheme, then a fallback is made to the fixed hard-coded
// definition. This fixed definition is deprecated and a migration to ProviderAwareObject should be made.
func ObjSupportsProvider(o client.Object, scheme *runtime.Scheme, provider string) bool {
	if provider == "" {
		return true
	}
	res, err := func() (bool, error) {
		kindInfo := objkind.ObjectKinds(o, scheme)
		if !kindInfo.AnyOK() {
			return false, errors.New("unable to determine object kind")
		}
		gk := kindInfo.RealObjGK()
		versions := scheme.VersionsForGroupKind(gk)
		if len(versions) == 0 {
			return false, fmt.Errorf("ne versions in scheme for GK %s", gk)
		}
		gvk := versions[0].WithKind(gk.Kind)
		obj, err := scheme.New(gvk)
		if err != nil {
			return false, err
		}
		if pa, ok := obj.(featuretypes.ProviderAwareObject); ok {
			supportedProviders := pa.SpecificToProviders()
			if len(supportedProviders) == 0 {
				return true, nil
			}
			for _, p := range supportedProviders {
				if p == provider {
					return true, nil
				}
			}
			return false, nil
		}
		return false, errors.New("not a ProviderAwareObject object")
	}()

	//res, err := func() (bool, error) {
	//	gvkList, _, err := scheme.ObjectKinds(o)
	//	if err != nil {
	//		return false, err
	//	}
	//	if len(gvkList) == 0 {
	//		return false, errors.New("empty gvk list")
	//	}
	//	obj, err := scheme.New(gvkList[0])
	//	if err != nil {
	//		return false, err
	//	}
	//	if pa, ok := obj.(featuretypes.ProviderAwareObject); ok {
	//		supportedProviders := pa.SpecificToProviders()
	//		if len(supportedProviders) == 0 {
	//			return true, nil
	//		}
	//		for _, p := range supportedProviders {
	//			if p == provider {
	//				return true, nil
	//			}
	//		}
	//		return false, nil
	//	}
	//	return false, errors.New("not a ProviderAwareObject object")
	//}()

	if err == nil {
		return res
	}

	pt, ptDefined := providerSpecificTypes[fmt.Sprintf("%T", o)]
	if !ptDefined {
		return true
	}
	if string(pt) == provider {
		return true
	}
	return false
}
