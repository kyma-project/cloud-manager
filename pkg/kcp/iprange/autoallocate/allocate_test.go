package autoallocate

import (
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestAllocateCidr(t *testing.T) {

	list := []struct {
		m int
		r []string
		s string
	}{
		{22, []string{"10.250.0.0/22", "10.96.0.0/13", "10.104.0.0/13"}, "10.250.4.0/22"},
	}
	for x, item := range list {
		t.Run(strconv.Itoa(x), func(t *testing.T) {
			actual, err := AllocateCidr(item.m, item.r)
			if item.s == "" {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, item.s, actual)
			}
		})
	}
}
