# Industry Patterns — Applied to Cloud Manager CRD API

Each pattern: where it comes from (with repo reference), what it solves, before/after for a CM CRD.

| # | Pattern                                                                    | Source | Applies to CM |
|---|----------------------------------------------------------------------------|--------|---------------|
| [1](./01-base-with-extensions.md) | [Common base + provider-specific extensions](./01-base-with-extensions.md) | CAPI, Crossplane, Gateway API | All CRD families |
| [2](./02-neutral-intent-resource.md) | [Neutral intent resource](./02-neutral-intent-resource.md)                 | PVC, Gardener DNSEntry | NfsVolume `capacity` field |
| [3](./03-normalize-status.md) | [Consistent field naming for shared concepts](./03-normalize-status.md)    | Crossplane provider mapping, k8s API conventions | Redis `engineVersion`, `replicasPerShard` |
| [4](./04-provider-specific-resource.md) | [Provider-specific resource](./04-provider-specific-resource.md)           | ACK, ASO | VpcPeering (correct as-is) |
| [5](./05-typed-provider-sub-struct.md) | [Typed provider sub-struct](./05-typed-provider-sub-struct.md)             | Crossplane `forProvider`, CAPI infrastructure contract | KCP NfsInstance sub-struct alignment |
| [6](./06-unstructured-payload.md) | [Portable container + unstructured payload](./06-unstructured-payload.md)               | StorageClass `parameters`, k8s `x-kubernetes-preserve-unknown-fields` | Future WafPolicy rules |
