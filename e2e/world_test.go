package e2e

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Feature: World", func() {

	It("Scenario: Dev test in progress", func() {

		By("try using KCP client", func() {
			arr, err := world.Sim().Keb().List(infra.Ctx())
			Expect(err).NotTo(HaveOccurred())
			fmt.Printf("found runtimes:\n")
			for _, x := range arr {
				fmt.Printf("* %+v\n", x.RuntimeID)
			}
		})

	})

})
