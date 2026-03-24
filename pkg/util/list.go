package util

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ExtractList(list client.ObjectList) ([]client.Object, error) {
	arr, err := meta.ExtractList(list)
	if err != nil {
		return nil, err
	}
	result := make([]client.Object, 0, len(arr))
	for _, o := range arr {
		if co, ok := o.(client.Object); ok {
			result = append(result, co)
		} else {
			return nil, fmt.Errorf("not a client.Object %T", o)
		}
	}
	return result, nil
}
