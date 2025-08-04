package e2e

import (
	"github.com/cucumber/godog"
	"github.com/rdumont/assistdog"
)

type ResourceDeclaration struct {
	Alias      string
	Kind       string
	ApiVersion string
	Name       string
}

var (
	ad = assistdog.NewDefault()
)

func ParseResourceDeclarations(tbl *godog.Table) ([]*ResourceDeclaration, error) {
	arr, err := ad.CreateInstance(ResourceDeclaration{}, tbl)
	if err != nil {
		return nil, err
	}
	return arr.([]*ResourceDeclaration), nil
}

type ResourceInfo struct {
	ResourceDeclaration
	Evaluated bool
}
