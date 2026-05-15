# API Validation Test Patterns

## Location and Naming

- Directory: `internal/api-tests/`
- KCP resources: `kcp_<resource>_test.go`
- SKR resources: `skr_<resource>_test.go`

Tests only validate CRD schema — no reconciliation, no cloud APIs.

## Builder Pattern

Two cases: some API types ship their own builder in the API package; others need a local test builder.

### Case 1: API package has a builder (use it directly)

Check whether `<Type>Builder` exists in the API package (e.g., `cloudcontrolv1beta1.NetworkBuilder`). If it does, use it directly — no local builder needed. Common shortcut aliases used in tests:

```go
// network_test.go — NetworkBuilder lives in the API package
nb := func() *cloudcontrolv1beta1.NetworkBuilder {
    return &cloudcontrolv1beta1.NetworkBuilder{}
}
bb := func(b Builder[*cloudcontrolv1beta1.Network]) *cloudcontrolv1beta1.NetworkBuilder {
    return b.(*cloudcontrolv1beta1.NetworkBuilder)
}

var _ Builder[*cloudcontrolv1beta1.Network] = &cloudcontrolv1beta1.NetworkBuilder{}

canNotCreateKcp(
    "Managed network w/out scope can not be created",
    nb().WithCidr("10.0.0.0/24"),
    "Scope is required",
)
canNotChangeKcp(
    "Managed network can not change scope",
    nb().WithScope("s").WithCidr("10.11.0.0/25"),
    func(b Builder[*cloudcontrolv1beta1.Network]) { bb(b).WithScope("xxx") },
    "Scope is immutable",
)
```

### Case 2: No API builder — define a local test builder

When the API type has no builder, define one in the test file. Start with an empty spec — do not pre-populate defaults the API might not require.

```go
type testGcpNfsVolumeBuilder struct {
    instance cloudresourcesv1beta1.GcpNfsVolume
}

func newTestGcpNfsVolumeBuilder() *testGcpNfsVolumeBuilder {
    return &testGcpNfsVolumeBuilder{
        instance: cloudresourcesv1beta1.GcpNfsVolume{
            Spec: cloudresourcesv1beta1.GcpNfsVolumeSpec{},
        },
    }
}

func (b *testGcpNfsVolumeBuilder) Build() *cloudresourcesv1beta1.GcpNfsVolume {
    return &b.instance
}

func (b *testGcpNfsVolumeBuilder) WithTier(tier cloudresourcesv1beta1.GcpFileTier) *testGcpNfsVolumeBuilder {
    b.instance.Spec.Tier = tier
    return b
}

func (b *testGcpNfsVolumeBuilder) WithCapacityGb(gb int) *testGcpNfsVolumeBuilder {
    b.instance.Spec.CapacityGb = gb
    return b
}

func (b *testGcpNfsVolumeBuilder) WithFileShareName(name string) *testGcpNfsVolumeBuilder {
    b.instance.Spec.FileShareName = name
    return b
}

// Convenience method for valid defaults (use in tests that don't care about this field)
func (b *testGcpNfsVolumeBuilder) WithValidFileShareName() *testGcpNfsVolumeBuilder {
    b.instance.Spec.FileShareName = "foo"
    return b
}
```

## The 7 Test Helpers

```go
// SKR
canCreateSkr(title string, builder Builder[*T])
canNotCreateSkr(title string, builder Builder[*T], expectedErrorMsg string)
canChangeSkr(title string, builder Builder[*T], modifyFn func(Builder[*T]))
canNotChangeSkr(title string, builder Builder[*T], modifyFn func(Builder[*T]), expectedErrorMsg string)

// KCP
canCreateKcp(title string, builder Builder[*T])
canNotCreateKcp(title string, builder Builder[*T], expectedErrorMsg string)
canNotChangeKcp(title string, builder Builder[*T], modifyFn func(Builder[*T]), expectedErrorMsg string)
```

`canNotCreate*` and `canNotChange*` require `expectedErrorMsg` — no exceptions.

## Test File Structure

