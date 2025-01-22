package v1beta1

type RedisAuthSecretSpec struct {
	Name        string            `json:"name,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	ExtraData   map[string]string `json:"extraData,omitempty"`
}
