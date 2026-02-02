// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	sdkoutput "github.com/StackEye-IO/stackeye-go-sdk/output"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

func TestNewProbeDepsTreeCmd(t *testing.T) {
	cmd := NewProbeDepsTreeCmd()

	if cmd.Use != "tree" {
		t.Errorf("expected Use to be 'tree', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if cmd.Long == "" {
		t.Error("expected Long description to be set")
	}

	// Verify RunE is set
	if cmd.RunE == nil {
		t.Error("expected RunE to be set")
	}

	// Verify --org flag exists
	orgFlag := cmd.Flags().Lookup("org")
	if orgFlag == nil {
		t.Error("expected --org flag to be defined")
	}

	// Verify --probe flag exists
	probeFlag := cmd.Flags().Lookup("probe")
	if probeFlag == nil {
		t.Error("expected --probe flag to be defined")
	}

	// Verify --ascii flag exists
	asciiFlag := cmd.Flags().Lookup("ascii")
	if asciiFlag == nil {
		t.Error("expected --ascii flag to be defined")
	}
}

func TestProbeDepsTreeCmd_InvalidOrgID(t *testing.T) {
	cmd := NewProbeDepsTreeCmd()

	// Create a parent command
	root := &cobra.Command{}
	root.AddCommand(cmd)

	// Set an invalid org ID
	root.SetArgs([]string{"tree", "--org", "not-a-valid-uuid"})

	// Capture output
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := root.Execute()
	if err == nil {
		t.Error("expected error for invalid org ID")
	}

	if !strings.Contains(err.Error(), "invalid organization ID") {
		t.Errorf("expected 'invalid organization ID' error, got: %v", err)
	}
}

func TestProbeDepsTreeCmd_InvalidProbeID(t *testing.T) {
	cmd := NewProbeDepsTreeCmd()

	// Create a parent command
	root := &cobra.Command{}
	root.AddCommand(cmd)

	// Set an invalid probe ID
	root.SetArgs([]string{"tree", "--probe", "not-a-valid-uuid"})

	// Capture output
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := root.Execute()
	if err == nil {
		t.Error("expected error for invalid probe ID")
	}

	if !strings.Contains(err.Error(), "invalid probe ID") {
		t.Errorf("expected 'invalid probe ID' error, got: %v", err)
	}
}

func TestBuildTreeNode(t *testing.T) {
	probe1ID := uuid.New()
	probe2ID := uuid.New()
	probe3ID := uuid.New()

	nodeMap := map[uuid.UUID]*client.DependencyTreeNode{
		probe1ID: {ProbeID: probe1ID, Name: "Router", Status: "up", IsUnreachable: false},
		probe2ID: {ProbeID: probe2ID, Name: "API Server", Status: "up", IsUnreachable: false},
		probe3ID: {ProbeID: probe3ID, Name: "Database", Status: "down", IsUnreachable: true},
	}

	childrenMap := map[uuid.UUID][]uuid.UUID{
		probe1ID: {probe2ID, probe3ID},
	}

	visited := make(map[uuid.UUID]bool)
	treeNode := buildTreeNode(probe1ID, nodeMap, childrenMap, visited)

	if treeNode == nil {
		t.Fatal("expected tree node to be created")
	}

	if treeNode.Name != "Router" {
		t.Errorf("expected name 'Router', got %q", treeNode.Name)
	}

	if treeNode.Status != "up" {
		t.Errorf("expected status 'up', got %q", treeNode.Status)
	}

	if len(treeNode.Children) != 2 {
		t.Errorf("expected 2 children, got %d", len(treeNode.Children))
	}

	// Check that unreachable status is properly set
	var foundUnreachable bool
	for _, child := range treeNode.Children {
		if child.Name == "Database" {
			if child.Status != "unreachable" {
				t.Errorf("expected unreachable status for Database, got %q", child.Status)
			}
			foundUnreachable = true
		}
	}
	if !foundUnreachable {
		t.Error("expected to find Database child node")
	}
}

func TestBuildTreeNode_CycleProtection(t *testing.T) {
	probeID := uuid.New()

	nodeMap := map[uuid.UUID]*client.DependencyTreeNode{
		probeID: {ProbeID: probeID, Name: "Probe", Status: "up"},
	}

	// Create a cycle (probe depends on itself)
	childrenMap := map[uuid.UUID][]uuid.UUID{
		probeID: {probeID},
	}

	visited := make(map[uuid.UUID]bool)
	treeNode := buildTreeNode(probeID, nodeMap, childrenMap, visited)

	if treeNode == nil {
		t.Fatal("expected tree node to be created")
	}

	// The cyclic reference should be ignored
	if len(treeNode.Children) != 0 {
		t.Errorf("expected 0 children (cycle should be broken), got %d", len(treeNode.Children))
	}
}

func TestTreePrinter_Unicode(t *testing.T) {
	printer := output.NewTreePrinter(sdkoutput.ColorNever, false) // Unicode mode

	var buf bytes.Buffer
	printer.SetWriter(&buf)

	root := &output.TreeNode{
		Name:   "Root",
		Status: "up",
		Children: []*output.TreeNode{
			{Name: "Child 1", Status: "up"},
			{Name: "Child 2", Status: "down"},
		},
	}

	printer.PrintTree([]*output.TreeNode{root})

	output := buf.String()

	// Verify output contains the node names
	if !strings.Contains(output, "Root") {
		t.Error("expected output to contain 'Root'")
	}
	if !strings.Contains(output, "Child 1") {
		t.Error("expected output to contain 'Child 1'")
	}
	if !strings.Contains(output, "Child 2") {
		t.Error("expected output to contain 'Child 2'")
	}

	// Verify Unicode box-drawing characters are used
	if !strings.Contains(output, "├") && !strings.Contains(output, "└") {
		t.Error("expected Unicode box-drawing characters in output")
	}
}

func TestTreePrinter_ASCII(t *testing.T) {
	printer := output.NewTreePrinter(sdkoutput.ColorNever, true) // ASCII mode

	var buf bytes.Buffer
	printer.SetWriter(&buf)

	root := &output.TreeNode{
		Name:   "Root",
		Status: "up",
		Children: []*output.TreeNode{
			{Name: "Child 1", Status: "up"},
		},
	}

	printer.PrintTree([]*output.TreeNode{root})

	output := buf.String()

	// Verify ASCII characters are used instead of Unicode
	if strings.Contains(output, "├") || strings.Contains(output, "└") {
		t.Error("expected ASCII characters, but found Unicode")
	}

	// Verify ASCII alternatives are present
	if !strings.Contains(output, "+") && !strings.Contains(output, "`") {
		t.Error("expected ASCII box-drawing characters in output")
	}
}

func TestTreePrinter_OrphanSection(t *testing.T) {
	printer := output.NewTreePrinter(sdkoutput.ColorNever, false)

	var buf bytes.Buffer
	printer.SetWriter(&buf)

	orphans := []*output.TreeNode{
		{Name: "Orphan 1", Status: "up"},
		{Name: "Orphan 2", Status: "down"},
	}

	printer.PrintOrphanSection("Orphan Probes", orphans)

	output := buf.String()

	if !strings.Contains(output, "Orphan Probes:") {
		t.Error("expected output to contain section title")
	}
	if !strings.Contains(output, "Orphan 1") {
		t.Error("expected output to contain 'Orphan 1'")
	}
	if !strings.Contains(output, "Orphan 2") {
		t.Error("expected output to contain 'Orphan 2'")
	}
}
