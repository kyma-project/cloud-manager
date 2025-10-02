package util

import "k8s.io/apimachinery/pkg/api/meta"

func IgnoreNoMatch(err error) error {
	if meta.IsNoMatchError(err) {
		return nil
	}
	return err
}
