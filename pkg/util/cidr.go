package util

import (
	"bytes"
	"errors"
	"fmt"
	"net"
)

func CidrEquals(n1, n2 *net.IPNet) bool {
	return n1.IP.Equal(n2.IP) &&
		bytes.Equal(n1.Mask, n2.Mask)
}

func CidrOverlap(n1, n2 *net.IPNet) bool {
	return n2.Contains(n1.IP) || n1.Contains(n2.IP)
}

func CidrParseIPnPrefix(cidr string) (string, int, error) {
	ip, net, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", 0, err
	}

	if !ip.Equal(net.IP) {
		return "", 0, errors.New(fmt.Sprintf("%s is not a valid network.", cidr))
	}

	prefix, _ := net.Mask.Size()
	return ip.String(), prefix, nil
}
