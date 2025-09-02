package crd

import (
	"embed"
	"fmt"
	"path"
	"strings"

	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

//go:embed bases/*.yaml gardener/*.yaml kim/*.yaml operator/*.yaml
var FS embed.FS

type ObjProvider func() ([]*unstructured.Unstructured, error)

func KCP_All() ([]*unstructured.Unstructured, error) {
	return mergeLoadedResources([]ObjProvider{
		KCP_CloudManager,
		KCP_KIM,
		KLM,
	})
}

func KCP_CloudManager() ([]*unstructured.Unstructured, error) {
	return loadResources("bases", "cloud-control.kyma-project.io_")
}

func SKR_CloudManagerAll() ([]*unstructured.Unstructured, error) {
	return loadResources("bases", "cloud-resources.kyma-project.io_")
}

func SKR_CloudManagerModule() ([]*unstructured.Unstructured, error) {
	return loadResources("bases", "cloud-resources.kyma-project.io_cloudresources.yaml")
}

func KCP_KIM() ([]*unstructured.Unstructured, error) {
	return loadResources("kim", "")
}

func KLM() ([]*unstructured.Unstructured, error) {
	return loadResources("operator", "")
}

func loadResources(dir string, prefix string) ([]*unstructured.Unstructured, error) {
	var result []*unstructured.Unstructured
	files, err := FS.ReadDir(dir)
	if err != nil {
		return result, fmt.Errorf("error listing %s crds: %w", dir, err)
	}
	for _, file := range files {
		if prefix != "" && !strings.HasPrefix(file.Name(), prefix) {
			continue
		}
		content, err := FS.ReadFile(path.Join(dir, file.Name()))
		if err != nil {
			return result, fmt.Errorf("error reading crd file %s/%s: %w", dir, file.Name(), err)
		}

		objArr, err := util.YamlMultiDecodeToUnstructured(content)
		if err != nil {
			return result, fmt.Errorf("error parsing crd yaml file %s/%s: %w", dir, file.Name(), err)
		}

		result = append(result, objArr...)
	}

	return result, nil
}

func mergeLoadedResources(providers []ObjProvider) ([]*unstructured.Unstructured, error) {
	var result []*unstructured.Unstructured
	for _, provider := range providers {
		arr, err := provider()
		if err != nil {
			return result, err
		}
		result = append(result, arr...)
	}
	return result, nil
}