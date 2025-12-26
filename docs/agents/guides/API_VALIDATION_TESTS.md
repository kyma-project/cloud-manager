# Guide: API Validation Tests

**Target Audience**: LLM coding agents  
**Prerequisites**: CRD definitions in `api/`, kubebuilder markers added  
**Purpose**: Test CRD validation rules, webhooks, field constraints without cloud providers  
**Context**: Verify OpenAPI schema validation, immutability, cross-field rules using envtest

## Authority: API Validation Testing Requirements

### MUST

- **MUST test valid AND invalid**: BOTH valid creation and invalid creation scenarios REQUIRED
- **MUST use builder pattern**: Use fluent builder for constructing test instances
- **MUST use helper functions**: Use `canCreateSkr/canNotCreateSkr/canChangeSkr/canNotChangeSkr`
- **MUST test boundary conditions**: Min, max, edge cases REQUIRED for numeric fields
- **MUST test immutability**: Verify immutable fields reject updates
- **MUST include expected errors**: `canNotCreate*` and `canNotChange*` REQUIRE expectedErrorMsg parameter
- **MUST test patterns**: Regex validation for string fields
- **MUST test enums**: Verify only allowed enum values accepted

### MUST NOT

- **NEVER test reconciliation logic**: API validation tests ONLY test CRD validation, not controllers
- **NEVER skip invalid cases**: Every validation rule MUST have negative test case
- **NEVER hardcode resource names**: Use builder with sensible defaults
- **NEVER test without expected error**: Invalid scenarios MUST specify expected error message
- **NEVER skip cross-field validation**: Test field dependencies (e.g., tier affects capacity rules)

### ALWAYS

- **ALWAYS use descriptive test names**: Name explains what's being validated
- **ALWAYS test one rule per test**: Clear separation for each validation constraint
- **ALWAYS group related tests**: Use Context blocks for Creation/Update/Immutability
- **ALWAYS verify error messages**: Check error messages are clear and helpful to users

### NEVER

- **NEVER mix validation and controller tests**: Keep API validation separate from controller logic
- **NEVER test implementation details**: Test observable validation behavior only
- **NEVER rely on test order**: Each test independent, no shared mutable state
- **NEVER skip cleanup**: Use Ordered tests or proper cleanup for resources

## Test Location and Structure

**Location**: `internal/api-tests/`

**Naming Convention**:
- KCP resources: `kcp_<resource>_test.go`
- SKR resources: `skr_<resource>_test.go`

### ❌ WRONG: No Builder Pattern

```go
// NEVER: Creating resources inline without builder
It("creates resource", func() {
    resource := &cloudresourcesv1beta1.GcpNfsVolume{
        ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
        Spec: cloudresourcesv1beta1.GcpNfsVolumeSpec{
            Tier:          cloudresourcesv1beta1.REGIONAL,
            CapacityGb:    1024,
            FileShareName: "test-share",
        },
    }
    // WRONG: Repetitive, error-prone
})
```

### ✅ CORRECT: Builder Pattern

```go
// ALWAYS: Use builder for constructing test instances
type testGcpNfsVolumeBuilder struct {
    instance cloudresourcesv1beta1.GcpNfsVolume
}

func newTestGcpNfsVolumeBuilder() *testGcpNfsVolumeBuilder {
    return &testGcpNfsVolumeBuilder{
        instance: cloudresourcesv1beta1.GcpNfsVolume{
            ObjectMeta: metav1.ObjectMeta{
                Name:      "test-volume",
                Namespace: "default",
            },
            Spec: cloudresourcesv1beta1.GcpNfsVolumeSpec{
                // Minimal valid defaults
                Tier:          cloudresourcesv1beta1.REGIONAL,
                CapacityGb:    1024,
                FileShareName: "valid-name",
            },
        },
    }
}

func (b *testGcpNfsVolumeBuilder) Build() *cloudresourcesv1beta1.GcpNfsVolume {
    return &b.instance
}

// Fluent setter methods
func (b *testGcpNfsVolumeBuilder) WithTier(tier cloudresourcesv1beta1.GcpFileTier) *testGcpNfsVolumeBuilder {
    b.instance.Spec.Tier = tier
    return b
}

func (b *testGcpNfsVolumeBuilder) WithCapacityGb(capacityGb int) *testGcpNfsVolumeBuilder {
    b.instance.Spec.CapacityGb = capacityGb
    return b
}

func (b *testGcpNfsVolumeBuilder) WithFileShareName(name string) *testGcpNfsVolumeBuilder {
    b.instance.Spec.FileShareName = name
    return b
}
```

