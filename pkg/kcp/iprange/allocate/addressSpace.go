package allocate

import (
	"errors"
	"fmt"
)

type AddressSpace interface {
	Reserve(r string) error
	Release(r string)
	Allocate(maskOnes int) (string, error)
	MustAllocate(maskOnes int) string
	AllocateWithPreference(maskOnes int, startAtRange string) (string, error)
}

func NewAddressSpace() AddressSpace {
	return &addressSpace{
		occupied: newRangeList(),
	}
}

type addressSpace struct {
	occupied *rngList
}

func (as *addressSpace) Reserve(r string) error {
	rr, err := parseRange(r)
	if err != nil {
		return err
	}
	as.occupied.add(rr)
	return nil
}

func (as *addressSpace) Release(r string) {
	as.occupied.removeString(r)
}

func (as *addressSpace) MustAllocate(maskOnes int) string {
	res, err := as.Allocate(maskOnes)
	if err != nil {
		panic(err)
	}
	return res
}

func (as *addressSpace) Allocate(maskOnes int) (string, error) {
	startAtRange := fmt.Sprintf("10.0.0.0/%d", maskOnes)
	return as.AllocateWithPreference(maskOnes, startAtRange)
}

func (as *addressSpace) AllocateWithPreference(maskOnes int, startAtRange string) (string, error) {
	current, err := parseRange(startAtRange)
	if err != nil {
		return "", fmt.Errorf("invalid startAtRange: %w", err)
	}

	for as.occupied.overlaps(current) {
		current = current.next()
		if current == nil {
			return "", errors.New("unable to find vacant cidr slot")
		}
	}

	as.occupied.add(current)

	return current.s, nil
}
