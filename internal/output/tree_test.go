// Package output provides CLI output helpers for StackEye commands.
// Task #7170
package output

import (
	"bytes"
	"strings"
	"testing"

	sdkoutput "github.com/StackEye-IO/stackeye-go-sdk/output"
)

func TestNewTreePrinter(t *testing.T) {
	printer := NewTreePrinter(sdkoutput.ColorNever, false)

	if printer == nil {
		t.Fatal("expected printer to be non-nil")
	}

	if printer.colorMgr == nil {
		t.Error("expected colorMgr to be non-nil")
	}

	if printer.useASCII {
		t.Error("expected useASCII to be false")
	}
}

func TestNewTreePrinter_ASCII(t *testing.T) {
	printer := NewTreePrinter(sdkoutput.ColorNever, true)

	if !printer.useASCII {
		t.Error("expected useASCII to be true")
	}
}

func TestTreePrinter_SetWriter(t *testing.T) {
	printer := NewTreePrinter(sdkoutput.ColorNever, false)
	var buf bytes.Buffer
	printer.SetWriter(&buf)

	// Verify it works by printing a simple tree
	printer.PrintTree([]*TreeNode{{Name: "test"}})

	if buf.Len() == 0 {
		t.Error("expected output after SetWriter")
	}
}

func TestTreePrinter_BranchChars_Unicode(t *testing.T) {
	printer := NewTreePrinter(sdkoutput.ColorNever, false)
	pipe, tee, corner, dash := printer.branchChars()

	if pipe != "│" {
		t.Errorf("expected Unicode pipe '│', got %q", pipe)
	}
	if tee != "├" {
		t.Errorf("expected Unicode tee '├', got %q", tee)
	}
	if corner != "└" {
		t.Errorf("expected Unicode corner '└', got %q", corner)
	}
	if dash != "─" {
		t.Errorf("expected Unicode dash '─', got %q", dash)
	}
}

func TestTreePrinter_BranchChars_ASCII(t *testing.T) {
	printer := NewTreePrinter(sdkoutput.ColorNever, true)
	pipe, tee, corner, dash := printer.branchChars()

	if pipe != "|" {
		t.Errorf("expected ASCII pipe '|', got %q", pipe)
	}
	if tee != "+" {
		t.Errorf("expected ASCII tee '+', got %q", tee)
	}
	if corner != "`" {
		t.Errorf("expected ASCII corner '`', got %q", corner)
	}
	if dash != "-" {
		t.Errorf("expected ASCII dash '-', got %q", dash)
	}
}

func TestTreePrinter_PrintTree_SingleRoot(t *testing.T) {
	printer := NewTreePrinter(sdkoutput.ColorNever, false)
	var buf bytes.Buffer
	printer.SetWriter(&buf)

	roots := []*TreeNode{
		{Name: "root-node"},
	}

	printer.PrintTree(roots)
	output := buf.String()

	if !strings.Contains(output, "root-node") {
		t.Errorf("expected output to contain 'root-node', got %q", output)
	}
}

func TestTreePrinter_PrintTree_WithChildren(t *testing.T) {
	printer := NewTreePrinter(sdkoutput.ColorNever, false)
	var buf bytes.Buffer
	printer.SetWriter(&buf)

	roots := []*TreeNode{
		{
			Name: "parent",
			Children: []*TreeNode{
				{Name: "child-1"},
				{Name: "child-2"},
			},
		},
	}

	printer.PrintTree(roots)
	output := buf.String()

	if !strings.Contains(output, "parent") {
		t.Errorf("expected output to contain 'parent', got %q", output)
	}
	if !strings.Contains(output, "child-1") {
		t.Errorf("expected output to contain 'child-1', got %q", output)
	}
	if !strings.Contains(output, "child-2") {
		t.Errorf("expected output to contain 'child-2', got %q", output)
	}

	// Verify tree structure characters are present (Unicode)
	if !strings.Contains(output, "├") {
		t.Errorf("expected tee character in output for non-last child")
	}
	if !strings.Contains(output, "└") {
		t.Errorf("expected corner character in output for last child")
	}
}

func TestTreePrinter_PrintTree_WithChildren_ASCII(t *testing.T) {
	printer := NewTreePrinter(sdkoutput.ColorNever, true)
	var buf bytes.Buffer
	printer.SetWriter(&buf)

	roots := []*TreeNode{
		{
			Name: "parent",
			Children: []*TreeNode{
				{Name: "child-1"},
				{Name: "child-2"},
			},
		},
	}

	printer.PrintTree(roots)
	output := buf.String()

	// Verify ASCII tree structure characters are present
	if !strings.Contains(output, "+--") {
		t.Errorf("expected ASCII tee '+--' in output, got %q", output)
	}
	if !strings.Contains(output, "`--") {
		t.Errorf("expected ASCII corner '`--' in output, got %q", output)
	}
}

func TestTreePrinter_PrintTree_NestedChildren(t *testing.T) {
	printer := NewTreePrinter(sdkoutput.ColorNever, false)
	var buf bytes.Buffer
	printer.SetWriter(&buf)

	roots := []*TreeNode{
		{
			Name: "root",
			Children: []*TreeNode{
				{
					Name: "level-1",
					Children: []*TreeNode{
						{Name: "level-2a"},
						{Name: "level-2b"},
					},
				},
			},
		},
	}

	printer.PrintTree(roots)
	output := buf.String()

	if !strings.Contains(output, "root") {
		t.Error("expected 'root' in output")
	}
	if !strings.Contains(output, "level-1") {
		t.Error("expected 'level-1' in output")
	}
	if !strings.Contains(output, "level-2a") {
		t.Error("expected 'level-2a' in output")
	}
	if !strings.Contains(output, "level-2b") {
		t.Error("expected 'level-2b' in output")
	}
}