## Test Helper Functions

**Location**: `internal/api-tests/builder_test.go`

### SKR Resource Helpers

```go
// Test resource CAN be created (valid scenario)
canCreateSkr(title string, builder Builder[*ResourceType])

// Test resource CANNOT be created (validation fails)
canNotCreateSkr(title string, builder Builder[*ResourceType], expectedErrorMsg string)

// Test field CAN be updated (mutable field)
canChangeSkr(title string, builder Builder[*ResourceType], modifyFunc func(Builder[*ResourceType]))

// Test field CANNOT be updated (immutable field)
canNotChangeSkr(title string, builder Builder[*ResourceType], modifyFunc func(Builder[*ResourceType]), expectedErrorMsg string)
```

### KCP Resource Helpers

```go
// Test KCP resource CAN be created
canCreateKcp(title string, builder Builder[*ResourceType])

// Test KCP resource CANNOT be created
canNotCreateKcp(title string, builder Builder[*ResourceType], expectedErrorMsg string)

// Test KCP field CANNOT be updated (immutable)
canNotChangeKcp(title string, builder Builder[*ResourceType], modifyFunc func(Builder[*ResourceType]), expectedErrorMsg string)
```

## Writing Validation Tests

### Test File Structure

```go
package apitests

import (
    cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
    . "github.com/onsi/ginkgo/v2"
)

var _ = Describe("Feature: SKR GcpNfsVolume", Ordered, func() {
    
    Context("Creation Validation", func() {
        // Valid creation scenarios
    })
    
    Context("Creation Validation - Errors", func() {
        // Invalid creation scenarios
    })
    
    Context("Update Validation", func() {
        // Allowed update scenarios
    })
    
    Context("Immutability Validation", func() {
        // Immutable field scenarios
    })
})
```

### Testing Valid Creation

#### ❌ WRONG: Not Testing Valid Cases

```go
// NEVER: Only testing invalid cases
Context("Creation Validation", func() {
    canNotCreateSkr("invalid capacity", builder.WithCapacityGb(100), "must be at least 1024")
    // WRONG: No valid cases tested!
})
```

#### ✅ CORRECT: Test Valid Scenarios

```go
// ALWAYS: Test valid creation scenarios
Context("Creation Validation", func() {
    
    canCreateSkr(
        "GcpNfsVolume REGIONAL tier with valid capacity: 1024",
        newTestGcpNfsVolumeBuilder().
            WithTier(cloudresourcesv1beta1.REGIONAL).
            WithCapacityGb(1024).
            WithFileShareName("valid-name"),
    )
    
    canCreateSkr(
        "GcpNfsVolume BASIC_SSD tier with minimum capacity: 2560",
        newTestGcpNfsVolumeBuilder().
            WithTier(cloudresourcesv1beta1.BASIC_SSD).
            WithCapacityGb(2560).
            WithFileShareName("valid-name"),
    )
    
    canCreateSkr(
        "GcpNfsVolume with fileShareName containing numbers and hyphens",
        newTestGcpNfsVolumeBuilder().
            WithTier(cloudresourcesv1beta1.REGIONAL).
            WithCapacityGb(1024).
            WithFileShareName("my-share-123"),
    )
})
```

### Testing Invalid Creation

