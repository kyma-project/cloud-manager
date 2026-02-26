package allocate

import (
	"fmt"
	"strings"

	"github.com/elliotchance/pie/v2"
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

func (l *rngList) clone() *rngList {
	if l == nil {
		return nil
	}

	result := &rngList{}
	for _, r := range l.items {
		result.items = append(result.items, r.clone())
	}
	return result
}

func (l *rngList) addStrings(items ...string) error {
	for _, s := range items {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		r, err := parseRange(s)
		if err != nil {
			return fmt.Errorf("invalid cidr %s: %w", s, err)
		}

		l.add(r)
	}
	return nil
}

func (l *rngList) add(items ...*rng) {
	l.items = append(l.items, items...)
}

func (l *rngList) removeString(s string) bool {
	found := false
	l.items = pie.FilterNot(l.items, func(r *rng) bool {
		if r.s == s {
			found = true
			return true
		}
		return false
	})
	return found
}

func (l *rngList) overlaps(o *rng) bool {
	for _, r := range l.items {
		if r.overlaps(o) {
			return true
		}
	}
	return false
}

func (l *rngList) contains(o *rng) bool {
	for _, r := range l.items {
		if r.contains(o) {
			return true
		}
	}
	return false
}

func (l *rngList) String() string {
	var arr []string
	for _, r := range l.items {
		arr = append(arr, r.String())
	}
	return strings.Join(arr, " ")
}
