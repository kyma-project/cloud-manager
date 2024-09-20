package v1beta1

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIpRangeSubnets(t *testing.T) {

	t.Run("equal when same elements in same order", func(t *testing.T) {
		a := IpRangeSubnets{
			{
				Id:    "a",
				Zone:  "aa",
				Range: "aaa",
			},
			{
				Id:    "b",
				Zone:  "bb",
				Range: "bbb",
			},
		}
		b := IpRangeSubnets{
			{
				Id:    "a",
				Zone:  "aa",
				Range: "aaa",
			},
			{
				Id:    "b",
				Zone:  "bb",
				Range: "bbb",
			},
		}
		assert.True(t, a.Equals(b))
	})

	t.Run("equal when same elements in different order", func(t *testing.T) {
		a := IpRangeSubnets{
			{
				Id:    "a",
				Zone:  "aa",
				Range: "aaa",
			},
			{
				Id:    "b",
				Zone:  "bb",
				Range: "bbb",
			},
		}
		b := IpRangeSubnets{
			{
				Id:    "b",
				Zone:  "bb",
				Range: "bbb",
			},
			{
				Id:    "a",
				Zone:  "aa",
				Range: "aaa",
			},
		}
		assert.True(t, a.Equals(b))
	})

	t.Run("not equal when same length different values", func(t *testing.T) {
		a := IpRangeSubnets{
			{
				Id:    "a",
				Zone:  "aa",
				Range: "aaa",
			},
			{
				Id:    "b",
				Zone:  "bb",
				Range: "bbb",
			},
		}
		b := IpRangeSubnets{
			{
				Id:    "a2",
				Zone:  "aa",
				Range: "aaa",
			},
			{
				Id:    "b2",
				Zone:  "bb",
				Range: "bbb",
			},
		}
		assert.False(t, a.Equals(b))
	})

	t.Run("not equal when first subset of the other", func(t *testing.T) {
		a := IpRangeSubnets{
			{
				Id:    "a",
				Zone:  "aa",
				Range: "aaa",
			},
		}
		b := IpRangeSubnets{
			{
				Id:    "a",
				Zone:  "aa",
				Range: "aaa",
			},
			{
				Id:    "b",
				Zone:  "bb",
				Range: "bbb",
			},
		}
		assert.False(t, a.Equals(b))
	})

	t.Run("not equal when first superset of the other", func(t *testing.T) {
		a := IpRangeSubnets{
			{
				Id:    "a",
				Zone:  "aa",
				Range: "aaa",
			},
			{
				Id:    "b",
				Zone:  "bb",
				Range: "bbb",
			},
		}
		b := IpRangeSubnets{
			{
				Id:    "b",
				Zone:  "bb",
				Range: "bbb",
			},
		}
		assert.False(t, a.Equals(b))
	})
}
