# Initializing AwsNfsVolume for Non-Privileged Workloads

This tutorial explains how to initialize an AwsNfsVolume so that workloads running with a non-root user ID (UID) and group ID (GID) can write to the mounted volume.

## Context

AwsNfsVolume is backed by Amazon Elastic File System (EFS). Amazon Web Services (AWS) provisions the EFS filesystem root directory with the permissions `drwxr-xr-x`, owned by UID `0` and GID `0` (`root:root`). This means only the root user can create files or directories directly under the root.

Workloads that run with a non-root UID/GID cannot write to the volume root and receive a permission error when their Pods try to create files or directories under the mount path.

To work around this constraint, run a one-time initialization Pod as root. The Pod creates the required subdirectories and assigns ownership of the mounted volume to the UID/GID that your workload uses. After initialization, your non-privileged workloads can write under the mount path.

## Prerequisites <!-- {docsify-ignore} -->

* You have the Cloud Manager module added.
* You have created an AwsNfsVolume resource and a PersistentVolumeClaim that binds to it. See [Using NFS in Amazon Web Services](./01-20-10-aws-nfs-volume.md).

> [!NOTE]
> The examples below assume that the PersistentVolumeClaim is named `my-vol` and is in the same namespace as the initialization Pod. Replace `1000:1000` with the UID/GID that your workload uses.

## Steps <!-- {docsify-ignore} -->

1. Export the namespace as an environment variable. Run:

   ```shell
   export NAMESPACE={NAMESPACE_NAME}
   ```

2. Create a one-time initialization Pod that runs as root, creates the required subdirectories, and assigns their ownership to your workload's UID/GID.

   ```shell
   cat <<EOF | kubectl -n $NAMESPACE apply -f -
   apiVersion: v1
   kind: Pod
   metadata:
     name: one-off-nfs-init
   spec:
     restartPolicy: Never
     containers:
     - name: volume-initializer
       image: busybox:latest
       command: ["sh", "-c", "mkdir -p /mnt/data/app-dirs && chown -R 1000:1000 /mnt/data"]
       securityContext:
         runAsUser: 0
       volumeMounts:
       - name: data
         mountPath: /mnt/data
     volumes:
     - name: data
       persistentVolumeClaim:
         claimName: my-vol
   EOF
   ```

   Replace `1000:1000` with the UID/GID that your workload uses.

3. Wait for the initialization Pod to complete.

   ```shell
   kubectl -n $NAMESPACE wait --for=jsonpath='{.status.phase}'=Succeeded pod/one-off-nfs-init --timeout=120s
   ```

   Once the Pod completes, you should see the following message:

   ```console
   pod/one-off-nfs-init condition met
   ```

4. Your non-privileged workloads can now write to the initialized subdirectories under the mount path.

## Next Steps <!-- {docsify-ignore} -->

To clean up, remove the initialization Pod:

```shell
kubectl delete -n $NAMESPACE pod one-off-nfs-init
```
