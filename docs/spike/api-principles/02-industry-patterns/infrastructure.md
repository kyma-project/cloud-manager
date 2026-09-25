# Established Infrastructure Provisioning Tools

Leading cloud provider infrastructure tooling:
- Terraform / OpenTofu
- Pulumi
- CrossPlane
None unifies cross provider, but keeps strict provider specific 1:1 scheme. Some provide mapping/templating engines which allows users to define their own simple intent based resources with predefined configuration and to many providers mapping, but none actually implements it nor provides design principles. They all leave it open for the platform providers to use it and shape the high level API and mapping to cloud API for their own needs.

## CloudOrchestrator

Similarly to CloudManager - Unifying API access across different cloud providers is not the focus of CloudOrchestrator. They learned that this is not the main concern of many their stakeholders. Either, you have no config options or you have all config options, first is unusable, second is unmanageable. Experiments where done, without much success (from checking with consumers). They are focused on basic building blocks and leave abstractions on top to be built by their customers.

## Kubernetes
General established pattern with claims, bindings and classes, with provider specific configuration either strictly limited (like volume having fixed set of protocols: configMap, secret, hostPath, local, csi, nfs) or flexible but in simple unstructured form (StorageClass parameters as key value pairs) still under strict predefined protocol (hard-coded or csi specific).

## Gardener
Gardener has mature, developed, and extensible general API for "I need k8s cluster" intent. The shoot resource carries the same delivery across all providers - the k8s cluster system of components. The only place Gardener is fragmented across providers is the InfrastructureConfig having the provider specific resources and not unified nor portable, and that's exactly what CloudManager covers. Current Gardener offering in InfrastructureConfig is opinionated and restrictive, but passes as non-flexible since not a primary user's concern. It's questionable if similar approach with CloudManager can be adopted. Historical experience shows due to onboarding efforts the first incident will result with a request change/new feature breaking the abstraction assumed in initial API design.

The distinction has to be made between Gardener and CloudManager deliveries. Although on first sight it might look similar, and that CloudManager delivers Redis, NFS, WAF across providers same way Gardener does with Kubernetes, there is a significant difference. Gardener deploys same Kubernetes software components thus have the same configuration for them across all providers. CloudManager does not deploy chosen software components but rather calls provider API to provision services that have common protocols or similar functionality, but different configuration. The case would be similar if Gardener would provision provider's Kubernetes offering each having different configuration.
