package util

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/stretchr/testify/assert"
	"k8s.io/utils/ptr"
	"testing"
)

func TestCastInterfaceToString(t *testing.T) {

	type someStringishType string
	type someIntishType int

	testCases := []struct {
		t string
		x interface{}
		s string
	}{
		{
			t: "string",
			x: "value",
			s: "value",
		},
		{
			t: "*string",
			x: ptr.To("value"),
			s: "value",
		},
		{
			t: "cloudcontrolv1beta1.ProviderType",
			x: cloudcontrolv1beta1.ProviderGCP,
			s: string(cloudcontrolv1beta1.ProviderGCP),
		},
		{
			t: "*cloudcontrolv1beta1.ProviderType",
			x: ptr.To(cloudcontrolv1beta1.ProviderGCP),
			s: string(cloudcontrolv1beta1.ProviderGCP),
		},
		{
			t: "*cloudcontrolv1beta1.ProviderType nil",
			x: (*cloudcontrolv1beta1.ProviderType)(nil),
			s: "",
		}, {
			t: "someStringishType",
			x: someStringishType("value"),
			s: "value",
		},
		{
			t: "*someStringishType",
			x: ptr.To(someStringishType("value")),
			s: "value",
		},
		{
			t: "*someStringishType nil",
			x: (*someStringishType)(nil),
			s: "",
		},
		{
			t: "someIntishType",
			x: someIntishType(123),
			s: "123",
		},
		{
			t: "*someIntishType",
			x: ptr.To(someIntishType(123)),
			s: "123",
		},
		{
			t: "*someIntishType nil",
			x: (*someIntishType)(nil),
			s: "",
		},
		{
			t: "nil",
			x: nil,
			s: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.t, func(t *testing.T) {
			assert.Equal(t, tc.s, CastInterfaceToString(tc.x))
		})
	}

}
