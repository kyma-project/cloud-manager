package v1beta1

const (
	ConditionTypeError = "error"

	ReasonInvalidKymaName     = "InvalidKymaName"
	ReasonInvalidCidr         = "InvalidCidr"
	ReasonCidrCanNotSplit     = "CidrCanNotSplit"
	ReasonVpcNotFound         = "VpcNotFound"
	ReasonShootAndVpcMismatch = "ShootAndVpcMismatch"
)
