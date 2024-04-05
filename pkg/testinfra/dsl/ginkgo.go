package dsl

import (
	"github.com/onsi/ginkgo/v2"
)

func SkipDescribe(text string, _ ...interface{}) bool {
	_ = ginkgo.Describe("SKIPPED: "+text, func() {
		ginkgo.It("Scenario: ???", func() {})
		ginkgo.It("Scenario: ???", func() {})
		ginkgo.It("Scenario: ???", func() {})
	})
	return false
}