#### ❌ WRONG: No Expected Error Message

```go
// NEVER: Missing expected error message
canNotCreateSkr(
    "GcpNfsVolume with invalid capacity",
    newTestGcpNfsVolumeBuilder().WithCapacityGb(100),
    // WRONG: No expected error parameter!
)
```

#### ✅ CORRECT: Specify Expected Error

```go
// ALWAYS: Include expected error message
Context("Creation Validation - Errors", func() {
    
    canNotCreateSkr(
        "GcpNfsVolume REGIONAL tier with invalid capacity: 1023",
        newTestGcpNfsVolumeBuilder().
            WithTier(cloudresourcesv1beta1.REGIONAL).
            WithCapacityGb(1023).
            WithFileShareName("valid-name"),
        "REGIONAL tier capacityGb must be between 1024 and 9984, and it must be divisible by 256",
    )
    
    canNotCreateSkr(
        "GcpNfsVolume BASIC_SSD tier with capacity below minimum: 2559",
        newTestGcpNfsVolumeBuilder().
            WithTier(cloudresourcesv1beta1.BASIC_SSD).
            WithCapacityGb(2559).
            WithFileShareName("valid-name"),
        "BASIC_SSD tier capacityGb must be at least 2560",
    )
    
    canNotCreateSkr(
        "GcpNfsVolume with invalid fileShareName containing uppercase",
        newTestGcpNfsVolumeBuilder().
            WithTier(cloudresourcesv1beta1.REGIONAL).
            WithCapacityGb(1024).
            WithFileShareName("InvalidName"),
        "fileShareName must be lowercase",
    )
})
```

### Testing Allowed Updates

#### ❌ WRONG: Not Providing Modify Function

```go
// NEVER: Missing modify function for canChangeSkr
canChangeSkr(
    "capacity can be increased",
    newTestGcpNfsVolumeBuilder().WithCapacityGb(1024),
    // WRONG: No modify function!
)
```

#### ✅ CORRECT: Modify Function Updates Field

```go
// ALWAYS: Provide modify function that changes field
Context("Update Validation", func() {
    
    canChangeSkr(
        "GcpNfsVolume REGIONAL tier capacity can be increased",
        newTestGcpNfsVolumeBuilder().
            WithTier(cloudresourcesv1beta1.REGIONAL).
            WithCapacityGb(1024).
            WithFileShareName("valid-name"),
        func(b Builder[*cloudresourcesv1beta1.GcpNfsVolume]) {
            b.(*testGcpNfsVolumeBuilder).WithCapacityGb(1280)
        },
    )
    
    canChangeSkr(
        "GcpNfsVolume BASIC_SSD tier capacity can be increased",
        newTestGcpNfsVolumeBuilder().
            WithTier(cloudresourcesv1beta1.BASIC_SSD).
            WithCapacityGb(2560).
            WithFileShareName("valid-name"),
        func(b Builder[*cloudresourcesv1beta1.GcpNfsVolume]) {
            b.(*testGcpNfsVolumeBuilder).WithCapacityGb(2816)
        },
    )
})
```

### Testing Immutability

#### ❌ WRONG: Using canChangeSkr for Immutable Field

```go
// NEVER: Using canChangeSkr for field that should be immutable
canChangeSkr(
    "tier can be changed",
    newTestGcpNfsVolumeBuilder().WithTier(REGIONAL),
    func(b Builder[*ResourceType]) {
        b.(*testGcpNfsVolumeBuilder).WithTier(BASIC_SSD)
    },
)
// WRONG: Will fail because tier is immutable!
```

#### ✅ CORRECT: Use canNotChangeSkr for Immutable Fields

