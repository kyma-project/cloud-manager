package allocate

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAllocateCidr(t *testing.T) {

	list := []struct {
		mask     int
		ranges   []string
		expected string
	}{
		{22, []string{"10.250.0.0/22", "10.96.0.0/13", "10.104.0.0/13"}, "10.250.4.0/22"},
		{22, []string{"10.250.0.0/16", "10.96.0.0/13", "10.104.0.0/13"}, "10.251.0.0/22"},
	}
	for x, item := range list {
		t.Run(strconv.Itoa(x), func(t *testing.T) {
			actual, err := AllocateCidr(item.mask, item.ranges)
			if item.expected == "" {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, item.expected, actual)
			}
		})
	}
}
