package dsl

import (
	"github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func SucceedIgnoreNotFound() types.GomegaMatcher {
	return &succeedIgnoreNotFoundMatcher{
		inner: gomega.Succeed(),
	}
}

type succeedIgnoreNotFoundMatcher struct {
	inner types.GomegaMatcher
}

func (m *succeedIgnoreNotFoundMatcher) Match(actual any) (success bool, err error) {
	if err, ok := actual.(error); ok {
		if client.IgnoreNotFound(err) == nil {
			return true, nil
		}
	}
	return m.inner.Match(actual)
}

func (m *succeedIgnoreNotFoundMatcher) FailureMessage(actual any) (message string) {
	return m.inner.FailureMessage(actual)
}

func (m *succeedIgnoreNotFoundMatcher) NegatedFailureMessage(actual any) (message string) {
	return m.inner.NegatedFailureMessage(actual)
}
