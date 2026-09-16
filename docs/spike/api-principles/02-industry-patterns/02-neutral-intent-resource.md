# Pattern 2: Neutral Intent Resource

**Origin:**
- Kubernetes Storage — `PersistentVolumeClaim` · [k8s.io/api/core/v1/types.go](https://github.com/kubernetes/api/blob/master/core/v1/types.go#L533)
- Gardener External DNS — `DNSEntry` · [gardener/external-dns-management](https://github.com/gardener/external-dns-management/blob/master/pkg/apis/dns/v1alpha1/dnsentry.go)

**What it solves:** User expresses *what* they need in provider-neutral vocabulary. The controller routes to the right backend at runtime. The user never writes a provider name or provider-specific units.

---

## Reference examples

**Kubernetes PVC — no storage backend mentioned:**
```yaml
kind: PersistentVolumeClaim
spec:
  accessModes: [ReadWriteOnce]
  resources:
    requests:
      storage: 10Gi        # k8s Quantity — same field on EBS, GCP PD, Azure Disk
  storageClassName: standard
```

**Gardener DNSEntry — no DNS provider mentioned:**
```yaml
kind: DNSEntry
spec:
  dnsName: "my-service.example.com"
  ttl: 120
  targets: ["1.2.3.4"]    # same fields regardless of Route53, Cloud DNS, or Azure DNS
```

---

## Cloud Manager — IpRange (already correct)

```yaml
kind: IpRange
spec:
  cidr: "10.250.0.0/22"   # works on AWS, GCP, Azure, Alicloud, OpenStack unchanged
```

No change needed. `IpRange` is the ideal expression of this pattern in CM.

---

## Cloud Manager — NfsVolume capacity (needs fix)

`capacity` expresses the same user decision on every provider — how much storage. GCP and SAP leak their API's integer unit into the user spec instead of using the k8s Quantity that every other storage resource in Kubernetes uses.

**Before:**
```yaml
kind: GcpNfsVolume
spec:
  capacityGb: 1024          # integer — GCP Filestore API unit in user spec

---
kind: SapNfsVolume
spec:
  capacityGb: 100           # integer — OpenStack Manila API unit in user spec

---
kind: AwsNfsVolume
spec:
  capacity: "1Ti"           # k8s Quantity — correct
```

**After:**
```yaml
kind: GcpNfsVolume
spec:
  capacity: "1Ti"           # was: capacityGb: 1024
                            # controller converts to integer for Filestore API

---
kind: SapNfsVolume
spec:
  capacity: "100Gi"         # was: capacityGb: 100
                            # controller converts to integer for Manila API

---
kind: AwsNfsVolume
spec:
  capacity: "1Ti"           # unchanged
```

User learns one field name and one unit from standard PVC/StorageClass documentation, and it works on every provider.
