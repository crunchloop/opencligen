package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// executeCommand runs the command with the given args and returns output
func executeCommand(root *cobra.Command, args ...string) (output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	err = root.Execute()
	return buf.String(), err
}

// createTestCommand creates a fresh instance of the CLI for testing
func createTestCommand() *cobra.Command {
	var (
		testSpecPath   string
		testOutDir     string
		testAppName    string
		testModuleName string
		testDryRun     bool
	)

	rootCmd := &cobra.Command{
		Use:   "opencligen",
		Short: "Generate CLI tools from OpenAPI specifications",
	}

	genCmd := &cobra.Command{
		Use:   "gen",
		Short: "Generate a CLI from an OpenAPI spec",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Reset global vars for testing
			specPath = testSpecPath
			outDir = testOutDir
			appName = testAppName
			moduleName = testModuleName
			dryRun = testDryRun
			doBuild = false

			return runGen(cmd, args)
		},
	}

	genCmd.Flags().StringVar(&testSpecPath, "spec", "", "Path to OpenAPI spec file (required)")
	genCmd.Flags().StringVar(&testOutDir, "out", "", "Output directory (required)")
	genCmd.Flags().StringVar(&testAppName, "name", "", "Application name (required)")
	genCmd.Flags().StringVar(&testModuleName, "module", "", "Go module name (optional)")
	genCmd.Flags().BoolVar(&testDryRun, "dry-run", false, "Print plan without generating files")

	_ = genCmd.MarkFlagRequired("spec")
	_ = genCmd.MarkFlagRequired("out")
	_ = genCmd.MarkFlagRequired("name")

	rootCmd.AddCommand(genCmd)
	return rootCmd
}

func TestGen_MissingRequiredFlags(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantErrMsg string
	}{
		{
			name:       "missing all flags",
			args:       []string{"gen"},
			wantErrMsg: "required flag",
		},
		{
			name:       "missing spec flag",
			args:       []string{"gen", "--out", "/tmp/out", "--name", "myapp"},
			wantErrMsg: `required flag(s) "spec" not set`,
		},
		{
			name:       "missing out flag",
			args:       []string{"gen", "--spec", "test.json", "--name", "myapp"},
			wantErrMsg: `required flag(s) "out" not set`,
		},
		{
			name:       "missing name flag",
			args:       []string{"gen", "--spec", "test.json", "--out", "/tmp/out"},
			wantErrMsg: `required flag(s) "name" not set`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := createTestCommand()
			_, err := executeCommand(cmd, tt.args...)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErrMsg) {
				t.Errorf("expected error containing %q, got %q", tt.wantErrMsg, err.Error())
			}
		})
	}
}

func TestGen_SpecFileNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	cmd := createTestCommand()

	_, err := executeCommand(cmd,
		"gen",
		"--spec", "/nonexistent/path/spec.json",
		"--out", tmpDir,
		"--name", "myapp",
	)

	if err == nil {
		t.Fatal("expected error for non-existent spec file")
	}

	if !strings.Contains(err.Error(), "spec file not found") {
		t.Errorf("expected 'spec file not found' error, got: %v", err)
	}
}

func TestGen_DryRun(t *testing.T) {
	// Get the path to the test spec file
	testSpecPath := filepath.Join("..", "..", "internal", "testdata", "dap.json")

	// Verify test file exists
	if _, err := os.Stat(testSpecPath); os.IsNotExist(err) {
		t.Skipf("test spec file not found at %s", testSpecPath)
	}

	tmpDir := t.TempDir()
	cmd := createTestCommand()

	_, err := executeCommand(cmd,
		"gen",
		"--spec", testSpecPath,
		"--out", tmpDir,
		"--name", "testcli",
		"--dry-run",
	)

	if err != nil {
		t.Fatalf("dry-run failed: %v", err)
	}

	// Verify no files were created (the key behavior of dry-run)
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("failed to read temp dir: %v", err)
	}

	if len(entries) > 0 {
		t.Error("expected no files to be created during dry-run")
	}
}

func TestGen_FullGeneration(t *testing.T) {
	// Get the path to the test spec file
	testSpecPath := filepath.Join("..", "..", "internal", "testdata", "dap.json")

	// Verify test file exists
	if _, err := os.Stat(testSpecPath); os.IsNotExist(err) {
		t.Skipf("test spec file not found at %s", testSpecPath)
	}

	tmpDir := t.TempDir()
	cmd := createTestCommand()

	_, err := executeCommand(cmd,
		"gen",
		"--spec", testSpecPath,
		"--out", tmpDir,
		"--name", "testcli",
		"--module", "github.com/test/testcli",
	)

	if err != nil {
		t.Fatalf("generation failed: %v", err)
	}

	// Verify expected files were created
	expectedFiles := []string{
		"go.mod",
		"cmd/testcli/main.go",
		"internal/commands/root.go",
		"internal/runtime/runtime.go",
	}

	for _, expectedFile := range expectedFiles {
		path := filepath.Join(tmpDir, expectedFile)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s to be created", expectedFile)
		}
	}
}

func TestGen_InvalidSpec(t *testing.T) {
	// Create a temp file with invalid JSON
	tmpDir := t.TempDir()
	invalidSpecPath := filepath.Join(tmpDir, "invalid.json")
	if err := os.WriteFile(invalidSpecPath, []byte("not valid json"), 0644); err != nil {
		t.Fatalf("failed to create invalid spec file: %v", err)
	}

	outDir := filepath.Join(tmpDir, "out")
	cmd := createTestCommand()

	_, err := executeCommand(cmd,
		"gen",
		"--spec", invalidSpecPath,
		"--out", outDir,
		"--name", "myapp",
	)

	if err == nil {
		t.Fatal("expected error for invalid spec file")
	}

	if !strings.Contains(err.Error(), "failed to load spec") {
		t.Errorf("expected 'failed to load spec' error, got: %v", err)
	}
}

func TestGen_DefaultModuleName(t *testing.T) {
	// When module name is not provided, it should default to app name
	testSpecPath := filepath.Join("..", "..", "internal", "testdata", "dap.json")

	if _, err := os.Stat(testSpecPath); os.IsNotExist(err) {
		t.Skipf("test spec file not found at %s", testSpecPath)
	}

	tmpDir := t.TempDir()
	cmd := createTestCommand()

	_, err := executeCommand(cmd,
		"gen",
		"--spec", testSpecPath,
		"--out", tmpDir,
		"--name", "myapp",
		// no --module flag
	)

	if err != nil {
		t.Fatalf("generation failed: %v", err)
	}

	// Read go.mod and verify module name is "myapp"
	goModPath := filepath.Join(tmpDir, "go.mod")
	content, err := os.ReadFile(goModPath)
	if err != nil {
		t.Fatalf("failed to read go.mod: %v", err)
	}

	if !strings.Contains(string(content), "module myapp") {
		t.Errorf("expected go.mod to contain 'module myapp', got: %s", string(content))
	}
}
