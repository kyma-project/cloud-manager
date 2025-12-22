package vpcnetwork

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSpecCidrBlocksValidate(t *testing.T) {

	testData := []struct {
		cidrBlocks  []string
		invalid     []string
		overlapping []string
		normalized  []string
	}{
		{
			cidrBlocks:  []string{"10.250.0.0/16"},
			invalid:     nil,
			overlapping: nil,
			normalized:  []string{"10.250.0.0/16"},
		},
		{
			cidrBlocks:  []string{"10.250.0.0/16", "something.foo"},
			invalid:     []string{"something.foo"},
			overlapping: nil,
			normalized:  nil,
		},
	}

	for i, data := range testData {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			invalid, overlapping, normalized := specCidrBlocksValidateImpl(data.cidrBlocks)
			assert.Equal(t, data.invalid, invalid)
			assert.Equal(t, data.overlapping, overlapping)
			assert.Equal(t, data.normalized, normalized)
		})
	}

}
