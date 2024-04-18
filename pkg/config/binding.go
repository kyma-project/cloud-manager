package config

import (
	"github.com/mitchellh/mapstructure"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type binding struct {
	fieldPath FieldPath
	obj       any
}

func (b *binding) Copy(raw map[string]interface{}) {
	val, found, err := unstructured.NestedFieldNoCopy(raw, b.fieldPath...)
	if !found {
		return
	}
	if err != nil {
		return
	}

	err = mapstructure.Decode(val, b.obj)
	if err != nil {
		return
	}
}
