# GCP NfsInstance V2 Feature Flag

## Overview

The `gcpNfsInstanceV2` feature flag controls which implementation of the GCP NfsInstance reconciler is used:
- **v1** (default): Legacy implementation using `google.golang.org/api/file/v1` client
- **v2**: Streamlined implementation using modern `cloud.google.com/go/filestore` client

## Configuration

### Feature Flag Definition

**Name**: `gcpNfsInstanceV2`  
**Type**: Boolean  
**Default**: `false` (uses v1 implementation)

### Usage in Code

```go
import "github.com/kyma-project/cloud-manager/pkg/feature"

func (r *nfsInstanceReconciler) gcpAction(ctx context.Context, state composed.State) (error, context.Context) {
    if feature.GcpNfsInstanceV2.Value(ctx) {
        // Use v2 implementation
        return gcpnfsinstancev2.New(r.gcpStateFactoryV2)(ctx, state)
    }
    // Use v1 implementation (default)
    return gcpnfsinstancev1.New(r.gcpStateFactoryV1)(ctx, state)
}
```

### Configuration Files

#### GA Environment (`ff_ga.yaml`)
```yaml
gcpNfsInstanceV2:
  variations:
    enabled: true
    disabled: false
  defaultRule:
    variation: disabled  # Default to v1
```

#### Edge Environment (`ff_edge.yaml`)
```yaml
gcpNfsInstanceV2:
  variations:
    enabled: true
    disabled: false
  defaultRule:
    variation: disabled  # Default to v1
```

## Rollout Plan

### Phase 1: Initial Deployment (v1 active)
- Deploy with feature flag = `false` (default)
- V1 implementation active
- V2 code present but inactive

### Phase 2: Testing (enable v2 in dev/stage)
- Enable flag in dev environment
- Run integration tests
- Monitor metrics and error rates
- Validate v2 behavior matches v1

### Phase 3: Gradual Production Rollout
- Enable for subset of production clusters
- Monitor for issues
- Compare performance metrics
- Expand to more clusters if stable

### Phase 4: Full V2 Rollout
- Enable flag globally
- V2 becomes default implementation
- V1 remains for rollback capability

### Phase 5: V1 Deprecation (future)
- Add deprecation notices to v1
- Plan v1 removal
- Remove feature flag after sunset period

## Enabling V2

### For Testing (local development)
Set environment variable:
```bash
export FEATURE_FLAG_CONFIG_FILE="path/to/custom-config.yaml"
```

With custom config:
```yaml
gcpNfsInstanceV2:
  defaultRule:
    variation: enabled
```

### For Kubernetes Deployment
Update ConfigMap with feature flag configuration:
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: feature-flags
data:
  config.yaml: |
    gcpNfsInstanceV2:
      defaultRule:
        variation: enabled
```

Mount ConfigMap and set environment variable:
```yaml
env:
  - name: FEATURE_FLAG_CONFIG_FILE
    value: /etc/feature-flags/config.yaml
volumes:
  - name: feature-flags
    configMap:
      name: feature-flags
volumeMounts:
  - name: feature-flags
    mountPath: /etc/feature-flags
```

## Monitoring

When v2 is enabled, monitor:
- Reconciliation success/failure rates
- GCP API error rates
- Operation completion times
- Resource creation/update/deletion success
- Status condition accuracy

## Rollback

If issues occur with v2:
1. Set feature flag to `false` (reverts to v1)
2. Restart affected pods to reload configuration
3. Monitor for stabilization
4. Investigate v2 issues

## Differences: V1 vs V2

| Aspect | V1 | V2 |
|--------|----|----|
| **GCP Client** | `google.golang.org/api/file/v1` | `cloud.google.com/go/filestore` |
| **Client Type** | REST API wrapper | Modern gRPC client |
| **Code Structure** | 10+ separate action files | Organized in subdirectories |
| **State Hierarchy** | 3 layers (OLD pattern) | 3 layers (OLD pattern maintained) |
| **Validations** | Distributed across files | Consolidated in `validation/` |
| **Operations** | Mixed with state logic | Separated in `operations/` |
| **Testing** | Extensive (includes trivial tests) | Focused on business logic |

## References

- Feature flag implementation: [pkg/feature/ffGcpNfsInstanceV2.go](../ffGcpNfsInstanceV2.go)
- V1 implementation: `pkg/kcp/provider/gcp/nfsinstance/v1/`
- V2 implementation: `pkg/kcp/provider/gcp/nfsinstance/v2/` (in progress)
- Architecture documentation: [docs/agents/architecture/](../../../../docs/agents/architecture/)

## Support

For issues or questions:
1. Check v2 implementation documentation in `pkg/kcp/provider/gcp/nfsinstance/v2/README.md`
2. Review refactoring plan: [GCP_NFSINSTANCE_REFACTOR_PLAN.md](../../../../GCP_NFSINSTANCE_REFACTOR_PLAN.md)
3. Contact the cloud-manager team
