package util

import (
	"errors"
	"fmt"
)

type ErrInvalidNameSyntax struct {
	path string
}

func NewErrInvalidNameSyntax(path string) ErrInvalidNameSyntax {
	return ErrInvalidNameSyntax{path: path}
}

func (e ErrInvalidNameSyntax) Error() string {
	return fmt.Sprintf("gcp name %s has invalid syntax", e.path)
}

func IsInvalidSyntax(err error) bool {
	return errors.Is(err, &ErrInvalidNameSyntax{})
}

type ErrInvalidNameSequence struct {
	path string
}

func NewErrInvalidNameSequence(path string) ErrInvalidNameSequence {
	return ErrInvalidNameSequence{path: path}
}

func (e ErrInvalidNameSequence) Error() string {
	return fmt.Sprintf("gcp name %s has valid syntax but has no valid sequence", e.path)
}

func IsInvalidSequence(err error) bool {
	return errors.Is(err, &ErrInvalidNameSequence{})
}

