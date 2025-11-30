package fixtures

import (
	"embed"

	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

//go:embed *.yaml
var FS embed.FS

func CloudProfiles(namespace string) ([]*unstructured.Unstructured, error) {
	result, err := util.LoadResources(&FS, ".", "cloudprofiles.yaml")
	if err != nil {
		return nil, err
	}
	result = util.FlattenLists(result)
	var zerotimestamp metav1.Time
	for _, obj := range result {
		obj.SetNamespace(namespace)
		obj.SetResourceVersion("")
		obj.SetUID("")
		obj.SetSelfLink("")
		obj.SetCreationTimestamp(zerotimestamp)
		obj.SetGeneration(0)
		obj.SetManagedFields(nil)
	}
	return result, nil
}
