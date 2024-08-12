# Create Vpc Peering in GCP

This tutorial explains how to create a VPC peering connection between a remote VPC network and Kyma in Google Cloud Platform (GCP).

This tutorial was written using a POSIX-compliant shell, and assumes the cloud-manager module for Kyma is active and enabled.  
If you are not using such a POSIX shell, please adjust the commands accordingly.  
i.e.: If you are using Windows, replace the export commands with set and use % % before and after the environment variables names.

For security reasons it is required that the VPC network in the remote project which will receive the VPC peering connection contains a tag with the Kyma shoot name.

```shell
set KYMA_SHOOT_ID=abc1234
set REMOTE_PROJECT_ID=remote-project-id
gcloud resource-manager tags keys create %KYMA_SHOOT_ID% --parent=projects/%REMOTE_PROJECT_ID%
```

## Steps <!-- {docsify-ignore} -->

1. Fetching your Kyma ID
    
    ```shell
   kubectl get cm -n kube-system shoot-info -o jsonpath='{.data.shootName}'
   # Exporting the result to be used on the next steps
   export KYMA_SHOOT_ID=abc1234
   ```

2. Create a tag key on the remote project with the Kyma shoot name.

   ```shell
   # Please ensure you have exported/set the KYMA_SHOOT_ID in the same shell session (from step 1)
   # Your project ID, referred as remote from the Kyma cluster
   export REMOTE_PROJECT_ID=remote-project-id
   gcloud resource-manager tags keys create $KYMA_SHOOT_ID --parent=projects/$REMOTE_PROJECT_ID
   ```
3. Fetch the tag created in the previous step.

   ```shell
   # Please ensure you have exported/set the REMOTE_PROJECT_ID in the same shell session (from step 2)
   gcloud resource-manager tags keys list --parent=projects/$REMOTE_PROJECT_ID
   # Export the Name value on the last command output, please ensure it is a proper value and not a placeholder like this one
   export TAG_KEY="tagKeys/123456789012345"
   ```
   This command will provide an output similar to this:
   ```console
   NAME                     SHORT_NAME                DESCRIPTION
   tagKeys/123456789012345  shoot--kyma-dev--abc1234
   ```

4. Create a tag value on the remote project with any valid value, i.e.: None.

   ```shell
   # Please ensure you have exported/set the TAG_KEY in the same shell session (from step 3)
   # Using None as tag value because it is the key we are interested in
   export TAG_VALUE=None
   gcloud resource-manager tags values create $TAG_VALUE --parent=$TAG_KEY
   ```

5. Fetch the tag with the value created in the previous step.

   ```shell
   # Please ensure you have exported/set the TAG_KEY in the same shell session (from step 3)
   gcloud resource-manager tags values list --parent=$TAG_KEY
   # Now we will have the value ID to use in the next step
   export TAG_VALUE="tagValues/1234567890123456789"
   ```

6. Fetch the network selfLinkWithId from the remote vpc network.

   ```shell
    export REMOTE_VPC_NETWORK=remote-vpc-network
    gcloud compute networks describe $REMOTE_VPC_NETWORK
   ```
    This command will provide an output similar to this:
    ```console
   ...
   routingConfig:
   routingMode: REGIONAL
   selfLink: https://www.googleapis.com/compute/v1/projects/remote-project-id/global/networks/remote-vpc
   selfLinkWithId: https://www.googleapis.com/compute/v1/projects/remote-project-id/global/networks/1234567890123456789
   subnetworks:
   - https://www.googleapis.com/compute/v1/projects/remote-project-id/regions/europe-west12/subnetworks/remote-vpc
   ...
   ```

7. Create the resource id environment variable to be used on the next step.

   ```shell
   # We will use the selfLinkWithId from the previous command output but replace the https://www.googleapis.com/compute/v1
   # with //compute.googleapis.com
   export RESOURCE_ID="//compute.googleapis.com/projects/remote-project-id/global/networks/1234567890123456789"
   ``` 

8. Adding the tag to the VPC network.

   ```shell
   # please ensure you have exported/set the REMOTE_VPC_NETWORK in the same shell session (from step 6)
   # please ensure you have exported/set the TAG_VALUE in the same shell session (from step 5)
   gcloud resource-manager tags bindings create --tag-value=$TAG_VALUE --parent=$RESOURCE_ID
   ```

9. Create a VPC Peering manifest file.

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

10. Apply the VPC Peering manifest file.

   ```shell
   kubectl apply -f vpc-peering.yaml
   ```

11. This operation usually takes less than 2 minutes. You can check the status of the VPC Peering by running:

    ```shell
    kubectl get gcpvpcpeering vpcpeering-dev -o yaml
    ```

    The output of this command should be similar to this:

    ```console
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
          message: VpcPeering :my-project-to-kyma-dev is provisioned
          reason: Ready
          status: "True"
          type: Ready
      ```
The field conditions under status, will contain relevant information about the VPC Peering status. 