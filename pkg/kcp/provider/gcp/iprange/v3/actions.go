package v3

import "github.com/kyma-project/cloud-manager/pkg/composed"

// Exported action functions for use in parent package.
// These wrap the internal action functions to provide clean public API.

var (
	// Validation and setup actions
	PreventCidrEdit  = preventCidrEdit
	CopyCidrToStatus = copyCidrToStatus
	ValidateCidr     = validateCidr

	// Load actions
	LoadAddress       = loadAddress
	LoadPsaConnection = loadPsaConnection

	// Operation management
	WaitOperationDone = waitOperationDone

	// Status management
	UpdateStatusId = updateStatusId
	UpdateStatus   = updateStatus

	// Address management
	CreateAddress = createAddress
	DeleteAddress = deleteAddress

	// PSA connection management
	IdentifyPeeringIpRanges     = identifyPeeringIpRanges
	CreateOrUpdatePsaConnection = createOrUpdatePsaConnection
	DeletePsaConnection         = deletePsaConnection

	// Allocation preparation
	PrepareAllocateIpRange = prepareAllocateIpRange
)

// Predicates
var (
	NeedsPsaConnection composed.Predicate = needsPsaConnection
)
