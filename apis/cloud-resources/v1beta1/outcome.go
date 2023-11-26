package v1beta1

type OutcomeType string

const (
	OutcomeTypeCreated OutcomeType = "Created"
	OutcomeTypeError   OutcomeType = "Error"
	OutcomeTypeDeleted OutcomeType = "Deleted"
)

type Outcome struct {
	Type    OutcomeType       `json:"type,omitempty"`
	Message string            `json:"message,omitempty"`
	Data    map[string]string `json:"data,omitempty"`
}