```go
// ALWAYS: Use canNotChangeSkr to verify immutability
Context("Immutability Validation", func() {
    
    canNotChangeSkr(
        "GcpNfsVolume BASIC_SSD tier capacity cannot be reduced",
        newTestGcpNfsVolumeBuilder().
            WithTier(cloudresourcesv1beta1.BASIC_SSD).
            WithCapacityGb(2816).
            WithFileShareName("valid-name"),
        func(b Builder[*cloudresourcesv1beta1.GcpNfsVolume]) {
            b.(*testGcpNfsVolumeBuilder).WithCapacityGb(2560)
        },
        "BASIC_SSD tier capacityGb cannot be reduced",
    )
    
    canNotChangeSkr(
        "GcpNfsVolume tier cannot be changed",
        newTestGcpNfsVolumeBuilder().
            WithTier(cloudresourcesv1beta1.REGIONAL).
            WithCapacityGb(1024).
            WithFileShareName("valid-name"),
        func(b Builder[*cloudresourcesv1beta1.GcpNfsVolume]) {
            b.(*testGcpNfsVolumeBuilder).WithTier(cloudresourcesv1beta1.BASIC_SSD)
        },
        "tier is immutable",
    )
    
    canNotChangeSkr(
        "GcpNfsVolume fileShareName cannot be changed",
        newTestGcpNfsVolumeBuilder().
            WithTier(cloudresourcesv1beta1.REGIONAL).
            WithCapacityGb(1024).
            WithFileShareName("original-name"),
        func(b Builder[*cloudresourcesv1beta1.GcpNfsVolume]) {
            b.(*testGcpNfsVolumeBuilder).WithFileShareName("new-name")
        },
        "fileShareName is immutable",
    )
})
```

## Testing Cross-Field Validation

### ❌ WRONG: Not Testing Field Dependencies

```go
// NEVER: Only testing individual field validation
canNotCreateSkr("invalid capacity: 1025", builder.WithCapacityGb(1025), "invalid")
// WRONG: Not testing that validation depends on tier!
```

### ✅ CORRECT: Test Field Interactions

```go
// ALWAYS: Test cross-field validation rules
Context("Cross-Field Validation", func() {
    
    canNotCreateSkr(
        "GcpNfsVolume REGIONAL tier capacity not divisible by 256",
        newTestGcpNfsVolumeBuilder().
            WithTier(cloudresourcesv1beta1.REGIONAL).
            WithCapacityGb(1025).  // Not divisible by 256
            WithFileShareName("valid-name"),
        "REGIONAL tier capacityGb must be divisible by 256",
    )
    
    canNotCreateSkr(
        "GcpNfsVolume BASIC_HDD tier capacity not divisible by 256",
        newTestGcpNfsVolumeBuilder().
            WithTier(cloudresourcesv1beta1.BASIC_HDD).
            WithCapacityGb(1537).  // 1536 + 1, not divisible
            WithFileShareName("valid-name"),
        "BASIC_HDD tier capacityGb must be divisible by 256",
    )
    
    canCreateSkr(
        "GcpNfsVolume BASIC_SSD tier capacity does not need divisibility rule",
        newTestGcpNfsVolumeBuilder().
            WithTier(cloudresourcesv1beta1.BASIC_SSD).
            WithCapacityGb(2561).  // Not divisible by 256, but OK for BASIC_SSD
            WithFileShareName("valid-name"),
    )
})
```

## Testing Pattern Validation

### ❌ WRONG: Only Testing Valid Patterns

```go
// NEVER: Not testing invalid patterns
Context("Pattern Validation", func() {
    canCreateSkr("valid name", builder.WithFileShareName("valid-123"))
    // WRONG: No invalid pattern tests!
})
```

### ✅ CORRECT: Test Valid and Invalid Patterns

