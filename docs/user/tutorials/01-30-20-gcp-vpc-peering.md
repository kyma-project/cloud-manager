# Create Virtual Private Cloud Peering in Google Cloud

This tutorial explains how to create a Virtual Private Cloud (VPC) peering connection between a remote VPC network and Kyma in Google Cloud.

## Prerequisites  <!-- {docsify-ignore} -->

- You have the Cloud Manager module added.
- Use a POSIX-compliant shell or adjust the commands accordingly. For example, if you use Windows, replace the `export` commands with `set` and use `%` before and after the environment variables names.

## Steps <!-- {docsify-ignore} -->

1. Fetch your Kyma ID.

    ```shell
   kubectl get cm -n kube-system shoot-info -o jsonpath='{.data.shootName}'
   ```

2. Replace the placeholder with the fetched Kyma ID and export it as an environment variable.

   ```shell
    export KYMA_SHOOT_ID={YOUR_KYMA_ID}
    ```

3. Replace the placeholder with your project ID and export it as an environment variable.

    ```shell
     export REMOTE_PROJECT_ID={YOUR_REMOTE_PROJECT_ID}
     ```

4. Create a tag key with the Kyma shoot name in the remote project.

   > [!NOTE]  
   > Due to security reasons, the VPC network in the remote project, which receives the VPC peering connection, must contain a tag with the Kyma shoot name.

   ```shell
   gcloud resource-manager tags keys create $KYMA_SHOOT_ID --parent=projects/$REMOTE_PROJECT_ID
   ```

5. Fetch the tag created in the previous step.

   ```shell
   gcloud resource-manager tags keys list --parent=projects/$REMOTE_PROJECT_ID
   ```

   The command returns an output similar to this one:

   ```console
   NAME                     SHORT_NAME                DESCRIPTION
   tagKeys/123456789012345  shoot--kyma-dev--abc1234
   ```

6. Replace the `tagKeys/123456789012345` placeholder with your tag key and export it as an environment variable. Your tag key is the value returned in the `NAME` column of the previous command's output.

    ```shell
    export TAG_KEY="tagKeys/123456789012345"
    ```

7. Export any valid tag value. For example, `None`.

    ```shell
    export TAG_VALUE=None
    ```

8. Create the tag value in the remote project.

    ```shell
    gcloud resource-manager tags values create $TAG_VALUE --tag-key=$TAG_KEY
    ```

9. Fetch the tag with the value created in the previous step.

    ```shell
    gcloud resource-manager tags values list --parent=$TAG_KEY
    ```

10. Replace the `tagValues/1234567890123456789` placeholder with the fetched tag value. Export it as an environment variable.

    ```shell
    export TAG_VALUE="tagValues/1234567890123456789"
    ```

11. Replace the placeholder with your VPC network name and export it as an environment variable.

    ```shell
    export REMOTE_VPC_NETWORK={REMOTE_VPC_NETWORK}
    ```

12. Fetch the network selfLinkWithId from the remote vpc network.

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

13. Export resource ID environment variable. Use the value of `selfLinkWithId` returned in the previous command's output, but replace `https://www.googleapis.com/compute/v1` with `//compute.googleapis.com`.

    ```shell
    export RESOURCE_ID="//compute.googleapis.com/projects/remote-project-id/global/networks/1234567890123456789"
    ```

14. Add the tag to the VPC network.

    ```shell
    gcloud resource-manager tags bindings create --tag-value=$TAG_VALUE --parent=$RESOURCE_ID
    ```

15. Create a GCP VPC Peering manifest file.

    ```shell
    cat <<EOF > vpc-peering.yaml
    apiVersion: cloud-resources.kyma-project.io/v1beta1
    kind: GcpVpcPeering
    metadata:
        name: "vpcpeering-dev"
    spec:
        remotePeeringName: "my-project-to-kyma-dev"
        remoteProject: "remote-project-id"
        remoteVpc: "remote-vpc-network"
        importCustomRoutes: false
    EOF
    ```

16. Apply the Google Cloud VPC peering manifest file.

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
