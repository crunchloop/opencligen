package gen

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/crunchloop/opencligen/internal/plan"
	"github.com/crunchloop/opencligen/internal/spec"
)

func TestE2E_GeneratedCLI_Help(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	// Load test spec
	ctx := context.Background()
	s, err := spec.Load(ctx, "../testdata/dap.json")
	if err != nil {
		t.Fatalf("failed to load spec: %v", err)
	}

	// Build plan
	p := plan.Build(s, "dap", "github.com/example/dap")

	// Create temp directory
	outDir := t.TempDir()

	// Generate
	gen := New(p, outDir)
	if err := gen.Generate(); err != nil {
		t.Fatalf("generation failed: %v", err)
	}

	// Run go mod tidy
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = outDir
	if output, err := tidyCmd.CombinedOutput(); err != nil {
		t.Fatalf("go mod tidy failed: %v\n%s", err, output)
	}

	// Build
	binaryPath := filepath.Join(outDir, "dap")
	buildCmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/dap")
	buildCmd.Dir = outDir
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, output)
	}

	// Test root help
	t.Run("root help", func(t *testing.T) {
		output, err := exec.Command(binaryPath, "--help").CombinedOutput()
		if err != nil {
			t.Fatalf("help command failed: %v", err)
		}

		helpText := string(output)

		// Check for expected groups
		expectedGroups := []string{"tasks", "workspaces", "stream", "health"}
		for _, group := range expectedGroups {
			if !strings.Contains(helpText, group) {
				t.Errorf("expected help to contain '%s' group", group)
			}
		}

		// Check for global flags
		if !strings.Contains(helpText, "--base-url") {
			t.Error("expected help to contain --base-url flag")
		}
		if !strings.Contains(helpText, "--timeout") {
			t.Error("expected help to contain --timeout flag")
		}
	})

	// Test tasks group help
	t.Run("tasks group help", func(t *testing.T) {
		output, err := exec.Command(binaryPath, "tasks", "--help").CombinedOutput()
		if err != nil {
			t.Fatalf("tasks help command failed: %v", err)
		}

		helpText := string(output)

		// Check for expected commands
		expectedCmds := []string{"list", "create", "get", "cancel"}
		for _, cmd := range expectedCmds {
			if !strings.Contains(helpText, cmd) {
				t.Errorf("expected tasks help to contain '%s' command", cmd)
			}
		}
	})

	// Test tasks create help
	t.Run("tasks create help", func(t *testing.T) {
		output, err := exec.Command(binaryPath, "tasks", "create", "--help").CombinedOutput()
		if err != nil {
			t.Fatalf("tasks create help command failed: %v", err)
		}

		helpText := string(output)

		// Check for --user-id flag (from X-User-Id header)
		if !strings.Contains(helpText, "--user-id") {
			t.Error("expected tasks create help to contain --user-id flag")
		}

		// Check for --data flag
		if !strings.Contains(helpText, "--data") {
			t.Error("expected tasks create help to contain --data flag")
		}
	})

	// Test tasks get help
	t.Run("tasks get help", func(t *testing.T) {
		output, err := exec.Command(binaryPath, "tasks", "get", "--help").CombinedOutput()
		if err != nil {
			t.Fatalf("tasks get help command failed: %v", err)
		}

		helpText := string(output)

		// Check for positional <id>
		if !strings.Contains(helpText, "<id>") {
			t.Error("expected tasks get help to show <id> positional")
		}
	})

	// Test workspaces list help
	t.Run("workspaces list help", func(t *testing.T) {
		output, err := exec.Command(binaryPath, "workspaces", "list", "--help").CombinedOutput()
		if err != nil {
			t.Fatalf("workspaces list help command failed: %v", err)
		}

		helpText := string(output)

		// Check for pagination flags
		if !strings.Contains(helpText, "--page") {
			t.Error("expected workspaces list help to contain --page flag")
		}
		if !strings.Contains(helpText, "--limit") {
			t.Error("expected workspaces list help to contain --limit flag")
		}
	})

	// Test stream subscribe help
	t.Run("stream subscribe help", func(t *testing.T) {
		output, err := exec.Command(binaryPath, "stream", "subscribe", "--help").CombinedOutput()
		if err != nil {
			t.Fatalf("stream subscribe help command failed: %v", err)
		}

		helpText := string(output)

		// Just verify the command exists and has help
		if !strings.Contains(helpText, "Subscribe") {
			t.Error("expected stream subscribe help to contain description")
		}
	})
}

