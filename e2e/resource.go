package e2e

import (
	"github.com/cucumber/godog"
	messages "github.com/cucumber/messages/go/v21"
	"github.com/rdumont/assistdog"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type ResourceDeclaration struct {
	Alias      string
	Kind       string
	ApiVersion string
	Name       string
	Namespace  string
}

var (
	ad = assistdog.NewDefault()
)

func NewTable(src [][]string) *godog.Table {
	rows := make([]*messages.PickleTableRow, len(src))
	for i, row := range src {
		cells := make([]*messages.PickleTableCell, len(row))
		for j, value := range row {
			cells[j] = &messages.PickleTableCell{Value: value}
		}

		rows[i] = &messages.PickleTableRow{Cells: cells}
	}

	return &godog.Table{Rows: rows}
}

func ParseResourceDeclarations(tbl *godog.Table) ([]*ResourceDeclaration, error) {
	arr, err := ad.CreateSlice(new(ResourceDeclaration), tbl)
	if err != nil {
		return nil, err
	}
	return arr.([]*ResourceDeclaration), nil
}

type ResourceInfo struct {
	ResourceDeclaration
	Evaluated    bool
	Loaded       bool
	GVK          schema.GroupVersionKind
	IsNamespaced bool
	Source       source.Source
}