```go
// ALWAYS: Test pattern validation with valid and invalid cases
Context("Pattern Validation", func() {
    
    canCreateSkr(
        "fileShareName with valid pattern: lowercase-with-numbers-123",
        newTestGcpNfsVolumeBuilder().
            WithTier(cloudresourcesv1beta1.REGIONAL).
            WithCapacityGb(1024).
            WithFileShareName("lowercase-with-numbers-123"),
    )
    
    canNotCreateSkr(
        "fileShareName with uppercase letters",
        newTestGcpNfsVolumeBuilder().
            WithTier(cloudresourcesv1beta1.REGIONAL).
            WithCapacityGb(1024).
            WithFileShareName("Invalid-Name"),
        "fileShareName must match pattern",
    )
    
    canNotCreateSkr(
        "fileShareName with special characters",
        newTestGcpNfsVolumeBuilder().
            WithTier(cloudresourcesv1beta1.REGIONAL).
            WithCapacityGb(1024).
            WithFileShareName("name_with_underscores"),
        "fileShareName must match pattern",
    )
    
    canNotCreateSkr(
        "fileShareName too long",
        newTestGcpNfsVolumeBuilder().
            WithTier(cloudresourcesv1beta1.REGIONAL).
            WithCapacityGb(1024).
            WithFileShareName("a" + strings.Repeat("b", 100)),  // Over limit
        "fileShareName exceeds maximum length",
    )
})
```

## Testing Boundary Conditions

### ❌ WRONG: Not Testing Min/Max

```go
// NEVER: Only testing mid-range values
canCreateSkr("valid capacity", builder.WithCapacityGb(5000))
// WRONG: No min/max boundary tests!
```

### ✅ CORRECT: Test Min, Max, and Edge Cases

```go
// ALWAYS: Test boundary conditions (min, max, just outside range)
Context("Boundary Conditions", func() {
    
    canCreateSkr(
        "REGIONAL tier with minimum capacity: 1024",
        newTestGcpNfsVolumeBuilder().
            WithTier(cloudresourcesv1beta1.REGIONAL).
            WithCapacityGb(1024).  // Minimum
            WithFileShareName("valid-name"),
    )
    
    canCreateSkr(
        "REGIONAL tier with maximum capacity: 9984",
        newTestGcpNfsVolumeBuilder().
            WithTier(cloudresourcesv1beta1.REGIONAL).
            WithCapacityGb(9984).  // Maximum
            WithFileShareName("valid-name"),
    )
    
    canNotCreateSkr(
        "REGIONAL tier with capacity below minimum: 1023",
        newTestGcpNfsVolumeBuilder().
            WithTier(cloudresourcesv1beta1.REGIONAL).
            WithCapacityGb(1023).  // Below minimum
            WithFileShareName("valid-name"),
        "capacityGb must be at least 1024",
    )
    
    canNotCreateSkr(
        "REGIONAL tier with capacity above maximum: 9985",
        newTestGcpNfsVolumeBuilder().
            WithTier(cloudresourcesv1beta1.REGIONAL).
            WithCapacityGb(9985).  // Above maximum
            WithFileShareName("valid-name"),
        "capacityGb must not exceed 9984",
    )
})
```

## Testing KCP Resources

### ✅ KCP Resource Test Pattern

```go
// KCP resources follow same pattern with KCP-specific helpers
package apitests

var _ = Describe("Feature: KCP AzureRedisEnterprise", Ordered, func() {
    
    Context("Creation Validation", func() {
        
        canCreateKcp(
            "AzureRedisEnterprise with valid SKU: Standard",
            newTestAzureRedisEnterpriseBuilder().
                WithScope("test-scope").
                WithSku("Standard").
                WithCapacity(2),
        )
        
        canNotCreateKcp(
            "AzureRedisEnterprise with invalid SKU",
            newTestAzureRedisEnterpriseBuilder().
                WithScope("test-scope").
                WithSku("InvalidSku").
                WithCapacity(2),
            "sku must be one of: Basic, Standard, Premium",
        )
    })
    
    Context("Immutability Validation", func() {
        
        canNotChangeKcp(
            "AzureRedisEnterprise scope cannot be changed",
            newTestAzureRedisEnterpriseBuilder().
                WithScope("original-scope").
                WithSku("Standard").
                WithCapacity(2),
            func(b Builder[*cloudcontrolv1beta1.AzureRedisEnterprise]) {
                b.(*testAzureRedisEnterpriseBuilder).WithScope("new-scope")
            },
            "scope is immutable",
        )
    })
})
```

