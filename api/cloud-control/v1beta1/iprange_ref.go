package v1beta1

type IpRangeRef struct {
	// +optional
	Name string `json:"name"`
}

const (
	IpRangeField = ".spec.ipRange"
)
