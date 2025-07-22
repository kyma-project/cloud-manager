package testinfra

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
)

func ReportAfterSuite(report ginkgo.Report) {
	root := &node{name: "root", root: true}
	for _, spec := range report.SpecReports {
		if spec.LeafNodeText != "" && len(spec.ContainerHierarchyTexts) > 0 {
			path := append(spec.ContainerHierarchyTexts, spec.LeafNodeText)
			root.add(path, spec.State, spec.RunTime)
			subState := spec.State
			for _, evt := range spec.SpecEvents {
				if evt.SpecEventType == types.SpecEventByEnd {
					root.add(append(path, evt.Message), subState, evt.Duration)
				}
			}
		}
	}
	root.print()
}

var spaces = strings.Repeat(" ", 1000)

type node struct {
	root     bool
	state    types.SpecState
	name     string
	items    []*node
	duration time.Duration
}

func (n *node) prepare() {
	addChildrenDurations := false // nolint:staticcheck
	if n.duration == 0 {
		addChildrenDurations = true
	}
	for _, child := range n.items {
		child.prepare()
		if addChildrenDurations {
			n.duration = n.duration + child.duration
		}
	}
}

func (n *node) print() {
	color.NoColor = false
	fmt.Println()
	n.prepare()
	n.printInternal(0, n.maxNameLen()+24, time.Second) // 14 estimated max indent
}

func (n *node) maxNameLen() int {
	result := len(n.name)
	for _, child := range n.items {
		cl := child.maxNameLen()
		if cl > result {
			result = cl
		}
	}
	return result
}

func (n *node) coloredName() string {
	return coloredText(n.name)
}

func (n *node) coloredRelativeDuration(totalTime time.Duration) (string, int) {
	if totalTime == 0 {
		return "", 0
	}
	pct := 100 * float64(n.duration) / float64(totalTime)
	if pct <= 0.01 {
		pct = 0.01234
	}
	ds := fmt.Sprintf("%.2f%%", pct)
	if len(ds) == 5 {
		ds = fmt.Sprintf(" %s", ds)
	}
	txt := color.BlueString(ds)

	return txt, len(ds)
}

func (n *node) coloredDuration() (string, int) {
	if n.duration == 0 {
		return "", 0
	}
	d := float64(n.duration) / float64(time.Millisecond)
	if d <= 0.01 {
		d = 0.01234
	}
	ds := fmt.Sprintf("%.2fms", d)
	txt := color.BlueString(ds)

	return txt, len(ds)
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

func (n *node) printInternal(level int, paddingRight int, totalDuration time.Duration) {
	nameTxt := n.coloredName()
	if n.root {
		fmt.Println(color.BlueString("Feature flags:"))
		ctx := context.Background()
		fmt.Printf("    nukeBackupsGcp: %v\n", feature.FFNukeBackupsGcp.Value(ctx))

		for _, child := range n.items {
			child.printInternal(0, paddingRight, n.duration)
		}

		fmt.Println(strings.Repeat("-", paddingRight+5))
		durationTxt, durationLen := n.coloredDuration()
		cnt := paddingRight - len(n.name) - durationLen - 2
		fmt.Println(color.RedString(fmt.Sprintf("Total time:%s%s", spaces[:cnt], durationTxt)))

	} else {
		durationTxt, durationLen := n.coloredDuration()
		relDurationTxt, relDurationLen := n.coloredRelativeDuration(totalDuration)

		indent := level * 4
		padding := ""
		cnt := paddingRight - len(n.name) - indent - durationLen - relDurationLen
		if cnt > 0 {
			padding = spaces[:cnt]
		}

		fmt.Printf("%s%v   %s  %v %v %v\n", spaces[:indent], nameTxt, padding, durationTxt, relDurationTxt, n.coloredState())

		for _, child := range n.items {
			child.printInternal(level+1, paddingRight, totalDuration)
		}
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

func coloredText(txt string) string {
	if strings.HasPrefix(txt, "//") || strings.HasPrefix(txt, "--") || strings.HasPrefix(txt, "#") {
		return color.HiBlackString(txt)
	}
	txt = strings.ReplaceAll(txt, "SKIPPED:", color.RedString("SKIPPED:"))
	txt = strings.ReplaceAll(txt, "Feature:", color.MagentaString("Feature:"))
	txt = strings.ReplaceAll(txt, "Scenario:", color.YellowString("Scenario:"))
	txt = strings.ReplaceAll(txt, "And Given ", color.CyanString("And Given "))
	txt = strings.ReplaceAll(txt, "And When ", color.CyanString("And When "))
	txt = strings.ReplaceAll(txt, "And Then ", color.CyanString("And Then "))
	txt = strings.ReplaceAll(txt, "Given ", color.CyanString("Given "))
	txt = strings.ReplaceAll(txt, "When ", color.CyanString("When "))
	txt = strings.ReplaceAll(txt, "Then ", color.CyanString("Then "))
	txt = strings.ReplaceAll(txt, "And ", color.CyanString("And "))
	txt = strings.ReplaceAll(txt, "By ", color.CyanString("By "))

	return txt
}
