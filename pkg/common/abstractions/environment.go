package abstractions

import (
	"os"
	"strings"
)

type Environment interface {
	Get(k string) string
	List() map[string]string
}

var _ Environment = &osEnvironment{}

type osEnvironment struct{}

func NewOSEnvironment() Environment {
	return &osEnvironment{}
}

func (e *osEnvironment) Get(k string) string {
	return os.Getenv(k)
}

func (e *osEnvironment) List() map[string]string {
	allEnvVars := os.Environ()
	result := make(map[string]string, len(allEnvVars))
	for _, env := range allEnvVars {
		k, v, _ := strings.Cut(env, "=")
		result[k] = v
	}
	return result
}

func NewMockedEnvironment(values map[string]string) Environment {
	if values == nil {
		values = map[string]string{}
	}
	return &MockedEnvironment{Values: values}
}

var _ Environment = &MockedEnvironment{}

type MockedEnvironment struct {
	Values map[string]string
}

func (e *MockedEnvironment) Get(k string) string {
	return e.Values[k]
}

func (e *MockedEnvironment) List() map[string]string {
	return e.Values
}
