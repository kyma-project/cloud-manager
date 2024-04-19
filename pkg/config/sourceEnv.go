package config

import (
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/tidwall/sjson"
	"regexp"
	"strings"
)

type sourceEnv struct {
	env          abstractions.Environment
	fieldPath    FieldPath
	envVarPrefix string
}

func (s *sourceEnv) Read(in string) string {
	for k, v := range s.env.List() {
		if !strings.HasPrefix(k, s.envVarPrefix) {
			continue
		}

		fieldPath := append(FieldPath{}, s.fieldPath...)
		kk := strings.TrimPrefix(k, s.envVarPrefix)
		kk = strings.TrimPrefix(kk, "_")
		kk = strings.TrimPrefix(kk, "_")
		kk = strings.TrimPrefix(kk, "-")
		kk = strings.TrimPrefix(kk, "-")
		parts := strings.Split(strings.ToLower(kk), "__")
		rx := regexp.MustCompile("([-_][a-z])")
		for _, p := range parts {
			res := rx.ReplaceAllStringFunc(p, func(s string) string {
				s = strings.ToUpper(s)
				s = strings.ReplaceAll(s, "_", "")
				s = strings.ReplaceAll(s, "-", "")
				return s
			})
			fieldPath = append(fieldPath, res)
		}

		changed, err := sjson.Set(in, fieldPath.String(), v)
		if err != nil {
			continue
		}
		in = changed
	}

	return in
}