func TestTreePrinter_PrintTree_MultipleRoots(t *testing.T) {
	printer := NewTreePrinter(sdkoutput.ColorNever, false)
	var buf bytes.Buffer
	printer.SetWriter(&buf)

	roots := []*TreeNode{
		{Name: "tree-1"},
		{Name: "tree-2"},
		{Name: "tree-3"},
	}

	printer.PrintTree(roots)
	output := buf.String()

	if !strings.Contains(output, "tree-1") {
		t.Error("expected 'tree-1' in output")
	}
	if !strings.Contains(output, "tree-2") {
		t.Error("expected 'tree-2' in output")
	}
	if !strings.Contains(output, "tree-3") {
		t.Error("expected 'tree-3' in output")
	}
}

func TestTreePrinter_PrintTree_EmptyRoots(t *testing.T) {
	printer := NewTreePrinter(sdkoutput.ColorNever, false)
	var buf bytes.Buffer
	printer.SetWriter(&buf)

	printer.PrintTree([]*TreeNode{})

	if buf.Len() != 0 {
		t.Errorf("expected no output for empty roots, got %q", buf.String())
	}
}

func TestTreePrinter_FormatNodeName_NoStatus(t *testing.T) {
	printer := NewTreePrinter(sdkoutput.ColorNever, false)

	node := &TreeNode{Name: "api-server"}
	result := printer.formatNodeName(node)

	if result != "api-server" {
		t.Errorf("expected 'api-server', got %q", result)
	}
}

func TestTreePrinter_FormatNodeName_WithStatus(t *testing.T) {
	printer := NewTreePrinter(sdkoutput.ColorNever, false)

	node := &TreeNode{Name: "api-server", Status: "up"}
	result := printer.formatNodeName(node)

	if !strings.Contains(result, "api-server") {
		t.Errorf("expected result to contain 'api-server', got %q", result)
	}
	if !strings.Contains(result, "UP") {
		t.Errorf("expected result to contain 'UP' (uppercased status), got %q", result)
	}
	if !strings.Contains(result, "[") || !strings.Contains(result, "]") {
		t.Errorf("expected result to contain brackets around status, got %q", result)
	}
}

func TestTreePrinter_PrintOrphanSection(t *testing.T) {
	printer := NewTreePrinter(sdkoutput.ColorNever, false)
	var buf bytes.Buffer
	printer.SetWriter(&buf)

	orphans := []*TreeNode{
		{Name: "orphan-1", Status: "down"},
		{Name: "orphan-2"},
	}

	printer.PrintOrphanSection("Unlinked Probes", orphans)
	output := buf.String()

	if !strings.Contains(output, "Unlinked Probes:") {
		t.Errorf("expected section title in output, got %q", output)
	}
	if !strings.Contains(output, "orphan-1") {
		t.Error("expected 'orphan-1' in output")
	}
	if !strings.Contains(output, "DOWN") {
		t.Error("expected 'DOWN' status in output")
	}
	if !strings.Contains(output, "orphan-2") {
		t.Error("expected 'orphan-2' in output")
	}
}

func TestTreePrinter_PrintOrphanSection_Empty(t *testing.T) {
	printer := NewTreePrinter(sdkoutput.ColorNever, false)
	var buf bytes.Buffer
	printer.SetWriter(&buf)

	printer.PrintOrphanSection("Unlinked Probes", []*TreeNode{})

	if buf.Len() != 0 {
		t.Errorf("expected no output for empty orphans, got %q", buf.String())
	}
}

func TestTreePrinter_PrintOrphanSection_WithoutStatus(t *testing.T) {
	printer := NewTreePrinter(sdkoutput.ColorNever, false)
	var buf bytes.Buffer
	printer.SetWriter(&buf)

	orphans := []*TreeNode{
		{Name: "orphan-no-status"},
	}

	printer.PrintOrphanSection("Orphans", orphans)
	output := buf.String()

	if !strings.Contains(output, "orphan-no-status") {
		t.Error("expected 'orphan-no-status' in output")
	}
	// Should not have brackets if no status
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "orphan-no-status") && strings.Contains(line, "[") {
			t.Error("expected no brackets for orphan without status")
		}
	}
}

func TestTreePrinter_PrintTree_WithStatus(t *testing.T) {
	printer := NewTreePrinter(sdkoutput.ColorNever, false)
	var buf bytes.Buffer
	printer.SetWriter(&buf)

	roots := []*TreeNode{
		{
			Name:   "web-app",
			Status: "up",
			Children: []*TreeNode{
				{Name: "database", Status: "up"},
				{Name: "cache", Status: "down"},
			},
		},
	}

	printer.PrintTree(roots)
	output := buf.String()

	if !strings.Contains(output, "[UP]") {
		t.Errorf("expected '[UP]' in output, got %q", output)
	}
	if !strings.Contains(output, "[DOWN]") {
		t.Errorf("expected '[DOWN]' in output, got %q", output)
	}
}
