package allocate

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddressSpace(t *testing.T) {

	t.Run("allocate", func(t *testing.T) {
		testData := []struct {
			space     string
			reserved  []string
			request   []int
			allocated []string
			snapshot  []string
		}{
			{
				"10.250.0.0/16",
				nil,
				[]int{17, 18, 19, 20},
				[]string{"10.250.0.0/17", "10.250.128.0/18", "10.250.192.0/19", "10.250.224.0/20"},
				[]string{"10.250.0.0/16(10.250.0.0/17)", "10.250.0.0/16(10.250.0.0/17 10.250.128.0/18)", "10.250.0.0/16(10.250.0.0/17 10.250.128.0/18 10.250.192.0/19)", "10.250.0.0/16(10.250.0.0/17 10.250.128.0/18 10.250.192.0/19 10.250.224.0/20)"},
			},
			{
				"10.250.0.0/16",
				[]string{"10.250.8.0/22", "10.250.16.0/22"},
				[]int{20, 20, 18, 22, 22, 22, 22},
				[]string{"10.250.32.0/20", "10.250.48.0/20", "10.250.64.0/18", "10.250.0.0/22", "10.250.4.0/22", "10.250.12.0/22", "10.250.20.0/22"},
				[]string{"", "", "", "", "", "", ""},
			},
			{
				"10.250.0.0/28",
				[]string{"10.250.0.8/29"},
				[]int{29, 29},
				[]string{"10.250.0.0/29", "err:address space exhausted"},
				[]string{"", ""},
			},
		}

		for i, data := range testData {
			t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
				if len(data.request) != len(data.allocated) {
					t.Fatalf("test data %d is invalid: request and allocated lengths do not match", i)
				}
				if len(data.request) != len(data.snapshot) {
					t.Fatalf("test data %d is invalid: request and snapshot lengths do not match", i)
				}
				as, err := NewAddressSpace(data.space)
				assert.NoError(t, err)
				err = as.Reserve(data.reserved...)
				assert.NoError(t, err)
				for j, mask := range data.request {
					cidr, err := as.Allocate(mask)
					if strings.HasPrefix(data.allocated[j], "err:") {
						expectedErr := strings.TrimPrefix(data.allocated[j], "err:")
						if assert.Errorf(t, err, "item %d", j) {
							assert.Equalf(t, expectedErr, err.Error(), "item %d", j)
						}
					} else {
						assert.NoErrorf(t, err, "item %d", j)
						assert.Equalf(t, data.allocated[j], cidr, "item %d", j)
					}
					if data.snapshot[j] != "" {
						actual := as.String()
						assert.Equalf(t, data.snapshot[j], actual, "item %d", j)
					}
				}
			})
		}
	})

	t.Run("allocate one ip", func(t *testing.T) {
		as, err := NewAddressSpace("10.250.0.0/16")
		assert.NoError(t, err)
		var ip string
		ip, err = as.AllocateOneIpAddress()
		assert.NoError(t, err)
		assert.Equal(t, "10.250.0.0", ip)
		ip, err = as.AllocateOneIpAddress()
		assert.NoError(t, err)
		assert.Equal(t, "10.250.0.1", ip)
		ip, err = as.AllocateOneIpAddress()
		assert.NoError(t, err)
		assert.Equal(t, "10.250.0.2", ip)
	})

}
