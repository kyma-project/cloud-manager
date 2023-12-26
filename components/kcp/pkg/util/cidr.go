package util

import (
	"bytes"
	"net"
)

func CidrEquals(n1, n2 *net.IPNet) bool {
	return n1.IP.Equal(n2.IP) &&
		bytes.Equal(n1.Mask, n2.Mask)
}

func CidrOverlap(n1, n2 *net.IPNet) bool {
	return n2.Contains(n1.IP) || n1.Contains(n2.IP)
}
