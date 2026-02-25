package allocate

import (
	"errors"
	"fmt"
	"strings"

	"github.com/kyma-project/cloud-manager/pkg/util"
)

var PrivateRanges = [...]string{
	"10.0.0.0/8",
	"172.16.0.0/12",
	"192.168.0.0/16",
	"100.64.0.0/10",
}

type AddressSpace interface {
	Reserve(arr ...string) error
	Release(r string) error
	ReleaseIpAddress(ip string) error
	Allocate(maskOnes int) (string, error)
	MustAllocate(maskOnes int) string
	AllocateOneIpAddress() (string, error)
	MustAllocateOneIpAddress() string
	Clone() AddressSpace
	fmt.Stringer
}

func MustNewAddressSpace(cidrs ...string) AddressSpace {
	return util.Must(NewAddressSpace(cidrs...))
}

func NewAddressSpace(cidrs ...string) (AddressSpace, error) {
	validSpace := newRangeList()
	if len(cidrs) == 0 {
		cidrs = []string{"0.0.0.0/0"}
	}
	if err := validSpace.addStrings(cidrs...); err != nil {
		return nil, err
	}
	return &addressSpace{
		validSpace: validSpace,
		occupied:   newRangeList(),
	}, nil
}

var _ AddressSpace = (*addressSpace)(nil)

type addressSpace struct {
	validSpace *rngList
	occupied   *rngList
}

func (as *addressSpace) Clone() AddressSpace {
	return &addressSpace{
		validSpace: as.validSpace.clone(),
		occupied:   as.occupied.clone(),
	}
}

func (as *addressSpace) Reserve(arr ...string) error {
	for _, r := range arr {
		rr, err := parseRange(r)
		if err != nil {
			return fmt.Errorf("error parsing ip range %s: %w", r, err)
		}
		if !as.validSpace.contains(rr) {
			return fmt.Errorf("range %s is out of address space %s", r, as.validSpace.String())
		}
		if as.occupied.overlaps(rr) {
			return fmt.Errorf("range %s overlaps with already occupied ranges %s", r, as.occupied.String())
		}
		as.occupied.add(rr)
	}
	return nil
}

func (as *addressSpace) ReleaseIpAddress(ip string) error {
	return as.Release(fmt.Sprintf("%s/32", ip))
}

func (as *addressSpace) Release(r string) error {
	if !as.occupied.removeString(r) {
		return fmt.Errorf("range %s is not allocated", r)
	}
	return nil
}

func (as *addressSpace) MustAllocateOneIpAddress() string {
	return util.Must(as.AllocateOneIpAddress())
}

func (as *addressSpace) AllocateOneIpAddress() (string, error) {
	startAtRange := fmt.Sprintf("%s/32", as.validSpace.items[0].n.IP.String())
	if startAtRange == "0.0.0.0/32" {
		startAtRange = "10.0.0.0/32"
	}
	result, err := as.allocateInternal(startAtRange)
	if err != nil {
		return result, err
	}
	parts := strings.Split(result, "/")
	return parts[0], nil
}

func (as *addressSpace) MustAllocate(maskOnes int) string {
	return util.Must(as.Allocate(maskOnes))
}

func (as *addressSpace) Allocate(maskOnes int) (string, error) {
	startAtRange := fmt.Sprintf("%s/%d", as.validSpace.items[0].n.IP.String(), maskOnes)
	if startAtRange == "0.0.0.0/32" {
		startAtRange = fmt.Sprintf("10.0.0.0/%d", maskOnes)
	}
	return as.allocateInternal(startAtRange)
}

func (as *addressSpace) allocateInternal(startAtRange string) (string, error) {
	current, err := parseRange(startAtRange)
	if err != nil {
		return "", fmt.Errorf("invalid startAtRange: %w", err)
	}
	if !as.validSpace.contains(current) {
		return "", fmt.Errorf("startAtRange %s is out of address space %s", startAtRange, as.validSpace.String())
	}

	for as.occupied.overlaps(current) {
		current = current.next()
		if current == nil {
			return "", errors.New("unable to find vacant cidr slot")
		}
		if !as.validSpace.contains(current) {
			return "", errors.New("address space exhausted")
		}
	}

	as.occupied.add(current)

	return current.s, nil
}

func (as *addressSpace) String() string {
	return fmt.Sprintf("%s(%s)", as.validSpace.String(), as.occupied.String())
}