```go
var _ = Describe("Feature: SKR GcpNfsVolume", Ordered, func() {

    Context("Creation Validation", func() {
        canCreateSkr(
            "REGIONAL tier with valid capacity: 1024",
            newTestGcpNfsVolumeBuilder().WithTier(cloudresourcesv1beta1.REGIONAL).WithCapacityGb(1024),
        )
        canCreateSkr(
            "BASIC_SSD tier with minimum capacity: 2560",
            newTestGcpNfsVolumeBuilder().WithTier(cloudresourcesv1beta1.BASIC_SSD).WithCapacityGb(2560),
        )
    })

    Context("Creation Validation - Errors", func() {
        canNotCreateSkr(
            "REGIONAL tier with capacity below minimum: 1023",
            newTestGcpNfsVolumeBuilder().WithTier(cloudresourcesv1beta1.REGIONAL).WithCapacityGb(1023),
            "REGIONAL tier capacityGb must be between 1024 and 9984",
        )
        canNotCreateSkr(
            "fileShareName with uppercase letters",
            newTestGcpNfsVolumeBuilder().WithFileShareName("InvalidName"),
            "fileShareName must be lowercase",
        )
    })

    Context("Update Validation", func() {
        canChangeSkr(
            "REGIONAL tier capacity can be increased",
            newTestGcpNfsVolumeBuilder().WithCapacityGb(1024),
            func(b Builder[*cloudresourcesv1beta1.GcpNfsVolume]) {
                b.(*testGcpNfsVolumeBuilder).WithCapacityGb(1280)
            },
        )
    })

    Context("Immutability Validation", func() {
        canNotChangeSkr(
            "tier cannot be changed",
            newTestGcpNfsVolumeBuilder().WithTier(cloudresourcesv1beta1.REGIONAL),
            func(b Builder[*cloudresourcesv1beta1.GcpNfsVolume]) {
                b.(*testGcpNfsVolumeBuilder).WithTier(cloudresourcesv1beta1.BASIC_SSD)
            },
            "tier is immutable",
        )
        canNotChangeSkr(
            "fileShareName cannot be changed",
            newTestGcpNfsVolumeBuilder().WithFileShareName("original-name"),
            func(b Builder[*cloudresourcesv1beta1.GcpNfsVolume]) {
                b.(*testGcpNfsVolumeBuilder).WithFileShareName("new-name")
            },
            "fileShareName is immutable",
        )
    })
})
```

## Boundary Conditions (required for numeric fields)

```go
// Test minimum, maximum, just-below-min, just-above-max
canCreateSkr("REGIONAL minimum capacity: 1024", builder.WithCapacityGb(1024))
canCreateSkr("REGIONAL maximum capacity: 9984", builder.WithCapacityGb(9984))
canNotCreateSkr("capacity below minimum: 1023", builder.WithCapacityGb(1023), "at least 1024")
canNotCreateSkr("capacity above maximum: 9985", builder.WithCapacityGb(9985), "must not exceed 9984")
```

## Cross-Field Validation

```go
// Test that validation depends on other fields (e.g., tier affects capacity rules)
canNotCreateSkr(
    "REGIONAL capacity not divisible by 256",
    newTestGcpNfsVolumeBuilder().WithTier(cloudresourcesv1beta1.REGIONAL).WithCapacityGb(1025),
    "REGIONAL tier capacityGb must be divisible by 256",
)
canCreateSkr(
    "BASIC_SSD capacity does not require 256 divisibility",
    newTestGcpNfsVolumeBuilder().WithTier(cloudresourcesv1beta1.BASIC_SSD).WithCapacityGb(2561),
)
```

## Pattern Validation

```go
canCreateSkr("valid fileShareName: lowercase-with-numbers-123", builder.WithFileShareName("my-share-123"))
canNotCreateSkr("fileShareName with uppercase", builder.WithFileShareName("InvalidName"), "must match pattern")
canNotCreateSkr("fileShareName with underscores", builder.WithFileShareName("name_with_underscores"), "must match pattern")
canNotCreateSkr("fileShareName too long", builder.WithFileShareName(strings.Repeat("a", 64)), "exceeds maximum length")
```

## KCP Resource Example

```go
var _ = Describe("Feature: KCP AzureRedisEnterprise", Ordered, func() {

    Context("Creation Validation", func() {
        canCreateKcp(
            "valid SKU: Standard",
            newTestAzureRedisEnterpriseBuilder().WithScope("test-scope").WithSku("Standard").WithCapacity(2),
        )
        canNotCreateKcp(
            "invalid SKU",
            newTestAzureRedisEnterpriseBuilder().WithScope("test-scope").WithSku("InvalidSku").WithCapacity(2),
            "sku must be one of: Basic, Standard, Premium",
        )
    })

    Context("Immutability Validation", func() {
        canNotChangeKcp(
            "scope cannot be changed",
            newTestAzureRedisEnterpriseBuilder().WithScope("original-scope").WithSku("Standard").WithCapacity(2),
            func(b Builder[*cloudcontrolv1beta1.AzureRedisEnterprise]) {
                b.(*testAzureRedisEnterpriseBuilder).WithScope("new-scope")
            },
            "scope is immutable",
        )
    })
})
```

## Checklist Per Field

- [ ] Valid value accepted (`canCreateSkr`)
- [ ] Invalid value rejected (`canNotCreateSkr` with expected error)
- [ ] Boundary: min, max, just-below-min, just-above-max (numeric)
- [ ] Pattern: valid and invalid strings (string with regex)
- [ ] Enum: all allowed values, at least one invalid (enum)
- [ ] Immutability: `canNotChangeSkr` with expected error (immutable fields)
- [ ] Cross-field rules tested explicitly

## Running

```bash
go test ./internal/api-tests -v
go test ./internal/api-tests -ginkgo.focus="GcpNfsVolume"
```

If `canNotCreate*` passes but the resource was created: kubebuilder validation markers are missing.  
Run `make manifests` after adding them.