## Running API Validation Tests

```bash
# Run all API validation tests
go test ./internal/api-tests -v

# Run specific test file
go test ./internal/api-tests -run TestAPIs -ginkgo.focus="GcpNfsVolume"

# Run with verbose output
go test ./internal/api-tests -v -ginkgo.v

# Run specific context
go test ./internal/api-tests -ginkgo.focus="Creation Validation"
```

## Common Pitfalls

### Pitfall 1: Test Passes But Validation Not Implemented

**Symptom**: `canNotCreateSkr` test passes but resource is created

**Cause**: Validation rule not actually implemented in CRD

**Solution**: 
1. Add kubebuilder markers to CRD: `// +kubebuilder:validation:...`
2. Run `make manifests` to regenerate CRD
3. Or implement webhook validation in `api/.../webhook.go`

### Pitfall 2: Expected Error Message Doesn't Match

**Symptom**: Test fails with "expected error not found"

**Cause**: Expected error string doesn't match actual error

**Solution**: Run test verbosely to see actual error:
```bash
go test ./internal/api-tests -v -ginkgo.v -ginkgo.focus="your test"
```
Update expected error message to match actual.

### Pitfall 3: Testing Controller Logic Instead of Validation

**Symptom**: Test tries to verify reconciliation, not just validation

**Cause**: Confusing API validation tests with controller tests

**Solution**: API validation tests ONLY test:
- Can resource be created?
- Can field be updated?
- Is validation rule enforced?

NOT:
- Does reconciler create cloud resources?
- Does status get updated?

## Validation Checklist

### For Each Field
- [ ] Valid values accepted (`canCreateSkr`)
- [ ] Invalid values rejected (`canNotCreateSkr`)
- [ ] Boundary conditions tested (min, max, edge cases)
- [ ] Pattern validation tested (if regex)
- [ ] Enum validation tested (if enum type)
- [ ] Immutability tested (`canChangeSkr` or `canNotChangeSkr`)
- [ ] Cross-field rules tested (if dependencies exist)

### Test Structure
- [ ] Builder pattern implemented
- [ ] Fluent setter methods for each field
- [ ] Minimal valid defaults in builder constructor
- [ ] Tests grouped by Context (Creation/Update/Immutability)
- [ ] Descriptive test names explaining validation

### Helper Usage
- [ ] Uses `canCreateSkr` for valid creation
- [ ] Uses `canNotCreateSkr` with expected error for invalid
- [ ] Uses `canChangeSkr` for mutable fields
- [ ] Uses `canNotChangeSkr` with expected error for immutable

### Error Messages
- [ ] All `canNotCreate*` have expected error message
- [ ] All `canNotChange*` have expected error message
- [ ] Error messages are clear and user-friendly
- [ ] Error messages match actual validation messages

## Summary: Key Rules

1. **Test Valid + Invalid**: BOTH scenarios REQUIRED for each validation rule
2. **Builder Pattern**: Use fluent builder for test resource construction
3. **Helper Functions**: Use `canCreateSkr/canNotCreateSkr/canChangeSkr/canNotChangeSkr`
4. **Expected Errors**: All negative tests MUST specify expected error message
5. **Boundary Conditions**: Test min, max, just-below-min, just-above-max
6. **Cross-Field Rules**: Test field dependencies explicitly
7. **Immutability**: Verify immutable fields reject updates with `canNotChangeSkr`
8. **Pattern Validation**: Test valid and invalid regex patterns

## Next Steps

- [Write Controller Tests](CONTROLLER_TESTS.md)
- [Create Mocks](CREATING_MOCKS.md)
- [Configure Feature Flags](FEATURE_FLAGS.md)
