package testinfra

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
	"strings"
	"time"
)

func ReportAfterSuite(report ginkgo.Report) {
	root := &node{name: "root"}
	for _, spec := range report.SpecReports {
		if spec.LeafNodeText != "" && len(spec.ContainerHierarchyTexts) > 0 {
			path := append(spec.ContainerHierarchyTexts, spec.LeafNodeText)
			root.add(path, spec.State, spec.RunTime)
			subState := spec.State
			for _, evt := range spec.SpecEvents {
				if evt.SpecEventType == types.SpecEventByStart {
					root.add(append(path, evt.Message), subState, 0)
				}
			}
		}
	}
	root.print()
}

var spaces = strings.Repeat(" ", 100)

type node struct {
	state    types.SpecState
	name     string
	items    []*node
	duration time.Duration
}

func (n *node) prepare() {
	for _, child := range n.items {
		child.prepare()
		n.duration = n.duration + child.duration
	}
}

func (n *node) print() {
	color.NoColor = false
	n.prepare()
	n.printInternal(0)
}

func (n *node) coloredName() string {
	txt := n.name
	txt = strings.ReplaceAll(txt, "And Given ", color.CyanString("And Given "))
	txt = strings.ReplaceAll(txt, "And When ", color.CyanString("And When "))
	txt = strings.ReplaceAll(txt, "And Then ", color.CyanString("And Then "))
	txt = strings.ReplaceAll(txt, "Given ", color.CyanString("Given "))
	txt = strings.ReplaceAll(txt, "When ", color.CyanString("When "))
	txt = strings.ReplaceAll(txt, "Then ", color.CyanString("Then "))
	txt = strings.ReplaceAll(txt, "And ", color.CyanString("And "))
	txt = strings.ReplaceAll(txt, "By ", color.CyanString("By "))

	if n.duration > 0 {
		d := float64(n.duration) / float64(time.Millisecond)
		ds := fmt.Sprintf("%.2fms", d)
		txt = fmt.Sprintf("%s      %s ", txt, color.BlueString(ds))
	}
	return txt
}

func (n *node) coloredState() string {
	if n.state == types.SpecStateInvalid {
		return ""
	}
	if n.state.Is(types.SpecStatePassed) {
		return color.GreenString("✔")
	}
	if n.state.Is(types.SpecStateAborted) || n.state.Is(types.SpecStateSkipped) {
		return color.YellowString("?")
	}
	return color.RedString("×")
}

func (n *node) printInternal(level int) {
	indent := level * 4
	fmt.Printf("%s%v     %v\n", spaces[:indent], n.coloredName(), n.coloredState())
	for _, child := range n.items {
		child.printInternal(level + 1)
	}
}

func (n *node) add(path []string, state types.SpecState, duration time.Duration) {
	if len(path) == 0 {
		return
	}
	for _, child := range n.items {
		if child.name == path[0] {
			child.add(path[1:], state, duration)
			return
		}
	}
	child := &node{name: path[0], state: state}
	n.items = append(n.items, child)
	restOfThePath := path[1:]
	if len(restOfThePath) == 0 {
		child.duration = duration
		return
	}
	child.add(restOfThePath, state, duration)
}
