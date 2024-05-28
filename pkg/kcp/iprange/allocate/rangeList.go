package allocate

import (
	"fmt"
	"strings"
)

type rngList struct {
	items []*rng
}

func newRangeList(items ...*rng) *rngList {
	var arr []*rng
	arr = append(arr, items...)
	return &rngList{
		items: arr,
	}
}

func (l *rngList) addStrings(items ...string) error {
	commonBits := 0
	for _, s := range items {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		r, err := parseRange(s)
		if err != nil {
			return fmt.Errorf("invalid cidr %s: %w", s, err)
		}

		_, bits := r.n.Mask.Size()
		if commonBits == 0 {
			commonBits = bits
		} else {
			if commonBits != bits {
				return fmt.Errorf(
					"all ranges must have same bits, expected %d but encountered %d for range '%s'",
					commonBits,
					bits,
					s,
				)
			}
		}

		l.add(r)
	}
	return nil
}

func (l *rngList) add(items ...*rng) {
	l.items = append(l.items, items...)
}

func (l *rngList) overlaps(o *rng) bool {
	for _, r := range l.items {
		if r.overlaps(o) {
			return true
		}
	}
	return false
}
