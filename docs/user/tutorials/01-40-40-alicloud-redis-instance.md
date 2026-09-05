# Using AlicloudRedisInstance Custom Resources

> [!WARNING]
> This is a beta feature available only per request for SAP-internal teams.

The Cloud Manager module offers an AlicloudRedisInstance Custom Resource Definition (CRD). When you apply an AlicloudRedisInstance custom resource (CR), it creates an Alibaba Cloud ApsaraDB for Redis (Tair) instance that is reachable within your Kubernetes cluster network.

## Prerequisites  <!-- {docsify-ignore} -->

You have the Cloud Manager module added.

## Steps

### Minimal Setup

This example showcases how to instantiate Redis using only the required fields, connect a Pod to it, and send a PING command.

1. Create a Redis instance. The operation may take more than 10 minutes.

   ```yaml
   apiVersion: cloud-resources.kyma-project.io/v1beta1
   kind: AlicloudRedisInstance
   metadata:
     name: alicloudredisinstance-simple-example
   spec:
     redisTier: "S1"
   ```

2. Wait for the instance to become ready.

   ```bash
   kubectl wait --for=condition=Ready alicloudredisinstance/alicloudredisinstance-simple-example --timeout=1200s
   ```

3. Instantiate the redis-cli Pod.

   ```yaml
   apiVersion: v1
   kind: Pod
   metadata:
     name: alicloudredisinstance-simple-example-probe
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
             name: alicloudredisinstance-simple-example
       - name: PORT
         valueFrom:
           secretKeyRef:
             key: port
             name: alicloudredisinstance-simple-example
       - name: AUTH_STRING
         valueFrom:
           secretKeyRef:
             key: authString
             name: alicloudredisinstance-simple-example
       volumeMounts:
       - name: mounted
         mountPath: /mnt
     volumes:
     - name: mounted
       secret:
         secretName: alicloudredisinstance-simple-example
   ```

4. Execute into the Pod.

   ```bash
   kubectl exec -i -t alicloudredisinstance-simple-example-probe -c redis-cli -- sh -c "clear; (bash || ash || sh)"
   ```

5. Run a PING command.

   ```bash
   redis-cli -h $HOST -p $PORT -a $AUTH_STRING --tls --cacert /mnt/CaCert.pem PING
   ```

   You should receive `PONG` back from the server.

### Advanced Setup

This example showcases how to instantiate Redis by using most of the spec fields, connect a Pod to it, and send a PING command.

1. Instantiate Redis. The operation may take more than 10 minutes.

   ```yaml
   apiVersion: cloud-resources.kyma-project.io/v1beta1
   kind: AlicloudRedisInstance
   metadata:
     name: alicloudredisinstance-complex-example
   spec:
     redisTier: "P1"
     engineVersion: "7.0"
     authSecret:
       name: custom-redis-secret
       labels:
         app: my-app
   ```

2. Wait for the instance to become ready.

   ```bash
   kubectl wait --for=condition=Ready alicloudredisinstance/alicloudredisinstance-complex-example --timeout=1200s
   ```

3. Instantiate the redis-cli Pod.

   ```yaml
   apiVersion: v1
   kind: Pod
   metadata:
     name: alicloudredisinstance-complex-example-probe
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
             name: custom-redis-secret
       - name: PORT
         valueFrom:
           secretKeyRef:
             key: port
             name: custom-redis-secret
       - name: AUTH_STRING
         valueFrom:
           secretKeyRef:
             key: authString
             name: custom-redis-secret
       volumeMounts:
       - name: mounted
         mountPath: /mnt
     volumes:
     - name: mounted
       secret:
         secretName: custom-redis-secret
   ```

4. Execute into the Pod.

   ```bash
   kubectl exec -i -t alicloudredisinstance-complex-example-probe -c redis-cli -- sh -c "clear; (bash || ash || sh)"
   ```

5. Run a PING command.

   ```bash
   redis-cli -h $HOST -p $PORT -a $AUTH_STRING --tls --cacert /mnt/CaCert.pem PING
   ```

   You should receive `PONG` back from the server.
