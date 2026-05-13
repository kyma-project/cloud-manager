package util

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/google/uuid"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
)

func CastInterfaceToString(x any) string {
	v := reflect.ValueOf(x)
	if x == nil || (v.Kind() == reflect.Pointer && v.IsNil()) || (v.Interface() == nil) {
		return ""
	}
	switch xx := x.(type) {
	case string:
		return xx
	case *string:
		return *xx
	case cloudcontrolv1beta1.ProviderType:
		return string(xx)
	case *cloudcontrolv1beta1.ProviderType:
		return string(*xx)
	default:
		if v.Kind() == reflect.Pointer {
			v = v.Elem()
		}
		return fmt.Sprintf("%v", v.Interface())
	}
}

func RandomString(length int) string {
	id := uuid.New()
	result := strings.ReplaceAll(id.String(), "-", "")
	f := fmt.Sprintf("%%.%ds", length)
	result = fmt.Sprintf(f, result)
	return result
}
