package abstractions

import "os"

type Environment interface {
	Get(k string) string
}

type osEnvironment struct{}

func NewOSEnvironment() Environment {
	return &osEnvironment{}
}

func (e *osEnvironment) Get(k string) string {
	return os.Getenv(k)
}

func NewMockedEnvironment(values map[string]string) Environment {
	if values == nil {
		values = map[string]string{}
	}
	return &MockedEnvironment{Values: values}
}

type MockedEnvironment struct {
	Values map[string]string
}

func (e *MockedEnvironment) Get(k string) string {
	return e.Values[k]
}
