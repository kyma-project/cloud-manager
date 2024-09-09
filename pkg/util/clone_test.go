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
