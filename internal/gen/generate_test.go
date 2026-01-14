package gen

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/crunchloop/opencligen/internal/plan"
	"github.com/crunchloop/opencligen/internal/spec"
)

func TestGenerate(t *testing.T) {
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

	// Verify files exist
	expectedFiles := []string{
		"go.mod",
		"cmd/dap/main.go",
		"internal/runtime/runtime.go",
		"internal/runtime/request.go",
		"internal/runtime/body.go",
		"internal/runtime/output.go",
		"internal/runtime/sse.go",
		"internal/runtime/config.go",
		"internal/commands/root.go",
		"internal/commands/tasks.go",
		"internal/commands/workspaces.go",
		"internal/commands/stream.go",
		"internal/commands/health.go",
	}

	for _, f := range expectedFiles {
		path := filepath.Join(outDir, f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist", f)
		}
	}
}

func TestGenerate_BuildsSuccessfully(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping build test in short mode")
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

	// Try to build
	buildCmd := exec.Command("go", "build", "./cmd/dap")
	buildCmd.Dir = outDir
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, output)
	}

	// Verify binary exists
	binaryPath := filepath.Join(outDir, "dap")
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Error("expected binary to be created")
	}
}
