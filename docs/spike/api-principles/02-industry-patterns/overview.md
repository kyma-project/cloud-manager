# Industry Patterns — Applied to Cloud Manager CRD API

Each pattern: where it comes from (with repo reference), what it solves, before/after for a CM CRD.

| # | Pattern | Source | Applies to CM |
|---|---------|--------|---------------|
| [1](./01-base-with-extensions.md) | Common base + provider-specific extensions | CAPI, Crossplane, Gateway API | All CRD families |
| [2](./02-neutral-intent-resource.md) | Neutral intent resource | PVC, Gardener DNSEntry | NfsVolume `capacity` field |
| [3](./03-normalize-status.md) | Normalize status, keep provider spec | Crossplane Managed Resources | Redis `engineVersion`, `replicasPerShard` |
| [4](./04-provider-specific-resource.md) | Provider-specific resource | ACK, ASO | VpcPeering (correct as-is) |
| [5](./05-typed-provider-sub-struct.md) | Typed provider sub-struct | Crossplane Composition, CAPI | Redis `parameters` field name |
| [6](./06-unstructured-payload.md) | Portable container + unstructured payload | StorageClass `parameters` | Future WafPolicy rules |
