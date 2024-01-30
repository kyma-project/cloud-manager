package ignorant

import (
	ctrl "sigs.k8s.io/controller-runtime"
)

type Ignorant interface {
	AddName(name string)
	ShouldIgnoreKey(req ctrl.Request) bool
}

func New() Ignorant {
	return &ignorant{
		nameMap: make(map[string]struct{}),
	}
}

var _ Ignorant = &ignorant{}

type ignorant struct {
	nameMap map[string]struct{}
}

func (i *ignorant) AddName(name string) {
	i.nameMap[name] = struct{}{}
}

func (i *ignorant) ShouldIgnoreKey(req ctrl.Request) bool {
	_, ok := i.nameMap[req.Name]
	return ok
}
