package util

import (
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCidr(t *testing.T) {

	t.Run("IpCompare", func(t *testing.T) {
		testData := []struct {
			a net.IP
			b net.IP
			c int
		}{
			{a: net.ParseIP("10.20.30.40"), b: net.ParseIP("2001:db8:3333:4444:5555:6666:7777:8888"), c: -1},
			{a: net.ParseIP("2001:db8:3333:4444:5555:6666:7777:8888"), b: net.ParseIP("10.20.30.40"), c: 1},
			{a: net.ParseIP("10.20.30.40"), b: net.ParseIP("9.10.11.12"), c: 1},
			{a: net.ParseIP("10.20.30.40"), b: net.ParseIP("11.12.13.14"), c: -1},
			{a: net.ParseIP("10.20.30.40"), b: net.ParseIP("10.20.30.41"), c: -1},
			{a: net.ParseIP("10.20.30.40"), b: net.ParseIP("10.20.30.39"), c: 1},
		}

		for _, data := range testData {
			t.Run(fmt.Sprintf("%s__%s", data.a.String(), data.b.String()), func(t *testing.T) {
				actual := IpCompare(data.a, data.b)
				assert.Equal(t, data.c, actual)
			})
		}
	})

}
