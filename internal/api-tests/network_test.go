package api_tests

import (
	"github.com/google/uuid"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func nb() *cloudcontrolv1beta1.NetworkBuilder {
	return &cloudcontrolv1beta1.NetworkBuilder{}
}

var _ = ginkgo.Describe("Feature: Network", func() {

	createScenario := func(title string, b *cloudcontrolv1beta1.NetworkBuilder, ok bool, errMsg string, focus bool) {
		handler := func() {
			net := b.Build()
			net.Name = uuid.NewString()
			net.Namespace = infra.KCP().Namespace()
			err := infra.KCP().Client().Create(infra.Ctx(), net)
			if ok {
				Expect(err).NotTo(HaveOccurred(), title)
				_ = infra.KCP().Client().Delete(infra.Ctx(), net)
			} else {
				Expect(err).To(HaveOccurred(), title)
				if errMsg != "" {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			}
		}
		if focus {
			ginkgo.It("Scenario: "+title, ginkgo.Focus, handler)
		} else {
			ginkgo.It("Scenario: "+title, handler)
		}
	}

	updateScenario := func(title string, b *cloudcontrolv1beta1.NetworkBuilder, cb func(b *cloudcontrolv1beta1.NetworkBuilder), ok bool, errMsg string, focus bool) {
		handler := func() {
			net := b.Build()
			net.Name = uuid.NewString()
			net.Namespace = infra.KCP().Namespace()
			err := infra.KCP().Client().Create(infra.Ctx(), net)
			Expect(err).NotTo(HaveOccurred())
			err = infra.KCP().Client().Get(infra.Ctx(), client.ObjectKeyFromObject(net), net)
			Expect(err).NotTo(HaveOccurred())
			defer func() {
				_ = infra.KCP().Client().Delete(infra.Ctx(), net)
			}()
			cb(b)
			err = infra.KCP().Client().Update(infra.Ctx(), net)
			if ok {
				Expect(err).NotTo(HaveOccurred(), title)
			} else {
				Expect(err).To(HaveOccurred(), title)
				if errMsg != "" {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			}
		}
		if focus {
			ginkgo.It("Scenario: "+title, ginkgo.Focus, handler)
		} else {
			ginkgo.It("Scenario: "+title, handler)
		}
	}

	canCreate := func(title string, b *cloudcontrolv1beta1.NetworkBuilder) {
		createScenario(title, b, true, "", false)
	}
	canNotCreate := func(title string, b *cloudcontrolv1beta1.NetworkBuilder, errMsg string) {
		createScenario(title, b, false, errMsg, false)
	}
	canNotChange := func(title string, b *cloudcontrolv1beta1.NetworkBuilder, cb func(b *cloudcontrolv1beta1.NetworkBuilder), errMsg string) {
		updateScenario(title, b, cb, false, errMsg, false)
	}

	canCreate(
		"Managed network can be created",
		nb().WithScope("s").WithCidr("10.0.0.0/24"),
	)
	canCreate(
		"GCP network reference can be created",
		nb().WithScope("s").WithGcpRef("proj", "net"),
	)
	canCreate(
		"AWS network reference can be created",
		nb().WithScope("s").WithAwsRef("acc", "us-east1", "vpc123", "name"),
	)
	canCreate(
		"Azure network reference can be created",
		nb().WithScope("s").WithAzureRef("tenant", "sub", "rg", "name"),
	)
	canCreate(
		"OpenStack network reference can be created",
		nb().WithScope("s").WithOpenStackRef("domain", "project", "id", "name"),
	)

	canNotCreate(
		"Managed network w/out scope can not be created",
		nb().WithCidr("10.0.0.0/24"),
		"Scope is required",
	)
	canNotCreate(
		"GCP network reference w/out scope can not be created",
		nb().WithGcpRef("proj", "net"),
		"Scope is required",
	)
	canNotCreate(
		"Network can not be managed and reference at the same time",
		nb().WithScope("s").WithCidr("10.0.0.0/24").WithGcpRef("proj", "net"),
		"Too many: 2: must have at most 1 items",
	)

	canNotChange(
		"Managed network can not change scope",
		nb().WithScope("s").WithCidr("10.11.0.0/25"),
		func(b *cloudcontrolv1beta1.NetworkBuilder) {
			b.WithScope("xxx")
		},
		"Scope is immutable",
	)
	canNotChange(
		"Managed network can not change cidr",
		nb().WithScope("s").WithCidr("10.11.0.0/25"),
		func(b *cloudcontrolv1beta1.NetworkBuilder) {
			b.WithCidr("10.22.0.0/24")
		},
		"Network is immutable",
	)
	canNotChange(
		"GCP network reference can not change scope",
		nb().WithScope("s").WithGcpRef("proj", "net"),
		func(b *cloudcontrolv1beta1.NetworkBuilder) {
			b.WithScope("xxx")
		},
		"Scope is immutable",
	)

	canNotChange(
		"GCP network reference can not change to AWS",
		nb().WithScope("s").WithGcpRef("proj", "net"),
		func(b *cloudcontrolv1beta1.NetworkBuilder) {
			b.WithGcpRef("", "").WithAwsRef("acc", "us-east1", "vpc123", "name")
		},
		"Network is immutable",
	)
	canNotChange(
		"Managed network can not change to GCP reference",
		nb().WithScope("s").WithCidr("10.10.0.0/25"),
		func(b *cloudcontrolv1beta1.NetworkBuilder) {
			b.WithCidr("").WithGcpRef("proj", "name")
		},
		"Network is immutable",
	)
	canNotChange(
		"Network reference can not change GCP attribute",
		nb().WithScope("s").WithGcpRef("proj", "net"),
		func(b *cloudcontrolv1beta1.NetworkBuilder) {
			b.WithGcpRef("xxx", "yyy")
		},
		"Network is immutable",
	)

})
