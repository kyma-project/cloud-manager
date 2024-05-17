package util

import "fmt"

func CastInterfaceToString(x interface{}) string {
	if x == nil {
		return ""
	}
	s, ok := x.(string)
	if !ok {
		return fmt.Sprintf("%v", x)
	}
	return s
}
