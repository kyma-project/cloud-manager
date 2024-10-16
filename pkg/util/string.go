package util

import (
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"reflect"
)

func CastInterfaceToString(x interface{}) string {
	if x == nil {
		return ""
	}
	switch x.(type) {
	case string:
		return x.(string)
	case *string:
		return *x.(*string)
	case cloudcontrolv1beta1.ProviderType:
		return string(x.(cloudcontrolv1beta1.ProviderType))
	case *cloudcontrolv1beta1.ProviderType:
		return string(*x.(*cloudcontrolv1beta1.ProviderType))
	default:
		v := reflect.ValueOf(x)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		return fmt.Sprintf("%v", v.Interface())
	}
}
