package util

import (
	"embed"
	"fmt"
	"path"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type ObjProvider func() ([]*unstructured.Unstructured, error)

func LoadResources(fs *embed.FS, dir string, prefix string) ([]*unstructured.Unstructured, error) {
	var result []*unstructured.Unstructured
	files, err := fs.ReadDir(dir)
	if err != nil {
		return result, fmt.Errorf("error listing %s crds: %w", dir, err)
	}
	for _, file := range files {
		if prefix != "" && !strings.HasPrefix(file.Name(), prefix) {
			continue
		}
		content, err := fs.ReadFile(path.Join(dir, file.Name()))
		if err != nil {
			return result, fmt.Errorf("error reading crd file %s/%s: %w", dir, file.Name(), err)
		}

		objArr, err := YamlMultiDecodeToUnstructured(content)
		if err != nil {
			return result, fmt.Errorf("error parsing crd yaml file %s/%s: %w", dir, file.Name(), err)
		}

		result = append(result, objArr...)
	}

	return result, nil
}

func MergeLoadedResources(providers ...ObjProvider) ([]*unstructured.Unstructured, error) {
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

func FlattenLists(arr []*unstructured.Unstructured) []*unstructured.Unstructured {
	var result []*unstructured.Unstructured
	for _, obj := range arr {
		if obj.IsList() {
			list, err := obj.ToList()
			if err == nil {
				for _, x := range list.Items {
					result = append(result, &x)
				}
				continue
			}
		}
		result = append(result, obj)
	}
	return result
}
