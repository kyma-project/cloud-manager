package composed

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SanitizeConditions returns a shallow copy of conditions with the KCP error condition
// message replaced by a tenant-safe string derived from its Reason.
// All other conditions are forwarded unchanged.
func SanitizeConditions(conditions []metav1.Condition) []metav1.Condition {
	result := make([]metav1.Condition, len(conditions))
	copy(result, conditions)
	for i := range result {
		if result[i].Type == cloudcontrolv1beta1.ConditionTypeError {
			result[i].Message = SanitizeErrorForTenant(result[i].Reason)
			break
		}
	}
	return result
}

// reasonMessages maps known KCP error condition reason codes to tenant-friendly messages.
// Reasons not in this map fall back to the generic message below.
var reasonMessages = map[string]string{
	// generic / shared
	cloudcontrolv1beta1.ReasonError:              "an unexpected error occurred",
	cloudcontrolv1beta1.ReasonUnknown:            "an unknown error occurred",
	cloudcontrolv1beta1.ReasonGcpError:           "a cloud provider error occurred",
	cloudcontrolv1beta1.ReasonCloudProviderError: "a cloud provider error occurred",
	cloudcontrolv1beta1.ReasonProviderError:      "a cloud provider error occurred",
	cloudcontrolv1beta1.ReasonValidationFailed:   "the resource configuration is invalid",
	cloudcontrolv1beta1.ReasonInvalidSpec:        "the resource specification is invalid",
	cloudcontrolv1beta1.ReasonMissingDependency:  "a required dependency is missing",
	cloudcontrolv1beta1.ReasonWaitingDependency:  "waiting for a required dependency",
	cloudcontrolv1beta1.ReasonDeleteWhileUsed:    "the resource cannot be deleted while it is in use",
	cloudcontrolv1beta1.ReasonInvalidDependency:  "a dependency is in an invalid state",
	cloudcontrolv1beta1.ReasonInvalidBinding:     "the resource binding is invalid",
	cloudcontrolv1beta1.ReasonScopeNotFound:      "the scope resource was not found", // constant value is "ScopeNoFound" — typo in the constant; keep in sync if ever corrected
	cloudcontrolv1beta1.ReasonNotFound:           "the requested resource was not found",
	cloudcontrolv1beta1.ReasonQuotaExceeded:      "a cloud provider quota limit has been exceeded",

	// IP range
	cloudcontrolv1beta1.ReasonInvalidCidr:                         "the provided CIDR is invalid",
	cloudcontrolv1beta1.ReasonCidrOverlap:                         "the provided CIDR overlaps with an existing range",
	cloudcontrolv1beta1.ReasonCidrCanNotChange:                    "the CIDR cannot be changed after creation",
	cloudcontrolv1beta1.ReasonCidrCanNotSplit:                     "the CIDR cannot be split into the required subnets",
	cloudcontrolv1beta1.ReasonCidrAssociationFailed:               "failed to associate the CIDR with the cloud network",
	cloudcontrolv1beta1.ReasonCidrAllocationFailed:                "failed to allocate the CIDR in the cloud network",
	cloudcontrolv1beta1.ReasonVpcNotFound:                         "the VPC network was not found",
	cloudcontrolv1beta1.ReasonShootAndVpcMismatch:                 "the Shoot cluster and VPC network do not match",
	cloudcontrolv1beta1.ReasonFailedExtendingVpcAddressSpace:      "failed to extend the VPC address space",
	cloudcontrolv1beta1.ReasonFailedDisassociatingVpcAddressSpace: "failed to disassociate the VPC address space",
	cloudcontrolv1beta1.ReasonInvalidIpRangeReference:             "the IP range reference is invalid",

	// VPC peering
	cloudcontrolv1beta1.ReasonFailedCreatingVpcPeeringConnection:      "failed to create the VPC peering connection",
	cloudcontrolv1beta1.ReasonFailedAcceptingVpcPeeringConnection:     "failed to accept the VPC peering connection",
	cloudcontrolv1beta1.ReasonFailedLoadingRemoteVpcNetwork:           "failed to load the remote VPC network",
	cloudcontrolv1beta1.ReasonFailedLoadingRemoteVpcPeeringConnection: "failed to load the remote VPC peering connection",
	cloudcontrolv1beta1.ReasonFailedCreatingRoutes:                    "failed to create routes for the VPC peering connection",
	cloudcontrolv1beta1.ReasonUnauthorized:                            "not authorized to perform the requested cloud operation",
	cloudcontrolv1beta1.ReasonUnauthenticated:                         "authentication to the cloud provider failed",
	cloudcontrolv1beta1.ReasonConflict:                                "a conflicting resource already exists in the cloud provider",

	// NFS
	cloudcontrolv1beta1.ReasonFailedCreatingFileSystem:        "failed to create the file system",
	cloudcontrolv1beta1.ReasonInvalidMountTargetsAlreadyExist: "mount targets already exist and conflict with the requested configuration",

	// Redis instance
	cloudcontrolv1beta1.ReasonCanNotLoadResourceGroup:     "failed to load the resource group",
	cloudcontrolv1beta1.ReasonCanNotDeleteResourceGroup:   "failed to delete the resource group",
	cloudcontrolv1beta1.ReasonCanNotCreateResourceGroup:   "failed to create the resource group",
	cloudcontrolv1beta1.ReasonFailedCreatingRedisInstance: "failed to create the Redis instance",

	// Redis cluster
	cloudcontrolv1beta1.ReasonFailedCreatingRedisCluster: "failed to create the Redis cluster",

	// Azure VNet link
	cloudcontrolv1beta1.ReasonFailedLoadingPrivateDnzZone:      "failed to load the private DNS zone", // constant name has typo "Dnz" — keep in sync if ever corrected
	cloudcontrolv1beta1.ReasonFailedLoadingVirtualNetworkLink:  "failed to load the virtual network link",
	cloudcontrolv1beta1.ReasonFailedCreatingVirtualNetworkLink: "failed to create the virtual network link",
}

const defaultTenantErrorMessage = "an error occurred while provisioning the resource"

// SanitizeErrorForTenant converts an internal KCP error condition reason into a
// stable, human-readable message safe for tenant-facing SKR conditions.
// The raw SDK error text stays server-side in KCP logs and is never exposed to tenants.
func SanitizeErrorForTenant(reason string) string {
	if msg, ok := reasonMessages[reason]; ok {
		return msg
	}
	return defaultTenantErrorMessage
}
