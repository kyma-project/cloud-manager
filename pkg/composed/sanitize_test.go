package composed_test

import (
	"testing"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSanitizeErrorForTenant(t *testing.T) {
	cases := []struct {
		reason string
		want   string
	}{
		{"FailedCreatingFileSystem", "failed to create the file system"},
		{"FailedCreatingVpcPeeringConnection", "failed to create the VPC peering connection"},
		{"CloudProviderError", "a cloud provider error occurred"},
		{"QuotaExceeded", "a cloud provider quota limit has been exceeded"},
		{"Unauthorized", "not authorized to perform the requested cloud operation"},
		// unknown reason falls back to generic
		{"CanNotConnectCloud", "an error occurred while provisioning the resource"},
		{"", "an error occurred while provisioning the resource"},
	}
	for _, c := range cases {
		t.Run(c.reason, func(t *testing.T) {
			got := composed.SanitizeErrorForTenant(c.reason)
			if got != c.want {
				t.Errorf("SanitizeErrorForTenant(%q) = %q, want %q", c.reason, got, c.want)
			}
		})
	}
}

func TestSanitizeConditions(t *testing.T) {
	const rawSDKMessage = "azure: RequestError\nStatusCode=403\nOriginal Error: autorest/authorization: failed to refresh the Token"

	readyCond := metav1.Condition{
		Type:    "Ready",
		Status:  metav1.ConditionFalse,
		Reason:  "NotReady",
		Message: "waiting for cloud resource",
	}
	errCond := metav1.Condition{
		Type:    cloudcontrolv1beta1.ConditionTypeError,
		Status:  metav1.ConditionTrue,
		Reason:  "Unauthorized",
		Message: rawSDKMessage,
	}

	t.Run("nil input returns empty slice", func(t *testing.T) {
		got := composed.SanitizeConditions(nil)
		if len(got) != 0 {
			t.Errorf("expected empty slice, got %v", got)
		}
	})

	t.Run("empty input returns empty slice", func(t *testing.T) {
		got := composed.SanitizeConditions([]metav1.Condition{})
		if len(got) != 0 {
			t.Errorf("expected empty slice, got %v", got)
		}
	})

	t.Run("no error condition: all conditions forwarded unchanged", func(t *testing.T) {
		input := []metav1.Condition{readyCond}
		got := composed.SanitizeConditions(input)
		if len(got) != 1 {
			t.Fatalf("expected 1 condition, got %d", len(got))
		}
		if got[0] != readyCond {
			t.Errorf("condition changed unexpectedly: got %+v, want %+v", got[0], readyCond)
		}
	})

	t.Run("error condition message replaced with tenant-safe string", func(t *testing.T) {
		input := []metav1.Condition{errCond}
		got := composed.SanitizeConditions(input)
		if len(got) != 1 {
			t.Fatalf("expected 1 condition, got %d", len(got))
		}
		if got[0].Message == rawSDKMessage {
			t.Error("raw SDK message was not sanitized")
		}
		want := composed.SanitizeErrorForTenant(errCond.Reason)
		if got[0].Message != want {
			t.Errorf("sanitized message = %q, want %q", got[0].Message, want)
		}
		if got[0].Reason != errCond.Reason {
			t.Errorf("Reason was modified: got %q, want %q", got[0].Reason, errCond.Reason)
		}
	})

	t.Run("non-error conditions co-present: only error message replaced", func(t *testing.T) {
		input := []metav1.Condition{readyCond, errCond}
		got := composed.SanitizeConditions(input)
		if len(got) != 2 {
			t.Fatalf("expected 2 conditions, got %d", len(got))
		}
		if got[0] != readyCond {
			t.Errorf("non-error condition changed: got %+v, want %+v", got[0], readyCond)
		}
		if got[1].Message == rawSDKMessage {
			t.Error("error condition message was not sanitized")
		}
	})

	t.Run("original slice is not mutated", func(t *testing.T) {
		input := []metav1.Condition{errCond}
		_ = composed.SanitizeConditions(input)
		if input[0].Message != rawSDKMessage {
			t.Error("SanitizeConditions mutated the input slice")
		}
	})

	t.Run("unknown error reason falls back to generic message", func(t *testing.T) {
		input := []metav1.Condition{{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  metav1.ConditionTrue,
			Reason:  "SomeUnknownInternalReason",
			Message: rawSDKMessage,
		}}
		got := composed.SanitizeConditions(input)
		want := "an error occurred while provisioning the resource"
		if got[0].Message != want {
			t.Errorf("got %q, want %q", got[0].Message, want)
		}
	})
}