func TestE2E_GeneratedCLI_RequiresBaseURL(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	// Load test spec
	ctx := context.Background()
	s, err := spec.Load(ctx, "../testdata/dap.json")
	if err != nil {
		t.Fatalf("failed to load spec: %v", err)
	}

	// Build plan
	p := plan.Build(s, "dap", "github.com/example/dap")

	// Create temp directory
	outDir := t.TempDir()

	// Generate and build
	gen := New(p, outDir)
	if err := gen.Generate(); err != nil {
		t.Fatalf("generation failed: %v", err)
	}

	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = outDir
	if output, err := tidyCmd.CombinedOutput(); err != nil {
		t.Fatalf("go mod tidy failed: %v\n%s", err, output)
	}

	binaryPath := filepath.Join(outDir, "dap")
	buildCmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/dap")
	buildCmd.Dir = outDir
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, output)
	}

	// Try to run a command without base URL - should fail
	cmd := exec.Command(binaryPath, "tasks", "list")
	cmd.Env = os.Environ() // Inherit environment but not set DAP_BASE_URL
	output, err := cmd.CombinedOutput()

	if err == nil {
		t.Fatal("expected command to fail without base URL")
	}

	if !strings.Contains(string(output), "base URL is required") {
		t.Errorf("expected error about missing base URL, got: %s", output)
	}
}

func TestE2E_AnnotatedCLI(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	// Load annotated spec
	ctx := context.Background()
	s, err := spec.Load(ctx, "../testdata/annotated.json")
	if err != nil {
		t.Fatalf("failed to load spec: %v", err)
	}

	// Build plan
	p := plan.Build(s, "annotated", "github.com/example/annotated")

	// Create temp directory
	outDir := t.TempDir()

	// Generate and build
	gen := New(p, outDir)
	if err := gen.Generate(); err != nil {
		t.Fatalf("generation failed: %v", err)
	}

	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = outDir
	if output, err := tidyCmd.CombinedOutput(); err != nil {
		t.Fatalf("go mod tidy failed: %v\n%s", err, output)
	}

	binaryPath := filepath.Join(outDir, "annotated")
	buildCmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/annotated")
	buildCmd.Dir = outDir
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, output)
	}

	// Test that internal group is hidden
	t.Run("internal is hidden", func(t *testing.T) {
		output, err := exec.Command(binaryPath, "--help").CombinedOutput()
		if err != nil {
			t.Fatalf("help command failed: %v", err)
		}

		helpText := string(output)

		if strings.Contains(helpText, "internal") {
			t.Error("internal group should be hidden")
		}
	})

	// Test that activities has aliases
	t.Run("activities has aliases", func(t *testing.T) {
		output, err := exec.Command(binaryPath, "tasks", "activities", "--help").CombinedOutput()
		if err != nil {
			t.Fatalf("activities help command failed: %v", err)
		}

		helpText := string(output)

		if !strings.Contains(helpText, "Aliases") {
			t.Error("expected to see Aliases in help")
		}
		if !strings.Contains(helpText, "act") {
			t.Error("expected to see 'act' alias")
		}
	})

	// Test that activities has shorthand flags
	t.Run("activities has shorthand flags", func(t *testing.T) {
		output, err := exec.Command(binaryPath, "tasks", "activities", "--help").CombinedOutput()
		if err != nil {
			t.Fatalf("activities help command failed: %v", err)
		}

		helpText := string(output)

		if !strings.Contains(helpText, "-p") {
			t.Error("expected to see -p shorthand for page")
		}
		if !strings.Contains(helpText, "-o") {
			t.Error("expected to see -o shorthand for org")
		}
	})
}
