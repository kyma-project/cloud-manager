package crd

import (
	"embed"

	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

//go:embed bases/*.yaml gardener/*.yaml kim/*.yaml operator/*.yaml
var FS embed.FS

func KCP_All() ([]*unstructured.Unstructured, error) {
	return util.MergeLoadedResources(
		KCP_CloudManager,
		KCP_KIM,
		KLM,
	)
}

func KCP_CloudManager() ([]*unstructured.Unstructured, error) {
	return util.LoadResources(&FS, "bases", "cloud-control.kyma-project.io_")
}

func SKR_CloudManagerAll() ([]*unstructured.Unstructured, error) {
	return util.LoadResources(&FS, "bases", "cloud-resources.kyma-project.io_")
}

func SKR_CloudManagerModule() ([]*unstructured.Unstructured, error) {
	return util.LoadResources(&FS, "bases", "cloud-resources.kyma-project.io_cloudresources.yaml")
}

func KCP_KIM() ([]*unstructured.Unstructured, error) {
	return util.LoadResources(&FS, "kim", "")
}

func KLM() ([]*unstructured.Unstructured, error) {
	return util.LoadResources(&FS, "operator", "")
}
