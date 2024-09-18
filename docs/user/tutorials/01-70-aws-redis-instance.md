# AWS Redis Instance Tutorial
Learn how to instantiate Redis and connect to it.
## Simple Example

This example showcases how to instantiate Redis, connect a Pod to it, and send a PING command.

1. Instantiate Redis. This may take 10+ minutes.


   ```yaml
   apiVersion: cloud-resources.kyma-project.io/v1beta1
   kind: AwsRedisInstance
   metadata:
     name: awsredisinstance-simple-example
   spec:
     cacheNodeType: cache.t3.micro
   ```

2. Instantiate the redis-cli Pod:

   ```yaml
   apiVersion: v1
   kind: Pod
   metadata:
     name: awsredisinstance-simple-example-probe
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
   kubectl exec -i -t awsredisinstance-simple-example-probe -c redis-cli -- sh -c "clear; (bash || ash || sh)"
   ```

4. Run a PING command:

   ```bash
   redis-cli -h $HOST -p $PORT
   ```
   If your setup was successful, you get `PONG` back from the server.

## Complex Example

This example showcases how to instantiate Redis by using most of the spec fields, connect a Pod to it, and send a PING command.

1. Instantiate Redis. This may take 10+ minutes.

   ```yaml
   apiVersion: cloud-resources.kyma-project.io/v1beta1
   kind: AwsRedisInstance
   metadata:
     name: awsredisinstance-complex-example
   spec:
     cacheNodeType: cache.t3.micro
     engineVersion: "7.0"
     authEnabled: true
     transitEncryptionEnabled: true
     parameters:
       maxmemory-policy: volatile-lru
       activedefrag: "yes"
     preferredMaintenanceWindow: sun:23:00-mon:01:30
     autoMinorVersionUpgrade: true
   ```

2. Instantiate the redis-cli Pod:

   ```yaml
   apiVersion: v1
   kind: Pod
   metadata:
     name: awsredisinstance-complex-example-probe
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
             name: awsredisinstance-complex-example
       - name: PORT
         valueFrom:
           secretKeyRef:
             key: port
             name: awsredisinstance-complex-example
       - name: AUTH_STRING
         valueFrom:
           secretKeyRef:
             key: authString
             name: awsredisinstance-complex-example
   ```

3. Exec into the Pod:

   ```bash
   kubectl exec -i -t awsredisinstance-complex-example-probe -c redis-cli -- sh -c "clear; (bash || ash || sh)"
   ```

4. Install and update ca-certificates:

   ```bash
   apt-get update && \
     apt-get install -y ca-certificates && \
     update-ca-certificate
   ```

5. Run a PING command:

   ```bash
   redis-cli -h $HOST -p $PORT -a $AUTH_STRING --tls
   ```
   If your setup was successful, you get `PONG` back from the server.
