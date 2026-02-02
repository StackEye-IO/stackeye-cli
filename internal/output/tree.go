// Package output provides CLI output helpers that bridge the CLI's global flags
// with the SDK's output formatters.
package output

import (
	"fmt"
	"io"
	"os"
	"strings"

	sdkoutput "github.com/StackEye-IO/stackeye-go-sdk/output"
)

// TreeNode represents a node in an ASCII tree.
// Task #8027: Supports dependency tree visualization.
type TreeNode struct {
	// Name is the display name for this node
	Name string
	// Status is the status text (e.g., "up", "down", "unreachable")
	Status string
	// Children are the child nodes in the tree
	Children []*TreeNode
}

// TreePrinter renders ASCII tree diagrams with optional color support.
type TreePrinter struct {
	writer   io.Writer
	colorMgr *sdkoutput.ColorManager
	useASCII bool // Use ASCII characters instead of Unicode box-drawing
}

// NewTreePrinter creates a new TreePrinter with the given options.
func NewTreePrinter(colorMode sdkoutput.ColorMode, useASCII bool) *TreePrinter {
	return &TreePrinter{
		writer:   os.Stdout,
		colorMgr: sdkoutput.NewColorManager(colorMode),
		useASCII: useASCII,
	}
}

// SetWriter changes the output destination.
func (p *TreePrinter) SetWriter(w io.Writer) {
	p.writer = w
}

// branchChars returns the characters used for tree branches.
func (p *TreePrinter) branchChars() (pipe, tee, corner, dash string) {
	if p.useASCII {
		return "|", "+", "`", "-"
	}
	return "│", "├", "└", "─"
}

// PrintTree renders the tree starting from the given root nodes.
// Multiple root nodes are rendered as separate trees.
func (p *TreePrinter) PrintTree(roots []*TreeNode) {
	for i, root := range roots {
		p.printNode(root, "", i == len(roots)-1, true)
	}
}

// printNode recursively prints a node and its children.
func (p *TreePrinter) printNode(node *TreeNode, prefix string, isLast bool, isRoot bool) {
	pipe, tee, corner, dash := p.branchChars()

	// Build the branch connector
	var connector string
	if isRoot {
		connector = ""
	} else if isLast {
		connector = corner + dash + dash + " "
	} else {
		connector = tee + dash + dash + " "
	}

	// Format the node line
	nameWithStatus := p.formatNodeName(node)
	_, _ = fmt.Fprintf(p.writer, "%s%s%s\n", prefix, connector, nameWithStatus)

	// Prepare prefix for children
	var childPrefix string
	if isRoot {
		childPrefix = prefix
	} else if isLast {
		childPrefix = prefix + "    "
	} else {
		childPrefix = prefix + pipe + "   "
	}

	// Print children
	for i, child := range node.Children {
		p.printNode(child, childPrefix, i == len(node.Children)-1, false)
	}
}

// formatNodeName formats the node name with status and color.
func (p *TreePrinter) formatNodeName(node *TreeNode) string {
	if node.Status == "" {
		return node.Name
	}

	statusDisplay := p.colorMgr.StatusColor(strings.ToUpper(node.Status))
	return fmt.Sprintf("%s [%s]", node.Name, statusDisplay)
}

// PrintOrphanSection prints a separate section for orphan nodes.
func (p *TreePrinter) PrintOrphanSection(title string, orphans []*TreeNode) {
	if len(orphans) == 0 {
		return
	}

	_, _ = fmt.Fprintf(p.writer, "\n%s:\n", title)
	for _, orphan := range orphans {
		statusDisplay := ""
		if orphan.Status != "" {
			statusDisplay = fmt.Sprintf(" [%s]", p.colorMgr.StatusColor(strings.ToUpper(orphan.Status)))
		}
		_, _ = fmt.Fprintf(p.writer, "  - %s%s\n", orphan.Name, statusDisplay)
	}
}
