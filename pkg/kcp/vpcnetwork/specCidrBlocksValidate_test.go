package vpcnetwork

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSpecCidrBlocksValidate(t *testing.T) {

	testData := []struct {
		title       string
		cidrBlocks  []string
		invalid     []string
		overlapping []string
		normalized  []string
	}{
		{
			title:       "One valid",
			cidrBlocks:  []string{"10.250.0.0/16"},
			invalid:     nil,
			overlapping: nil,
			normalized:  []string{"10.250.0.0/16"},
		},
		{
			title:       "One valid and one invalid",
			cidrBlocks:  []string{"10.250.0.0/16", "something.foo"},
			invalid:     []string{"something.foo"},
			overlapping: nil,
			normalized:  nil,
		},
		{
			title:       "Two valid in reverse order",
			cidrBlocks:  []string{"10.251.0.0/16", "10.250.0.0/16"},
			invalid:     nil,
			overlapping: nil,
			normalized:  []string{"10.250.0.0/16", "10.251.0.0/16"},
		},
		{
			title:       "Two overlapping and one valid",
			cidrBlocks:  []string{"10.250.0.0/15", "10.251.0.0/24", "10.252.0.0/16"},
			invalid:     nil,
			overlapping: []string{"[10.250.0.0/15 10.251.0.0/24]"},
			normalized:  nil,
		},
	}

	for i, data := range testData {
		t.Run(fmt.Sprintf("%d %s", i, data.title), func(t *testing.T) {
			invalid, overlapping, normalized := specCidrBlocksValidateImpl(data.cidrBlocks)
			assert.Equal(t, data.invalid, invalid)
			assert.Equal(t, data.overlapping, overlapping)
			assert.Equal(t, data.normalized, normalized)
		})
	}

}
