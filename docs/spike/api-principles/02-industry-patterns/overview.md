# Industry Patterns — Applied to Cloud Manager CRD API

Each pattern: where it comes from (with repo reference), what it solves, before/after for a CM CRD.

| # | Pattern | Source | Applies to CM |
|---|---------|--------|---------------|
| [1](./01-base-with-extensions.md) | Common base + provider-specific extensions | CAPI, Crossplane, Gateway API | All CRD families |
| [2](./02-neutral-intent-resource.md) | Neutral intent resource | PVC, Gardener DNSEntry | NfsVolume `capacity` field |
| [3](./03-normalize-status.md) | Consistent field naming for shared concepts | Crossplane provider mapping, k8s API conventions | Redis `engineVersion`, `replicasPerShard` |
| [4](./04-provider-specific-resource.md) | Provider-specific resource | ACK, ASO | VpcPeering (correct as-is) |
| [5](./05-typed-provider-sub-struct.md) | Typed provider sub-struct | Crossplane `forProvider`, CAPI infrastructure contract | KCP status contract completeness |
| [6](./06-unstructured-payload.md) | Portable container + unstructured payload | StorageClass `parameters`, Gateway API `parametersRef` | Future WafPolicy rules |
