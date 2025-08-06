package util

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"

	"github.com/google/uuid"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
)

func CastInterfaceToString(x interface{}) string {
	v := reflect.ValueOf(x)
	if x == nil || (v.Kind() == reflect.Ptr && v.IsNil()) || (v.Interface() == nil) {
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
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		return fmt.Sprintf("%v", v.Interface())
	}
}

func RandomId(length int) string {
	if length < 1 {
		length = 7
	}
	id := uuid.New()
	result := strings.ReplaceAll(id.String(), "-", "")
	if !unicode.IsLetter(rune(result[0])) {
		result = "c-" + result[1:]
	}
	f := fmt.Sprintf("%%.%ds", length)
	result = fmt.Sprintf(f, result)
	return result
}
