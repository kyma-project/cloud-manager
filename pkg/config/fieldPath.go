package config

import (
	"fmt"
	"github.com/elliotchance/pie/v2"
	"strings"
)

type FieldPath []string

func NewFiledPath(fields ...string) FieldPath {
	return append(FieldPath{}, fields...)
}

func PrependFieldPath(fp FieldPath, fields ...string) FieldPath {
	result := NewFiledPath(fields...)
	result = append(result, fp...)
	return result
}

func AppendFieldPath(fp FieldPath, fields ...string) FieldPath {
	result := NewFiledPath(fp...)
	result = append(result, fields...)
	return result
}

func (fp FieldPath) String() string {
	return strings.Join(pie.Map(fp, func(s string) string {
		if strings.Contains(s, ".") {
			return fmt.Sprintf("\"%s\"", s)
		} else {
			return s
		}
	}), ".")
}

func (fp FieldPath) Len() int {
	return len(fp)
}
