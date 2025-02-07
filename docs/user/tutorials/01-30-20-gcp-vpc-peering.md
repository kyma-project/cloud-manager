# Create Virtual Private Cloud Peering in Google Cloud

This tutorial explains how to create a Virtual Private Cloud (VPC) peering connection between a remote VPC network and Kyma in Google Cloud.

## Prerequisites  <!-- {docsify-ignore} -->

- You have the Cloud Manager module added.
- Use a POSIX-compliant shell or adjust the commands accordingly. For example, if you use Windows, replace the `export` commands with `set` and use `%` before and after the environment variables names.

## Steps <!-- {docsify-ignore} -->

1. Fetch your Kyma ID and save it as an environment variable KYMA_SHOOT_ID.

    ```shell
   export KYMA_SHOOT_ID=`kubectl get cm -n kube-system shoot-info -o jsonpath='{.data.shootName}'`
   ```

2. Save your project ID and VPC network name into environment variables, replace the placeholders with the correct values.

    ```shell
     export REMOTE_PROJECT_ID={YOUR_REMOTE_PROJECT_ID}
     export REMOTE_VPC_NETWORK={REMOTE_VPC_NETWORK}
     ```

3. Create a tag key with the Kyma shoot name in the remote project.

   ```shell
   gcloud resource-manager tags keys create $KYMA_SHOOT_ID --parent=projects/$REMOTE_PROJECT_ID
   ```

4. Create the tag value in the remote project.

    ```shell
    gcloud resource-manager tags values create None --tag-key=$REMOTE_PROJECT_ID/$KYMA_SHOOT_ID
    ```

5. Fetch the network selfLinkWithId from the remote vpc network.

    ```shell
    gcloud compute networks describe $REMOTE_VPC_NETWORK
    ```

    The command returns an output similar to this one:

    ```shell
    ...
    routingConfig:
    routingMode: REGIONAL
    selfLink: https://www.googleapis.com/compute/v1/projects/remote-project-id/global/networks/remote-vpc
    selfLinkWithId: https://www.googleapis.com/compute/v1/projects/remote-project-id/global/networks/1234567890123456789
    subnetworks:
    - https://www.googleapis.com/compute/v1/projects/remote-project-id/regions/europe-west12/subnetworks/remote-vpc
    ...
    ```

6. Export resource ID environment variable. Use the value of `selfLinkWithId` returned in the previous command's output, but replace `https://www.googleapis.com/compute/v1` with `//compute.googleapis.com`.

    ```shell
    export RESOURCE_ID="//compute.googleapis.com/projects/remote-project-id/global/networks/1234567890123456789"
    ```

7. Add the tag to the VPC network.

    ```shell
    gcloud resource-manager tags bindings create --tag-value=$REMOTE_PROJECT_ID/$KYMA_SHOOT_ID/None --parent=$RESOURCE_ID
    ```

8. Create a GCP VPC Peering manifest file.

    ```shell
    cat <<EOF > vpc-peering.yaml
    apiVersion: cloud-resources.kyma-project.io/v1beta1
    kind: GcpVpcPeering
    metadata:
        name: "vpcpeering-dev"
    spec:
        remotePeeringName: "my-project-to-kyma-dev"
        remoteProject: "$REMOTE_PROJECT_ID"
        remoteVpc: "$REMOTE_VPC_NETWORK"
        importCustomRoutes: false
    EOF
    ```

9. Apply the Google Cloud VPC peering manifest file.

    ```shell
    kubectl apply -f vpc-peering.yaml
    ```

    This operation usually takes less than 2 minutes. To check the status of the VPC peering, run:

    ```shell
    kubectl get gcpvpcpeering vpcpeering-dev -o yaml
    ```

    The command returns an output similar to this one:

    ```yaml
    apiVersion: cloud-resources.kyma-project.io/v1beta1
    kind: GcpVpcPeering
      finalizers:
      - cloud-control.kyma-project.io/deletion-hook
        generation: 2
        name: vpcpeering-dev
        resourceVersion: "12345678"
        uid: 8545cdaa-66d3-4fa7-b20b-7c716148552f
        spec:
        remotePeeringName: my-project-to-kyma-dev
        remoteProject: remote-project-id
        remoteVpc: remote-vpc-network
        status:
        conditions:
        - lastTransitionTime: "2024-08-12T15:29:59Z"
          message: VpcPeering: my-project-to-kyma-dev is provisioned
          reason: Ready
          status: "True"
          type: Ready
    ```

    The **status.conditions** field contains information about the VPC Peering status.


## Removing the VPC Peering

When the VPC peering is not needed anymore, it can be removed by deleting the GcpVpcPeering resource:

```shell
kubectl delete gcpvpcpeering vpcpeering-dev
```

Once the gcpvpcpeering object is deleted, it is possible to remove the inactive vpc peering from the remote project by executing the following command:

```shell
gcloud compute networks peerings delete my-project-to-kyma-dev --network=remote-vpc-network --project=remote-project-id
```