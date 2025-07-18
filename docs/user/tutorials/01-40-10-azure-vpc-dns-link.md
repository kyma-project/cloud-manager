# Creating VPC DNS Link in Microsoft Azure

This tutorial explains how to link SAP, BTP Kuma runtime network to remote Private DNS zone in Microsoft Azure. Learn how to create a new resource group, VPC network and a virtual machine (VM), and assign required roles to the provided Kyma service principal in your Microsoft Azure subscription.

## Prerequisites

* You have the Cloud Manager module added. See [Add and Delete a Kyma Module](https://help.sap.com/docs/btp/sap-business-technology-platform-internal/enable-and-disable-kyma-module?state=DRAFT&version=Internal#loio1b548e9ad4744b978b8b595288b0cb5c).
* Azure CLI

## Steps

### Authorize Cloud Manager in the Remote Subscription

1. Log in to your Microsoft Azure account and set the active subscription:

   ```shell
   export SUBSCRIPTION={SUBSCRIPTION}
   az login
   az account set --subscription $SUBSCRIPTION
   ```

2. Verify if the Cloud Manager service principal exists in your tenant.
   ```shell
   export APPLICATION_ID={APPLICATION_ID}
   az ad sp show --id $APPLICATION_ID
   ```
3. **Optional:** If the service principal doesn't exist, create one for the Cloud Manager application in your tenant.
   ```shell
   az ad sp create --id $APPLICATION_ID
   ```
4. Assign the required `Classic Network Contributor` and `Network Contributor` Identity and Access Management (IAM) roles to the Cloud Manager service principal. See [Authorizing Cloud Manager in the Remote Cloud Provider](../00-31-vpc-peering-authorization.md#microsoft-azure) to identify the Cloud Manager principal.
    ```shell
    export SUBSCRIPTION_ID=$(az account show --query id -o tsv)
    export OBJECT_ID=$(az ad sp show --id $APPLICATION_ID --query "id" -o tsv)
    
    az role assignment create --assignee $OBJECT_ID \
    --role "Network Contributor" \
    --scope "/subscriptions/$SUBSCRIPTION_ID"
   
    az role assignment create --assignee $OBJECT_ID \
    --role "Classic Network Contributor" \
    --scope "/subscriptions/$SUBSCRIPTION_ID"

### Set Up a Test Environment in the Remote Subscription

1. Set the region that is closest to your Kyma cluster. Use `az account list-locations` to list available locations.

   ```shell
   export REGION={REGION}
   ```

2. Create a resource group as a container for related resources:

   ```shell
   export RESOURCE_GROUP_NAME="MyResourceGroup"
   az group create --name $RESOURCE_GROUP_NAME --location $REGION
   ```

3. Create a Private DNS zone:

   ```shell
   export ZONE_NAME="example.com"
   az network private-dns zone create --resource-group $RESOURCE_GROUP_NAME --name $ZONE_NAME
   ```

4. Create Private DNS A record:

   ```shell
   export RECORD_SET_NAME=test
   export IP_ADDRESS=10.0.0.1
   az network private-dns record-set a add-record --resource-group $RESOURCE_GROUP_NAME --zone-name $ZONE_NAME --record-set-name $RECORD_SET_NAME --ipv4-address $IP_ADDRESS
   ```

### Allow SAP BTP, Kyma Runtime to link with your Private DNS zone

Tag the Private DNS zone with the Kyma shoot name:

   ```shell
   export SHOOT_NAME=$(kubectl get cm -n kube-system shoot-info -o jsonpath='{.data.shootName}') 
   export ZONE_ID=$(az network private-dns show --name $ZONE_NAME --resource-group $RESOURCE_GROUP_NAME --query id --output tsv)
   az tag update --resource-id $ZONE_ID --operation Merge --tags $SHOOT_NAME
   ```

### Create VPC DNS Link

1. Create an AzureVpcDnsLink resource:

   ```shell
   kubectl apply -f - <<EOF
   apiVersion: cloud-resources.kyma-project.io/v1beta1
   kind: AzureVpcDnsLink
   metadata:
     name: kyma-vpc-dns-link
   spec:
     remoteLinkName: kyma-vpc-dns-link
     remotePrivateDnsZone: $ZONE_ID
   EOF
   ```

2. Wait for the AzureVpcDnsLink to be in the `Ready` state.

   ```shell
   kubectl wait --for=condition=Ready azurevpcdnslink/kyma-vpc-dns-link --timeout=300s
   ```

   Once the newly created AzureVpcDnsLink is provisioned, you should see the following message:

   ```console
   azurevpcdnslink.cloud-resources.kyma-project.io/kyma-vpc-dns-link condition met
   ```

3. Create a namespace and export its value as an environment variable:

   ```shell
   export NAMESPACE={NAMESPACE_NAME}
   kubectl create ns $NAMESPACE
   ```

4. Create a workload that queries previously created Private DNS A record:

   ```shell
   kubectl apply -n $NAMESPACE -f - <<EOF
   apiVersion: apps/v1
   kind: Deployment
   metadata:
     name: azurevpcdnslink-demo
   spec:
     selector:
       matchLabels:
         app: azurevpcdnslink-demo
     template:
       metadata:
         labels:
           app: azurevpcdnslink-demo
       spec:
         containers:
         - name: my-container
           resources:
             limits:
               memory: 512Mi
               cpu: "1"
             requests:
               memory: 256Mi
               cpu: "0.2"
           image: ubuntu
           command:
             - "/bin/bash"
             - "-c"
             - "--"
           args:
             - "apt update; apt install dnsutils -y; dig $RECORD_SET_NAME.$ZONE_NAME +noall +answer"
   EOF
   ```

   This workload should print resolved IP address of the Private DNS A record to stdout.

5. To print the logs of one of the workloads, run:

   ```shell
   kubectl logs -n $NAMESPACE `kubectl get pod -n $NAMESPACE -l app=azurevpcdnslink-demo -o=jsonpath='{.items[0].metadata.name}'`
   ```

   The command prints an output similar to the following:

   ```console
   ...
   test.example.com. 30  IN      A       10.0.0.1
   ```

## Next Steps

To clean up Kubernetes resources and your subscription resources, follow these steps:

1. Remove the created workloads:

   ```shell
   kubectl delete -n $NAMESPACE deployment azurevpcdnslink-demo
   ```

2. Remove the created AzureVpcDnsLink resource:

    ```shell
    kubectl delete -n $NAMESPACE azurevpcdnslink kyma-vpc-dns-link
    ```

3. Remove the created namespace:

    ```shell
    kubectl delete namespace $NAMESPACE
    ```

4. In your Microsoft Azure account, remove the created Azure resource group:

    ```shell
    az group delete --name $RESOURCE_GROUP_NAME --yes
    ```
