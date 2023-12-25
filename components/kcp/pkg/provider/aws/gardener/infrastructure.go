package gardener

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

/**
 * To avoid as reference copied from
 * https://github.com/gardener/gardener-extension-provider-aws/blob/master/pkg/apis/aws/v1alpha1/types_infrastructure.go
 */

type InfrastructureConfig struct {
	metav1.TypeMeta `json:",inline"`
	Networks        Networks `json:"networks"`
}

type Networks struct {
	// VPC indicates whether to use an existing VPC or create a new one.
	VPC VPC `json:"vpc"`
	// Zones belonging to the same region
	Zones []Zone `json:"zones"`
}

type Zone struct {
	// Name is the name for this zone.
	Name string `json:"name"`
	// Internal is the private subnet range to create (used for internal load balancers).
	Internal string `json:"internal"`
	// Public is the public subnet range to create (used for bastion and load balancers).
	Public string `json:"public"`
	// Workers is the workers subnet range to create (used for the VMs).
	Workers string `json:"workers"`
	// ElasticIPAllocationID contains the allocation ID of an Elastic IP that will be attached to the NAT gateway in
	// this zone (e.g., `eipalloc-123456`). If it's not provided then a new Elastic IP will be automatically created
	// and attached.
	// Important: If this field is changed then the already attached Elastic IP will be disassociated from the NAT gateway
	// (and potentially removed if it was created by this extension). Also, the NAT gateway will be deleted. This will
	// disrupt egress traffic for a while.
	// +optional
	ElasticIPAllocationID *string `json:"elasticIPAllocationID,omitempty"`
}

type VPC struct {
	// ID is the VPC id.
	// +optional
	ID *string `json:"id,omitempty"`
	// CIDR is the VPC CIDR.
	// +optional
	CIDR *string `json:"cidr,omitempty"`
	// GatewayEndpoints service names to configure as gateway endpoints in the VPC.
	// +optional
	GatewayEndpoints []string `json:"gatewayEndpoints,omitempty"`
}
