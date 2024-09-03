package ignorant

import (
	ctrl "sigs.k8s.io/controller-runtime"
	"sync"
)

type Ignorant interface {
	AddName(name string)
	RemoveName(name string)
	ShouldIgnoreKey(req ctrl.Request) bool
}

func New() Ignorant {
	return &ignorant{
		nameMap: make(map[string]struct{}),
	}
}

var _ Ignorant = &ignorant{}

type ignorant struct {
	m       sync.Mutex
	nameMap map[string]struct{}
}

func (i *ignorant) AddName(name string) {
	i.m.Lock()
	defer i.m.Unlock()
	i.nameMap[name] = struct{}{}
}

func (i *ignorant) RemoveName(name string) {
	i.m.Lock()
	defer i.m.Unlock()
	delete(i.nameMap, name)
}

func (i *ignorant) ShouldIgnoreKey(req ctrl.Request) bool {
	i.m.Lock()
	defer i.m.Unlock()
	_, ok := i.nameMap[req.Name]
	return ok
}
