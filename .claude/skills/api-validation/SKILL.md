---
name: api-validation
description: Test CRD validation rules and Kubebuilder markers. Use when writing validation tests, testing CRD constraints, or verifying API field rules.
---

# API Validation Tests

Test CRD validation rules using Kubebuilder markers.

## Quick Start

1. Add validation markers to CRD types
2. Run `make manifests` to generate CRD
3. Write validation tests in `internal/controller/`
4. Use builder pattern for test objects
5. Test both valid and invalid inputs

## Common Validation Markers

```go
// Required field
// +kubebuilder:validation:Required

// Optional field
// +kubebuilder:validation:Optional

// Numeric constraints
// +kubebuilder:validation:Minimum=1
// +kubebuilder:validation:Maximum=100
// +kubebuilder:validation:ExclusiveMinimum=true

// String constraints
// +kubebuilder:validation:MinLength=1
// +kubebuilder:validation:MaxLength=63
// +kubebuilder:validation:Pattern=`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`

// Enum values
// +kubebuilder:validation:Enum=value1;value2;value3

// Array constraints
// +kubebuilder:validation:MinItems=1
// +kubebuilder:validation:MaxItems=10

// Default value
// +kubebuilder:default=defaultValue

// Immutable (CEL)
// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="field is immutable"
```

## Test Pattern

```go
var _ = Describe("GcpResource Validation", func() {

    It("Should reject empty required field", func() {
        obj := &cloudcontrolv1beta1.GcpResource{
            ObjectMeta: metav1.ObjectMeta{
                Name:      "test",
                Namespace: "default",
            },
            Spec: cloudcontrolv1beta1.GcpResourceSpec{
                // Missing required field
            },
        }

        err := k8sClient.Create(ctx, obj)
        Expect(err).To(HaveOccurred())
        Expect(err.Error()).To(ContainSubstring("spec.requiredField"))
    })

    It("Should reject value below minimum", func() {
        obj := &cloudcontrolv1beta1.GcpResource{
            ObjectMeta: metav1.ObjectMeta{
                Name:      "test",
                Namespace: "default",
            },
            Spec: cloudcontrolv1beta1.GcpResourceSpec{
                Count: 0,  // Minimum is 1
            },
        }

        err := k8sClient.Create(ctx, obj)
        Expect(err).To(HaveOccurred())
        Expect(err.Error()).To(ContainSubstring("minimum"))
    })

    It("Should accept valid object", func() {
        obj := &cloudcontrolv1beta1.GcpResource{
            ObjectMeta: metav1.ObjectMeta{
                Name:      "test",
                Namespace: "default",
            },
            Spec: cloudcontrolv1beta1.GcpResourceSpec{
                RequiredField: "value",
                Count:         5,
            },
        }

        err := k8sClient.Create(ctx, obj)
        Expect(err).NotTo(HaveOccurred())
    })
})
```

## Builder Pattern

```go
// Create builder for test objects
func NewGcpResourceBuilder() *gcpResourceBuilder {
    return &gcpResourceBuilder{
        obj: &cloudcontrolv1beta1.GcpResource{
            ObjectMeta: metav1.ObjectMeta{
                Name:      "test-" + rand.String(5),
                Namespace: "default",
            },
        },
    }
}

func (b *gcpResourceBuilder) WithName(name string) *gcpResourceBuilder {
    b.obj.Name = name
    return b
}

func (b *gcpResourceBuilder) WithCount(count int32) *gcpResourceBuilder {
    b.obj.Spec.Count = count
    return b
}

func (b *gcpResourceBuilder) Build() *cloudcontrolv1beta1.GcpResource {
    return b.obj
}

// Usage
obj := NewGcpResourceBuilder().
    WithName("my-resource").
    WithCount(5).
    Build()
```

## CEL Validation Rules

```go
// Immutable field
// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="field is immutable"
Field string `json:"field"`

// Conditional validation
// +kubebuilder:validation:XValidation:rule="self.type != 'premium' || self.capacity >= 100",message="premium requires capacity >= 100"

// Cross-field validation
// +kubebuilder:validation:XValidation:rule="self.minReplicas <= self.maxReplicas",message="minReplicas must be <= maxReplicas"
```

## Test CEL Validation

```go
It("Should reject immutable field change", func() {
    // Create resource
    obj := NewGcpResourceBuilder().
        WithImmutableField("original").
        Build()
    Expect(k8sClient.Create(ctx, obj)).To(Succeed())

    // Try to update immutable field
    obj.Spec.ImmutableField = "changed"
    err := k8sClient.Update(ctx, obj)

    Expect(err).To(HaveOccurred())
    Expect(err.Error()).To(ContainSubstring("immutable"))
})
```

## Generate After Changes

```bash
make manifests
make generate
```

## Checklist

- [ ] Validation markers added to types
- [ ] `make manifests` run
- [ ] Tests for required fields
- [ ] Tests for numeric bounds
- [ ] Tests for enum values
- [ ] Tests for pattern matching
- [ ] Tests for CEL rules (if used)
- [ ] Tests for valid inputs pass

## Related

- Full guide: [docs/agents/guides/API_VALIDATION_TESTS.md](../../../docs/agents/guides/API_VALIDATION_TESTS.md)
- Kubebuilder markers: https://book.kubebuilder.io/reference/markers
