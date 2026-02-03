package allocate

import (
	"errors"
	"fmt"
	"strings"

	"github.com/kyma-project/cloud-manager/pkg/util"
)

type AddressSpace interface {
	Reserve(arr ...string) error
	Release(r string) error
	ReleaseIpAddress(ip string) error
	Allocate(maskOnes int) (string, error)
	MustAllocate(maskOnes int) string
	AllocateOneIpAddress() (string, error)
	MustAllocateOneIpAddress() string
	fmt.Stringer
}

func NewAddressSpace(cidr string) (AddressSpace, error) {
	rr, err := parseRange(cidr)
	if err != nil {
		return nil, err
	}
	return &addressSpace{
		space:    rr,
		occupied: newRangeList(),
	}, nil
}

type addressSpace struct {
	space    *rng
	occupied *rngList
}

func (as *addressSpace) Reserve(arr ...string) error {
	for _, r := range arr {
		rr, err := parseRange(r)
		if err != nil {
			return err
		}
		if !as.space.contains(rr) {
			return fmt.Errorf("requested range %s is out of address space %s", r, as.space.s)
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
	startAtRange := fmt.Sprintf("%s/32", as.space.n.IP.String())
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
	startAtRange := fmt.Sprintf("%s/%d", as.space.n.IP.String(), maskOnes)
	return as.allocateInternal(startAtRange)
}

func (as *addressSpace) allocateInternal(startAtRange string) (string, error) {
	current, err := parseRange(startAtRange)
	if err != nil {
		return "", fmt.Errorf("invalid startAtRange: %w", err)
	}
	if !as.space.contains(current) {
		return "", fmt.Errorf("startAtRange %s is out of address space %s", startAtRange, as.space.s)
	}

	for as.occupied.overlaps(current) {
		current = current.next()
		if current == nil {
			return "", errors.New("unable to find vacant cidr slot")
		}
		if !as.space.contains(current) {
			return "", errors.New("address space exhausted")
		}
	}

	as.occupied.add(current)

	return current.s, nil
}

func (as *addressSpace) String() string {
	return fmt.Sprintf("%s(%s)", as.space.s, as.occupied.String())
}
