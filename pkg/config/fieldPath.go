package config

import (
	"fmt"
)

func ConcatFieldPath(prefix, sufix string) string {
	if prefix == "" {
		return sufix
	}
	if sufix == "" {
		return prefix
	}
	return fmt.Sprintf("%s.%s", prefix, sufix)
}
