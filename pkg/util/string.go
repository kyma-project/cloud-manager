package util

func CaseInterfaceToString(x interface{}) string {
	if x == nil {
		return ""
	}
	s, ok := x.(string)
	if !ok {
		return ""
	}
	return s
}
