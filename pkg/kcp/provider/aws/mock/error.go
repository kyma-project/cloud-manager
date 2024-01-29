package mock

import "fmt"

func newNotATypeError(expectedType string, obj any) error {
	return &notATypeError{
		obj:          obj,
		expectedType: expectedType,
	}
}

type notATypeError struct {
	obj          any
	expectedType string
}

func (e *notATypeError) Error() string {
	return fmt.Sprintf("object %T is not an instance of %s", e.obj, e.expectedType)
}
