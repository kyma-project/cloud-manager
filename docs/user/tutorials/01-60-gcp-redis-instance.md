# GCP Redis Instance Tutorial
This tutorial shows how to instantiate Redis and connect to it.

## Simple Example

This example showcases how to instantiate Redis, connect a Pod to it, and send a PING command.

1. Instantiate Redis

> [!NOTE]
> This action may take 10+ min

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: GcpRedisInstance
metadata:
  name: gcpredisinstance-simple-example
spec:
  memorySizeGb: 1
  tier: "BASIC"
```

2. Instantiate the redis-cli Pod

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: gcpredisinstance-simple-example-probe
spec:
  containers:
  - name: redis-cli
    image: redis:latest
    command: ["/bin/sleep"]
    args: ["999999999999"]
    env:
    - name: HOST
      valueFrom:
        secretKeyRef:
          key: host
          name: simple-example-probe
    - name: PORT
      valueFrom:
        secretKeyRef:
          key: port
          name: simple-example-probe
```

3. Exec into the Pod:

```bash
kubectl exec -i -t gcpredisinstance-simple-example-probe -c redis-cli -- sh -c "clear; (bash || ash || sh)"
```

4. Exec a PING command:

```bash
redis-cli -h $HOST -p $PORT
```
You should receive `PONG` back from the server.

## Complex Example

This example showcases how to instantiate Redis by using most of the spec fields, connect a Pod to it, and send a PING command.

1. Instantiate Redis

> [!NOTE]
> This action may take 10+ min
```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: GcpRedisInstance
metadata:
  name: gcpredisinstance-complex-example
spec:
  memorySizeGb: 5
  tier: "STANDARD_HA"
  redisVersion: REDIS_7_1
  authEnabled: true
  transitEncryption:
    serverAuthentication: true
  redisConfigs:
    maxmemory-policy: volatile-lru
    activedefrag: "yes"
  maintenancePolicy:
    dayOfWeek:
      day: "SATURDAY"
      startTime:
          hours: 15
          minutes: 45
```

2. Instantiate the redis-cli Pod

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: gcpredisinstance-complex-example-probe
spec:
  containers:
  - name: redis-cli
    image: redis:latest
    command: ["/bin/sleep"]
    args: ["999999999999"]
    env:
    - name: HOST
      valueFrom:
        secretKeyRef:
          key: host
          name: gcpredisinstance-complex-example
    - name: PORT
      valueFrom:
        secretKeyRef:
          key: port
          name: gcpredisinstance-complex-example
    - name: AUTH_STRING
      valueFrom:
        secretKeyRef:
          key: authString
          name: gcpredisinstance-complex-example
    volumeMounts:
    - name: mounted
      mountPath: /mnt
  volumes:
  - name: mounted
    secret:
      secretName: gcpredisinstance-complex-example
```

3. Exec into the Pod:

```bash
kubectl exec -i -t gcpredisinstance-complex-example-probe -c redis-cli -- sh -c "clear; (bash || ash || sh)"
```

4. Exec a PING command:

```bash
redis-cli -h $HOST -p $PORT -a $AUTH_STRING --tls --cacert /mnt/CaCert.pem PING
```
You should receive `PONG` back from the server.