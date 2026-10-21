# Pattern 2: Neutral Intent Resource

The user expresses *what* they need using concepts they already know. The controller routes to the right provider at runtime. No provider name or provider-specific unit appears in the user-facing spec.

---

## Where This Pattern Comes From

### Kubernetes PersistentVolumeClaim
**Repo:** [kubernetes/api](https://github.com/kubernetes/api)
**Key file:** [`core/v1/types.go`](https://github.com/kubernetes/api/blob/master/core/v1/types.go#L533)

`PVC` expresses what the user needs (`storage: 10Gi`, `ReadWriteOnce`). `StorageClass` names the provisioner. The PV is the provisioned result owned by the controller. The user never writes `ebs.csi.aws.com`, `pd.csi.storage.gke.io`, or `disk.csi.azure.com`.

```yaml
kind: PersistentVolumeClaim
spec:
  accessModes: [ReadWriteOnce]
  resources:
    requests:
      storage: 10Gi       # k8s Quantity — same field on EBS, GCP PD, Azure Disk
  storageClassName: standard
```

**Lesson for CM:** `capacity` is a user concept. `capacityGb: 1024` is a GCP Filestore API unit. The controller converts — the user spec does not carry the provider's unit.

---

### Gardener External DNS — DNSEntry
**Repo:** [gardener/external-dns-management](https://github.com/gardener/external-dns-management)
**Key file:** [`pkg/apis/dns/v1alpha1/dnsentry.go`](https://github.com/gardener/external-dns-management/blob/master/pkg/apis/dns/v1alpha1/dnsentry.go)

`DNSEntry` has only `dnsName`, `ttl`, and `targets`. The controller matches the domain against `DNSProvider.spec.domains.include` at runtime. The user is completely oblivious to whether Route53, Cloud DNS, or Azure DNS handles the record.

```yaml
kind: DNSEntry
spec:
  dnsName: "my-service.example.com"
  ttl: 120
  targets: ["1.2.3.4"]    # same fields regardless of Route53, Cloud DNS, Azure DNS
```

**Lesson for CM:** `IpRange` already follows this pattern exactly. One neutral `cidr` field, provider routing at runtime. This is the target shape for any CM resource where the user decision is the same across providers.

---

## How Cloud Manager Uses This Pattern

**`IpRange` — already correct:**

```yaml
kind: IpRange
spec:
  cidr: "10.250.0.0/22"   # works on AWS, GCP, Azure, Alicloud, OpenStack unchanged
```

No change needed. `IpRange` is the ideal expression of this pattern in CM.
