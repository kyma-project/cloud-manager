package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type CloneTestStruct struct {
	Name    string
	Details *CloneTestDetailsStruct
}

type CloneTestDetailsStruct struct {
	Address string
	Age     int
}

func TestJsonClone(t *testing.T) {
	x := &CloneTestStruct{
		Name: "John",
		Details: &CloneTestDetailsStruct{
			Address: "Main Street",
			Age:     10,
		},
	}

	y, err := JsonClone(x)
	assert.NoError(t, err)

	assert.Equal(t, x.Name, y.Name)
	assert.NotNil(t, y.Details)
	assert.Equal(t, x.Details.Address, y.Details.Address)
	assert.Equal(t, x.Details.Age, y.Details.Age)
}

type CloneTestStructBigger struct {
	Name    string
	Surname string
	Details *CloneTestDetailsStructBigger
}

type CloneTestDetailsStructBigger struct {
	Address string
	Info    string
	Age     int
	Limit   int
}

func TestJsonCloneInto(t *testing.T) {
	x := &CloneTestStructBigger{
		Name:    "John",
		Surname: "Smith",
		Details: &CloneTestDetailsStructBigger{
			Address: "Main Street",
			Info:    "some info",
			Age:     10,
			Limit:   20,
		},
	}

	y := &CloneTestStruct{}
	err := JsonCloneInto(x, y)
	assert.NoError(t, err)

	assert.Equal(t, x.Name, y.Name)
	assert.NotNil(t, y.Details)
	assert.Equal(t, x.Details.Address, y.Details.Address)
	assert.Equal(t, x.Details.Age, y.Details.Age)
}
