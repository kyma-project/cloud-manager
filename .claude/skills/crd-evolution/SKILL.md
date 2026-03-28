---
name: crd-evolution
description: Understand CRD evolution and migration patterns. Use when learning about NEW vs OLD patterns, understanding why patterns changed, or planning migrations.
---

# CRD Evolution

Understand why Cloud Manager patterns evolved and how to work with both.

## Pattern Overview

### NEW Pattern (Required for new code)

- **CRD Name**: Provider-specific (`GcpSubnet`, `AzureVNetLink`)
- **State**: 2 layers (focal → provider)
- **Location**: `pkg/kcp/provider/<provider>/<resource>/`
- **Use for**: All new resources (post-2024)

### OLD Pattern (Maintenance only)

- **CRD Name**: Multi-provider (`RedisInstance`, `NfsInstance`, `IpRange`)
- **State**: 3 layers (focal → shared → provider)
- **Location**: `pkg/kcp/<resource>/` + `pkg/kcp/provider/<provider>/<resource>/`
- **Use for**: Only maintaining existing resources

## Why Patterns Changed

### Problems with OLD Pattern

1. **Complex state hierarchy**: 3 layers increased coupling
2. **Provider switching**: `BuildSwitchAction` added complexity
3. **Shared code confusion**: Which layer owns what?
4. **Testing difficulty**: More layers = more mocks

### Benefits of NEW Pattern

1. **Simpler state**: Direct focal → provider
2. **Isolated providers**: No switching logic
3. **Clear ownership**: Provider package owns everything
4. **Easier testing**: Fewer layers to mock

## Identifying Patterns

| Characteristic | NEW Pattern | OLD Pattern |
|---------------|-------------|-------------|
| CRD name | `GcpSubnet` | `RedisInstance` |
| Spec structure | Direct fields | `.Spec.Instance.Gcp/Azure/Aws` |
| Package location | `pkg/kcp/provider/<provider>/<resource>/` | `pkg/kcp/<resource>/` + providers |
| State layers | 2 | 3 |
| Reconciler | Direct composition | `BuildSwitchAction` |

## Code Examples

### NEW Pattern Identification

```go
// CRD - provider-specific name
type GcpSubnet struct {
    Spec   GcpSubnetSpec   `json:"spec"`
    Status GcpSubnetStatus `json:"status"`
}

// State - extends focal directly
type State struct {
    focal.State
    subnetClient SubnetClient
    subnet       *compute.Subnet
}

// Reconciler - direct composition
composed.ComposeActions("main",
    loadSubnet,
    createSubnet,
    updateStatus,
)
```

### OLD Pattern Identification

```go
// CRD - multi-provider name
type RedisInstance struct {
    Spec   RedisInstanceSpec   `json:"spec"`
}

type RedisInstanceSpec struct {
    Instance RedisInstanceInfo `json:"instance"`
}

type RedisInstanceInfo struct {
    Gcp   *RedisInstanceGcp   `json:"gcp,omitempty"`
    Azure *RedisInstanceAzure `json:"azure,omitempty"`
    Aws   *RedisInstanceAws   `json:"aws,omitempty"`
}

// Reconciler - provider switching
composed.BuildSwitchAction("provider-switch",
    composed.NewCase(IsGcp, gcpReconciler),
    composed.NewCase(IsAzure, azureReconciler),
    composed.NewCase(IsAws, awsReconciler),
)
```

## Decision Tree

```
Is this a NEW resource?
├─ YES: MUST use NEW Pattern
│   └─ Create: pkg/kcp/provider/<provider>/<resource>/
│
└─ NO: Is it RedisInstance, NfsInstance, or IpRange?
    ├─ YES: Use OLD Pattern (maintenance)
    │   └─ Location: pkg/kcp/<resource>/
    │
    └─ NO: Check CRD name
        ├─ Provider-specific name → NEW Pattern
        └─ Multi-provider name → OLD Pattern
```

## Resources by Pattern

### NEW Pattern Resources
- `GcpSubnet`
- `GcpVpcPeering`
- `AzureVNetLink`
- All resources created after 2024

### OLD Pattern Resources (Maintenance)
- `RedisInstance`
- `NfsInstance`
- `IpRange`

## Migration Considerations

**DO NOT migrate existing OLD pattern resources** unless explicitly approved. Migration involves:

1. Creating new provider-specific CRDs
2. Writing data migration
3. Updating all consumers
4. Deprecation period
5. Removal of old CRDs

This is a significant undertaking requiring project-wide coordination.

## Working with OLD Pattern

When maintaining OLD pattern resources:

1. Follow existing patterns in the package
2. Don't introduce NEW pattern concepts
3. Keep changes minimal
4. Test all affected providers
5. Document any quirks discovered

## Related

- Full comparison: [docs/agents/architecture/RECONCILER_PATTERN_COMPARISON.md](../../../docs/agents/architecture/RECONCILER_PATTERN_COMPARISON.md)
- NEW pattern: [docs/agents/architecture/RECONCILER_NEW_PATTERN.md](../../../docs/agents/architecture/RECONCILER_NEW_PATTERN.md)
- OLD pattern: [docs/agents/architecture/RECONCILER_OLD_PATTERN.md](../../../docs/agents/architecture/RECONCILER_OLD_PATTERN.md)
