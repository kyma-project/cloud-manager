# Creating VPC DNS Link in Microsoft Azure

> [!WARNING]
> This is a beta feature available only per request for SAP-internal teams.

This tutorial explains how to link the SAP, BTP Kuma runtime network to a remote private DNS zone in Microsoft Azure. Learn how to create a new resource group, private DNS zone, and record-set, and assign required roles to the provided Kyma service principal in your Microsoft Azure subscription.

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
4. Assign the required `Classic Network Contributor` and `Network Contributor` Identity and Access Management (IAM) roles to the Cloud Manager service principal.
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

3. Create a network and subnets:

   ```shell
   export VNET_NAME="MyVnet"
   export ADDRESS_PREFIX=172.16.0.0/16
   export SUBNET_PREFIX=172.16.0.0/24
   export SUBNET_NAME="MySubnet"
   export VNET_ID=$(az network vnet create -g $RESOURCE_GROUP_NAME -n $VNET_NAME --address-prefix $ADDRESS_PREFIX --subnet-name $SUBNET_NAME --subnet-prefixes $SUBNET_PREFIX --query id --output tsv)
   
   export INBOUND_SUBNET_NAME="DnsInboundSubnet"
   export INBOUND_SUBNET_ADDRESS_PREFIX="172.16.128.0/28"
   export INBOUND_SUBNET_ID=$(az network vnet subnet create -g $RESOURCE_GROUP_NAME --vnet-name $VNET_NAME -n $INBOUND_SUBNET_NAME --address-prefix $INBOUND_SUBNET_ADDRESS_PREFIX --query id --output tsv
 
   
   export OUTBOUND_SUBNET_NAME="DnsOutboundSubnet"
   export OUTBOUND_SUBNET_ADDRESS_PREFIX="172.16.128.16/28"
   export OUTBOUND_SUBNET_ID=$(az network vnet subnet create -g $RESOURCE_GROUP_NAME --vnet-name $VNET_NAME -n $OUTBOUND_SUBNET_NAME --address-prefix $OUTBOUND_SUBNET_ADDRESS_PREFIX --query id --output tsv)
   ```

4. Create a DNS private resolver:

   ```shell
   export DNS_RESOLVER_NAME="MyDnsResolver"
   az dns-resolver create --name $DNS_RESOLVER_NAME --location $REGION --id $VNET_ID --resource-group $RESOURCE_GROUP_NAME
   ```
5. Create an inbound endpoint for a DNS resolver:

   ```shell
   export INBOUND_ENDPOINT_NAME="MyInboundEndpoint"
   az dns-resolver inbound-endpoint create --dns-resolver-name $DNS_RESOLVER_NAME --name $INBOUND_ENDPOINT_NAME --location $REGION --ip-configurations "[{private-ip-address:'',private-ip-allocation-method:'Dynamic',id:$INBOUND_SUBNET_ID}]" --resource-group $RESOURCE_GROUP_NAME
   export INBOUND_ENDPOINT_IP=$(az dns-resolver inbound-endpoint show --dns-resolver-name $DNS_RESOLVER_NAME -n $INBOUND_ENDPOINT_NAME -g $RESOURCE_GROUP_NAME --query "ipConfigurations[0].privateIpAddress" --output tsv)
   ```

6. Create an outbound endpoint for a DNS resolver:

   ```shell
   export OUTBOUND_ENDPOINT_NAME="MyOutboundEndpoint"
   az dns-resolver outbound-endpoint create --name $OUTBOUND_ENDPOINT_NAME --resource-group $RESOURCE_GROUP_NAME --dns-resolver-name $DNS_RESOLVER_NAME --location $REGION --id $OUTBOUND_SUBNET_ID
   export OUTBOUND_ENDPOINT_ID=$(az dns-resolver outbound-endpoint create --name $OUTBOUND_ENDPOINT_NAME --resource-group $RESOURCE_GROUP_NAME --dns-resolver-name $DNS_RESOLVER_NAME --location $REGION --id $OUTBOUND_SUBNET_ID --query "id" --output tsv)
   ```

7. Create a DNS forwarding ruleset:

   ```shell
   export RULESET_NAME="MyRuleset"
   az dns-resolver forwarding-ruleset create --name $RULESET_NAME --location $REGION --outbound-endpoints "[{id:$OUTBOUND_ENDPOINT_ID}]" --resource-group $RESOURCE_GROUP_NAME
   ```
   
8. Create a forwarding rule in a DNS forwarding ruleset:
   ```shell
   export RULE_NAME="MyRule"
   az dns-resolver forwarding-rule create --ruleset-name $RULESET_NAME --name $RULE_NAME --domain-name "test.example.com." --forwarding-rule-state "Enabled" --target-dns-servers "[{ip-address:$INBOUND_ENDPOINT_IP,port:53}]" --resource-group $RESOURCE_GROUP_NAME
   ```
   
9. Create a DNS private zone:

   ```shell
   export ZONE_NAME="example.com"
   az network private-dns zone create --resource-group $RESOURCE_GROUP_NAME --name $ZONE_NAME
   ```

10. Add an A record:

    ```shell
    export RECORD_SET_NAME=test
    export IP_ADDRESS=10.0.0.1
    az network private-dns record-set a add-record --resource-group $RESOURCE_GROUP_NAME --zone-name $ZONE_NAME --record-set-name $RECORD_SET_NAME --ipv4-address $IP_ADDRESS
    ```
11. Link your network with private DNS zone:

    ```shell
    export DNS_ZONE_LINK_NAME="MyLink"
    az network private-dns link vnet create --name $DNS_ZONE_LINK_NAME --resource-group $RESOURCE_GROUP_NAME --virtual-network $VNET_ID --zone-name $ZONE_NAME
    ```

### Allow SAP BTP, Kyma Runtime to link with your DNS private resolver

1. Tag the DNS forwarding ruleset with the Kyma shoot name:

   ```shell
   export SHOOT_NAME=$(kubectl get cm -n kube-system shoot-info -o jsonpath='{.data.shootName}') 
   export RULESET_ID=$(az dns-resolver forwarding-ruleset show --name $RULESET_NAME --resource-group $RESOURCE_GROUP_NAME --query id --output tsv)
   az tag update --resource-id $RULESET_ID --operation Merge --tags $SHOOT_NAME
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
     remoteDnsResolverRuleset: $RULESET_ID
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

4. Create a workload that queries previously created private DNS A record:

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

   This workload should print the resolved IP address of the private DNS A record to stdout.

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
