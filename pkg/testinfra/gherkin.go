package testinfra

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
	"strings"
)

func ReportAfterSuite(report ginkgo.Report) {
	root := &node{name: "root"}
	for _, spec := range report.SpecReports {
		if spec.LeafNodeText != "" && len(spec.ContainerHierarchyTexts) > 0 {
			root.add(append(spec.ContainerHierarchyTexts, spec.LeafNodeText), spec.State)
		}
	}
	root.print()
}

var spaces = strings.Repeat(" ", 100)

type node struct {
	state types.SpecState
	name  string
	items []*node
}

func (n *node) prepare() {

}

func (n *node) print() {
	color.NoColor = false
	n.prepare()
	n.printInternal(0)
}

func (n *node) coloredName() string {
	txt := n.name
	txt = strings.ReplaceAll(txt, "And Given", color.CyanString("And Given"))
	txt = strings.ReplaceAll(txt, "And When", color.CyanString("And When"))
	txt = strings.ReplaceAll(txt, "And Then", color.CyanString("And Then"))
	txt = strings.ReplaceAll(txt, "Given", color.CyanString("Given"))
	txt = strings.ReplaceAll(txt, "When", color.CyanString("When"))
	txt = strings.ReplaceAll(txt, "Then", color.CyanString("Then"))
	txt = strings.ReplaceAll(txt, "And", color.CyanString("And"))
	return txt
}

func (n *node) coloredState() string {
	if n.state.Is(types.SpecStatePassed) {
		return color.GreenString("\u2713")
	}
	if n.state.Is(types.SpecStateAborted) || n.state.Is(types.SpecStateSkipped) {
		return color.YellowString("?")
	}
	return color.RedString("\u10102")
}

func (n *node) printInternal(level int) {
	indent := level * 4
	fmt.Printf("%s%v     %v\n", spaces[:indent], n.coloredName(), n.coloredState())
	for _, child := range n.items {
		child.printInternal(level + 1)
	}
}

func (n *node) add(path []string, state types.SpecState) {
	if len(path) == 0 {
		return
	}
	for _, child := range n.items {
		if child.name == path[0] {
			child.add(path[1:], state)
			return
		}
	}
	child := &node{name: path[0], state: state}
	n.items = append(n.items, child)
	child.add(path[1:], state)
}
