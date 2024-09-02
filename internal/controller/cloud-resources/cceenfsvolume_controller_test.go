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

package cloudresources

import (
	"github.com/kyma-project/cloud-manager/pkg/feature"
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("Feature: SKR CceeNfsVolume", func() {

	It("Scenario: SKR CceeNfsVolume is created with empty IpRange when default IpRange does not exist", func() {

		By("Given ff IpRangeAutomaticCidrAllocation is enabled", func() {
			if !feature.IpRangeAutomaticCidrAllocation.Value(infra.Ctx()) {
				Skip("IpRangeAutomaticCidrAllocation is disabled")
			}
		})

	})
})
