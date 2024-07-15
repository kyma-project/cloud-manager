# Cloud Resources Technical Documentation

Cloud Resources is a Kyma module managing cloud resources from the cloud provider in the SKR VPC network. 

## API and Reconcilers

Cloud resources defines API in two groups:
* `cloud-resources.kyma-project.io` - user facing API available in the **SKR**
* `cloud-control.kyma-project.io` - low level API in the **KCP** projected from the user facing API in SKR

It has one central active component running in the KCP that runs two sets of reconciliation loops - one for each API group.

Cloud Resources reconcilers are doing remote reconciliation of the `cloud-resources.kyma-project.io` API Group 
from the SKR into the low level `cloud-control.kyma-project.io` API group in KCP. In short - the **SKR Loop**.

Cloud Control reconcilers are doing local reconciliation of `cloud-control.kyma-project.io` API group into the 
resources of the cloud provider - hyperscaler. In short - the **KCP Loop**.

Both set of reconcilers have to also maintain the status of the resources they reconcile. This mean that Cloud Control 
reconcilers will observe the status of the cloud resources and project it into the status of the Cloud Control 
resources in KCP. And that Cloud Resource reconcilers will observe the status of the Cloud Control resources and 
project it into the status of the Cloud Resource resources in SKR.


![API and Reconcilers](./assets/api-and-reconcilers.drawio.png "API and Reconcilers")]

## KCP Cloud Control Controller Manager

The Cloud Control reconcilers are managed by standard 
[controller-runtime controller manager](https://github.com/kubernetes/sample-controller/blob/master/docs/controller-client-go.md) 
maintaining an active long living connection to the KCP Kubernetes API.

[![client-go under the hood](https://raw.githubusercontent.com/kubernetes/sample-controller/master/docs/images/client-go-controller-interaction.jpeg)](https://github.com/kubernetes/sample-controller/blob/master/docs/controller-client-go.md)

# SKR Cloud Resources Controller Manager

Due to the non-scalable concurrent reconciliation of large number of clusters the SKR Cloud Resources Controller Manager 
can not keep long living connections on the remote clusters permanently watching for changes. Instead, a custom
SKR Looper component will loop trough SKRs with enabled Cloud Resources module, and instantiate new 
ControllerRuntime manager that will list all the "watched" (reconciler registered with manager with `.For()` or `.Watches()` 
methods as defined in controller-runtime) and with them maintain a shot lived "cache" until all SKR reconcilers are 
called with respective resources they are managing. Once all done, all resources for that SKR, shot living cache, client... 
will be disposed and same process repeated for the next SKR. 

The reconciler facing API like `Reconcile()` and `.SetupWithManager()` functions will remain as close as possible to 
the one defined by controller-runtime and used by Kubebuilder. 

![SKR Controller Manager](./assets/skr-controller-manager.drawio.png)


## CloudControl Scope resource

Different cloud providers APIs require different connection options to define the scope of the operations:
* GCP - project
* Azure - tenant and subscription
* AWS - account

The SKR loop, when projecting into the KCP resources, will set the `kymaName` field as the indication which SKR that 
resources refers two. So, the starting point for the KCP loop reconcilers is the Kyma CR name. Following the path from 
Kyma CR in KCP to the Shoot CR in Garden it is possible to determine the cloud provider scope. 

The Cloud Control KCP resource Scope was introduced to 
* improve performance by not reading so many different resources and reaching out to the Garden
* simply the development landscape by not avoiding the necessity to beside local KCP cluster also have the Garden cluster 
  with all the relevant resources

Upon the appearance of the first resource from certain SKR the cloud provider scope is determined, saved in the 
Scope KCP resource, and original Cloud Control KCP resource updated with the `scopeRef`. In dev, scope determination 
can be avoided by creating a resource already referring to an existing Scope resource. 

## IpRange

Some cloud resources require an allocation of the private IP, like NFS instance. Since network security is managed 
differently in different cloud providers where some require firewall kind of access approval on the subnet level and 
in order not to be in a situation to modify the security configuation created by Gardener, a decision was made to 
allocate new subnets and not to provision such cloud resources into the nodes subnets. 

For those purposes the IpRange resource is introduced where a CIDR will be specified and subnets provisioned from it. 
So, the reconciliation of the IpRange results in the creating of the subnets in the cloud provider.

Some cloud providers define zone attribute on the subnet, resulting with as many nodes subnets as many the zone are. 
And some cloud providers don't have the zone on subnet, but only the region, resulting in only one node's subnet. 
Similarly, as many nodes subnets there are, as many cloud resources subnets will have to be created. If there's only 
one subnet, then the whole IpRange CIDR will be used for that one subnet. If there are more than one subnet, then 
the IpRange CIDR is split into smaller ranges that are allocated to each subnet. 

Any API resource representing a cloud resource that requires a private IP must have the reference to the IpRange resource. 

