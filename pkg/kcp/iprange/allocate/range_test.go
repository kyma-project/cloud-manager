package allocate

import (
	"fmt"
	"math/big"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRange(t *testing.T) {

	t.Run("parseRange", func(t *testing.T) {
		list := []struct {
			s string
			f string
			l string
		}{
			// https://www.ipaddressguide.com/cidr
			{"10.0.0.0/8", "167772160", "184549375"},
			{"10.250.0.0/22", "184156160", "184157183"},
			// https://www.vultr.com/resources/subnet-calculator-ipv6/
			{
				"2001:0db8:85a3:0000:0000:8a2e:0370:7334/64",
				big.NewInt(0).SetBytes(net.ParseIP("2001:0db8:85a3:0000:0000:0000:0000:0000")).String(),
				big.NewInt(0).SetBytes(net.ParseIP("2001:0db8:85a3:0000:ffff:ffff:ffff:ffff")).String(),
			},
		}
		for _, info := range list {
			t.Run(info.s, func(t *testing.T) {
				r, err := parseRange(info.s)
				assert.NoError(t, err)
				assert.Equal(t, info.f, r.first.String())
				assert.Equal(t, info.l, r.last.String())
			})
		}

	})

	t.Run("overlap", func(t *testing.T) {
		list := []struct {
			a string
			b string
			o bool
		}{
			{"10.250.0.0/22", "10.250.16.0/20", false},
			{"10.250.0.0/22", "10.250.0.0/20", true},
		}
		for _, item := range list {
			t.Run(fmt.Sprintf("%s-%s", item.a, item.b), func(t *testing.T) {
				a, err := parseRange(item.a)
				assert.NoError(t, err)
				b, err := parseRange(item.b)
				assert.NoError(t, err)
				assert.Equal(t, item.o, a.overlaps(b))
				assert.Equal(t, item.o, b.overlaps(a))
			})
		}
	})

	t.Run("contains", func(t *testing.T) {
		list := []struct {
			a string
			b string
			c bool
		}{
			{"10.250.0.0/16", "10.250.10.0/30", true},
			{"10.250.0.0/16", "10.250.0.0/30", true},
			{"10.250.0.0/16", "10.250.255.252/30", true},
			{"10.250.0.0/17", "10.250.255.252/30", false},
		}
		for _, item := range list {
			t.Run(fmt.Sprintf("%s-%s", item.a, item.b), func(t *testing.T) {
				a, err := parseRange(item.a)
				assert.NoError(t, err)
				b, err := parseRange(item.b)
				assert.NoError(t, err)
				assert.Equal(t, item.c, a.contains(b))
			})
		}
	})

	t.Run("len", func(t *testing.T) {
		list := []struct {
			s string
			l int
		}{
			{"10.250.0.0/22", 1024},
			{"10.0.0.0/8", 16777216},
		}
		for _, item := range list {
			t.Run(item.s, func(t *testing.T) {
				r, err := parseRange(item.s)
				assert.NoError(t, err)
				assert.Equal(t, item.l, r.len())
			})
		}
	})

	t.Run("next", func(t *testing.T) {
		list := []struct {
			a string
			b string
		}{
			{"10.250.0.0/22", "10.250.4.0/22"},
			{"10.250.8.0/24", "10.250.9.0/24"},
			{"10.0.0.0/8", "11.0.0.0/8"},
			{"255.0.0.0/8", ""},
			{"2001:0db8:85a3:0000:0000:0000:0000:0000/64", "2001:db8:85a3:1::/64"},
		}
		for _, item := range list {
			t.Run(item.a, func(t *testing.T) {
				r, err := parseRange(item.a)
				assert.NoError(t, err)
				n := r.next()
				if len(item.b) == 0 {
					assert.Nil(t, n)
				} else {
					assert.Equal(t, item.b, n.s)
				}
			})
		}
	})
}

func TestRangeList(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		list := newRangeList()
		err := list.addStrings("10.250.0.0/17", "10.250.240.0/21", "10.250.255.244/30")
		assert.NoError(t, err)
		actual := list.String()
		expected := "10.250.0.0/17 10.250.240.0/21 10.250.255.244/30"
		assert.Equal(t, expected, actual)
	})
}
