# Using AlicloudRedisCluster Custom Resources

> [!WARNING]
> This is a beta feature available only per request for SAP-internal teams.

The Cloud Manager module offers an AlicloudRedisCluster Custom Resource Definition (CRD). When you apply an AlicloudRedisCluster custom resource (CR), it creates an Alibaba Cloud ApsaraDB for Redis cluster (proxy-based sharded) instance that is reachable within your Kubernetes cluster network.

## Prerequisites  <!-- {docsify-ignore} -->

You have the Cloud Manager module added.

## Steps

This example showcases how to instantiate a Redis cluster, connect a Pod to it, and send a PING command.

1. Create a Redis cluster. The operation may take more than 10 minutes.

   ```yaml
   apiVersion: cloud-resources.kyma-project.io/v1beta1
   kind: AlicloudRedisCluster
   metadata:
     name: alicloudrediscluster-simple-example
   spec:
     redisTier: "C3"
     shardCount: 3
   ```

2. Wait for the cluster to become ready.

   ```bash
   kubectl wait --for=condition=Ready alicloudrediscluster/alicloudrediscluster-simple-example --timeout=1200s
   ```

3. Instantiate the redis-cli Pod.

   ```yaml
   apiVersion: v1
   kind: Pod
   metadata:
     name: alicloudrediscluster-simple-example-probe
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
             name: alicloudrediscluster-simple-example
       - name: PORT
         valueFrom:
           secretKeyRef:
             key: port
             name: alicloudrediscluster-simple-example
       - name: AUTH_STRING
         valueFrom:
           secretKeyRef:
             key: authString
             name: alicloudrediscluster-simple-example
       volumeMounts:
       - name: mounted
         mountPath: /mnt
     volumes:
     - name: mounted
       secret:
         secretName: alicloudrediscluster-simple-example
   ```

4. Execute into the Pod.

   ```bash
   kubectl exec -i -t alicloudrediscluster-simple-example-probe -c redis-cli -- sh -c "clear; (bash || ash || sh)"
   ```

5. Run a PING command.

   ```bash
   redis-cli -h $HOST -p $PORT -a $AUTH_STRING --tls --cacert /mnt/CaCert.pem -c PING
   ```

   You should receive `PONG` back from the server.
