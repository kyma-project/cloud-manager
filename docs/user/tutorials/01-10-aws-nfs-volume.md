# AWS NFS Volume

This tutorial explains how to create and use RWX volume in AWS cloud provider that can be used from multiple workloads.

## Steps <!-- {docsify-ignore} -->

1. Create a namespace and export its value as an environment variable. Run:

   ```shell
   export NAMESPACE={NAMESPACE_NAME}
   kubectl create ns $NAMESPACE
   ```
   
2. Create AwsNfsVolume resource.

   ```shell
   cat <<EOF | kubectl -n $NAMESPACE apply -f -
   apiVersion: cloud-resources.kyma-project.io/v1beta1
   kind: AwsNfsVolume
   metadata:
     name: my-vol
   spec:
     capacity: 100G
   EOF
   ```
   
3. Wait for AwsNfsVolume to get Ready condition.

   ```shell
   kubectl -n $NAMESPACE wait --for=condition=Ready awsnfsvolume/my-vol --timeout=300s
   ```

   Once created AwsNfsVolume is provisioned this command should terminate with message:

   ```
   awsnfsvolume.cloud-resources.kyma-project.io/my-vol condition met
   ```
   
4. Observe the generated PersistentVolume:
   
   ```shell
   kubectl -n $NAMESPACE get persistentvolume my-vol
   ```
   
   This command should print:
   
   ```
   NAME     CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM            STORAGECLASS
   my-vol   100G       RWX            Retain           Bound    test-mt/my-vol               
   ```
   
   Note the RWX access mode which allows volume to be readable and writable from multiple workloads, and
   Bound status which means PersistentVolumeClaim claiming this PV is created.
   
5. Observe the generated PersistentVolumeClaim:

   ```shell
   kubectl -n $NAMESPACE get persistentvolumeclaim my-vol
   ```

   This command should print:

   ```
   NAME     STATUS   VOLUME   CAPACITY   ACCESS MODES   STORAGECLASS 
   my-vol   Bound    my-vol   100G       RWX                         
   ```

   Similarly to PV, note the RWX access mode and Bound status.

6. Create two workloads that both will write to the volume

   ```shell
   cat <<EOF | kubectl -n $NAMESPACE apply -f -
   apiVersion: v1
   kind: ConfigMap
   metadata:
     name: my-script
   data:
     my-script.sh: |
       #!/bin/bash
       for xx in {1..20}; do 
         echo "Hello from \$NAME: \$xx" | tee -a /mnt/data/test.log
         sleep 1
       done
       echo "File content:"
       cat /mnt/data/test.log
       sleep 864000
   ---
   apiVersion: apps/v1
   kind: Deployment
   metadata:
     name: awsnfsvolume-demo
   spec:
     selector:
       matchLabels:
         app: awsnfsvolume-demo
     replicas: 2
     template:
       metadata:
         labels:
           app: awsnfsvolume-demo
       spec:
         containers:
           - name: my-container
             image: ubuntu  
             command: 
               - "/bin/bash"
             args:
               - "/script/my-script.sh"
             env:
               - name: NAME
                 valueFrom:
                   fieldRef:
                     fieldPath: metadata.name
             volumeMounts:
               - name: my-script-volume
                 mountPath: /script
               - name: data
                 mountPath: /mnt/data
         volumes:
           - name: my-script-volume
             configMap:
               name: my-script
               defaultMode: 0744 
           - name: data
             persistentVolumeClaim:
               claimName: my-vol 
   EOF
   ``` 
   
   This workload will print a sequence of 20 lines to stdout and a file on the nfs volume.
   Then it will print the content of the file.
   
7. Print the logs of one of the workloads

   ```shell
   kubectl logs -n $NAMESPACE `kubectl get pod -n $NAMESPACE -l app=awsnfsvolume-demo -o=jsonpath='{.items[0].metadata.name}'`
   ```
   
   The command should print something like:
   ```
   ...
   Hello from awsnfsvolume-demo-869c89df4c-dsw97: 19
   Hello from awsnfsvolume-demo-869c89df4c-dsw97: 20
   File content:
   Hello from awsnfsvolume-demo-869c89df4c-8z9zl: 20
   Hello from awsnfsvolume-demo-869c89df4c-l8hrb: 1
   Hello from awsnfsvolume-demo-869c89df4c-dsw97: 1
   Hello from awsnfsvolume-demo-869c89df4c-l8hrb: 2
   Hello from awsnfsvolume-demo-869c89df4c-dsw97: 2
   Hello from awsnfsvolume-demo-869c89df4c-l8hrb: 3
   ...
   ```
   
   Note that the content after `File content:` contains prints from both workloads. This 
   demonstrates the ReadWriteMany capability of the volume.

8. Clean up

   Remove the created workloads:
   ```shell
   kubectl delete -n $NAMESPACE deployment awsnfsvolume-demo
   ```

   Remove the created configmap:
   ```shell
   kubectl delete -n $NAMESPACE configmap my-script
   ```

   Remove the created awsnfsvolume:
   ```shell
   kubectl delete -n $NAMESPACE awsnfsvolume my-vol
   ```

   Remove the created default iprange:

   > [!NOTE]
   > If you have other cloud resources using the default IpRange,
   > then you should skip this step, and not delete the default IpRange.

   ```shell
   kubectl delete -n kyma-system iprange default
   ```

   Remove the created namespace:
   ```shell
   kubectl delete namespace $NAMESPACE
   ```
