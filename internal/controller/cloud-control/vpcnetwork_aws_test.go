/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cloudcontrol

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Feature: VpcNetwork", func() {

	It("Scenario: VpcNetwork AWS is created", func() {
		const vpcNetworkName = "3262da42-3fa7-485f-9487-bc66a5fcacc2"

		var vpcNetwork *cloudcontrolv1beta1.VpcNetwork

		By("When VpcNetwork AWS is created", func() {
			// TODO: fix this, finish the test
			vpcNetwork = cloudcontrolv1beta1.NewVpcNetworkBuilder().
				WithSubscription("x").
				WithCidrBlocks("10.250.0.0/16").
				Build()

			Expect(
				CreateObj(infra.Ctx(), infra.KCP().Client(), vpcNetwork, WithName(vpcNetworkName)),
			).To(Succeed())
		})
	})

})
