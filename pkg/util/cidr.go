package util

import (
	"bytes"
	"encoding/binary"
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

func LastCidrAddress(cidr *net.IPNet) net.IP {
	mask := binary.BigEndian.Uint32(cidr.Mask)
	start := binary.BigEndian.Uint32(cidr.IP)
	finish := (start & mask) | (mask ^ 0xffffffff)
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, finish)
	return ip
}

func CidrParseIPnPrefix(cidr string) (string, int, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", 0, err
	}

	if !ip.Equal(ipnet.IP) {
		return "", 0, fmt.Errorf("%s is not a valid network", cidr)
	}

	prefix, _ := ipnet.Mask.Size()
	return ip.String(), prefix, nil
}
